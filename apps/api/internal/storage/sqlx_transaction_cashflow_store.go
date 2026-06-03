package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
	"github.com/lib/pq"
)

type SQLXBankTransactionStore struct {
	db        *DB
	tableName string
}

type CashflowTransactionQuery struct {
	Limit       int
	Offset      int
	SortBy      string
	SortOrder   string
	Q           string
	Description string
	Note        string
	Source      string
	Direction   string
	Tags        []string
	Untagged    bool
	HideIgnored bool
	From        *time.Time
	To          *time.Time
}

type CashflowTransactionResult struct {
	Total        int
	Transactions []*cashflow.Transaction
}

func NewSQLXBankTransactionStore(db *DB) *SQLXBankTransactionStore {
	return &SQLXBankTransactionStore{
		db:        db,
		tableName: qualifyTable(db, SchemaCashflow, TableTransactions),
	}
}

func parseRows(rows *sqlx.Rows) ([]*cashflow.Transaction, error) {
	var transactions []*cashflow.Transaction
	for rows.Next() {
		var tx cashflow.Transaction
		if err := rows.StructScan(&tx); err != nil {
			return nil, fmt.Errorf("sqlx_transaction_store: failed to scan transaction record: %w", err)
		}
		transactions = append(transactions, &tx)
	}
	return transactions, nil
}

func (s *SQLXBankTransactionStore) Create(ctx context.Context, tx *cashflow.Transaction) error {
	query := fmt.Sprintf(`
        INSERT INTO %s (
            id, account_id, description, note, source, amount_cents,
            direction, date, checksum, created_at, updated_at, tag,
			row_number, ignored, import_id, account_type
        ) VALUES (
            :id, :account_id, :description, :note, :source, :amount_cents,
            :direction, :date, :checksum, :created_at, :updated_at, :tag,
			:row_number, :ignored, :import_id, :account_type
        )
    `, s.tableName)
	executor := s.db.GetExecutor(ctx)
	namedQuery, args, err := sqlx.Named(query, tx)
	if err != nil {
		return fmt.Errorf("sqlx_transaction_store: failed to bind named params: %w", err)
	}
	namedQuery = sqlx.Rebind(sqlx.BindType(s.db.DriverName()), namedQuery)
	_, err = executor.ExecContext(ctx, namedQuery, args...)
	if err != nil {
		if isCashflowDuplicate(err) {
			return cashflow.ErrDuplicateTransaction
		}
		return fmt.Errorf("sqlx_transaction_store: failed to save transaction: %w", err)
	}
	return nil
}

// CreateTransaction persists one cashflow transaction for the application command boundary.
func (s *SQLXBankTransactionStore) CreateTransaction(ctx context.Context, tx *cashflow.Transaction) error {
	return s.Create(ctx, tx)
}

// CreateTransactions persists a batch of cashflow transactions in a single insert,
// skipping rows whose checksum already exists, and returns the number inserted.
func (s *SQLXBankTransactionStore) CreateTransactions(ctx context.Context, txs []*cashflow.Transaction) (int, error) {
	if len(txs) == 0 {
		return 0, nil
	}
	query := fmt.Sprintf(`
        INSERT INTO %s (
            id, account_id, description, note, source, amount_cents,
            direction, date, checksum, created_at, updated_at, tag,
			row_number, ignored, import_id, account_type
        ) VALUES (
            :id, :account_id, :description, :note, :source, :amount_cents,
            :direction, :date, :checksum, :created_at, :updated_at, :tag,
			:row_number, :ignored, :import_id, :account_type
        )
		ON CONFLICT (checksum) DO NOTHING
    `, s.tableName)
	res, err := s.db.NamedExecContext(ctx, query, txs)
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to bulk insert transactions: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to fetch rows affected: %w", err)
	}
	return int(affected), nil
}

func isCashflowDuplicate(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && strings.Contains(strings.ToLower(pqErr.Constraint), "checksum")
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") && strings.Contains(msg, "checksum")
}

func (s *SQLXBankTransactionStore) FetchUntagged(ctx context.Context, page, pageSize int) ([]*cashflow.Transaction, error) {
	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`SELECT * FROM %s 
		WHERE tag IS NULL OR tag = '' 
		ORDER BY date DESC 
		LIMIT ?
		OFFSET ?
	`, s.tableName)
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)
	rows, err := executor.QueryxContext(ctx, query, pageSize, offset)

	if err == sql.ErrNoRows {
		return []*cashflow.Transaction{}, nil // return empty slice if no transactions found
	}
	if err != nil {
		return nil, fmt.Errorf("sqlx_transaction_store: failed to fetch untagged transactions: %w", err)
	}
	transactions, err := parseRows(rows)
	closeErr := rows.Close()
	if err != nil {
		if closeErr != nil {
			return nil, fmt.Errorf("sqlx_transaction_store: failed to parse transaction rows: %w (close failed: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("sqlx_transaction_store: failed to parse transaction rows: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("sqlx_transaction_store: failed to close transaction rows: %w", closeErr)
	}
	return transactions, nil
}

func (s *SQLXBankTransactionStore) Fetch(ctx context.Context, query CashflowTransactionQuery) (*CashflowTransactionResult, error) {
	limit := query.Limit
	if limit == 0 {
		limit = 100
	}
	offset := query.Offset
	sortBy := normalizeCashflowSortBy(query.SortBy)
	sortOrder := normalizeSortOrder(query.SortOrder)

	whereClause, args := buildCashflowWhereClause(query)
	executor := s.db.GetExecutor(ctx)

	totalQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s%s", s.tableName, whereClause)
	totalQuery = s.db.Rebind(totalQuery)
	total := 0
	if err := sqlx.GetContext(ctx, executor, &total, totalQuery, args...); err != nil {
		return nil, fmt.Errorf("sqlx_transaction_store: failed to count transactions: %w", err)
	}

	dataQuery := fmt.Sprintf(`
		SELECT * FROM %s
		%s
		ORDER BY %s %s
		LIMIT ?
		OFFSET ?
	`, s.tableName, whereClause, sortBy, sortOrder)
	dataArgs := append(append([]any{}, args...), limit, offset)
	dataQuery = s.db.Rebind(dataQuery)

	rows, err := executor.QueryxContext(ctx, dataQuery, dataArgs...)
	if err == sql.ErrNoRows {
		return &CashflowTransactionResult{
			Total:        total,
			Transactions: []*cashflow.Transaction{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("sqlx_transaction_store: failed to fetch transactions: %w", err)
	}

	transactions, err := parseRows(rows)
	closeErr := rows.Close()
	if err != nil {
		if closeErr != nil {
			return nil, fmt.Errorf("sqlx_transaction_store: failed to parse transaction rows: %w (close failed: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("sqlx_transaction_store: failed to parse transaction rows: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("sqlx_transaction_store: failed to close transaction rows: %w", closeErr)
	}

	return &CashflowTransactionResult{
		Total:        total,
		Transactions: transactions,
	}, nil
}

// ListTransactions returns cashflow transactions for the application read model.
func (s *SQLXBankTransactionStore) ListTransactions(ctx context.Context, query cashflow.TransactionListQuery) (*cashflow.TransactionListResult, error) {
	direction := ""
	if query.Direction != nil {
		direction = string(*query.Direction)
	}

	result, err := s.Fetch(ctx, CashflowTransactionQuery{
		Limit:       query.Limit,
		Offset:      query.Offset,
		SortBy:      string(query.Sort.Field),
		SortOrder:   strings.ToLower(query.Sort.Direction.SQL()),
		Q:           query.Q,
		Description: query.Description,
		Note:        query.Note,
		Source:      query.Source,
		Direction:   direction,
		Tags:        query.Tags,
		Untagged:    query.Untagged,
		HideIgnored: query.HideIgnored,
		From:        query.From,
		To:          query.To,
	})
	if err != nil {
		return nil, err
	}

	return &cashflow.TransactionListResult{
		Total:        result.Total,
		Transactions: result.Transactions,
	}, nil
}

func (s *SQLXBankTransactionStore) CountByQuery(ctx context.Context, query CashflowTransactionQuery) (int, error) {
	whereClause, args := buildCashflowWhereClause(query)
	countQuery := fmt.Sprintf("SELECT COUNT(1) FROM %s%s", s.tableName, whereClause)
	countQuery = s.db.Rebind(countQuery)
	executor := s.db.GetExecutor(ctx)

	total := 0
	if err := sqlx.GetContext(ctx, executor, &total, countQuery, args...); err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to count transactions by query: %w", err)
	}
	return total, nil
}

// CountByFilter counts cashflow transactions matching application-level filters.
func (s *SQLXBankTransactionStore) CountByFilter(ctx context.Context, filters cashflow.TransactionFilters) (int, error) {
	return s.CountByQuery(ctx, cashflowTransactionQueryFromFilters(filters))
}

func (s *SQLXBankTransactionStore) UpdateTagByQuery(ctx context.Context, query CashflowTransactionQuery, tag string) (int, error) {
	whereClause, args := buildCashflowWhereClause(query)
	updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET tag = ?, updated_at = ?
		%s
	`, s.tableName, whereClause)
	updateQuery = s.db.Rebind(updateQuery)
	executor := s.db.GetExecutor(ctx)

	arguments := make([]any, 0, len(args)+2)
	arguments = append(arguments, tag, time.Now().UTC())
	arguments = append(arguments, args...)

	res, err := executor.ExecContext(ctx, updateQuery, arguments...)
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to update tag by query: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to fetch rows affected: %w", err)
	}
	return int(affected), nil
}

// UpdateTagByFilter updates tags for cashflow transactions matching application-level filters.
func (s *SQLXBankTransactionStore) UpdateTagByFilter(ctx context.Context, filters cashflow.TransactionFilters, tag string) (int, error) {
	return s.UpdateTagByQuery(ctx, cashflowTransactionQueryFromFilters(filters), tag)
}

func (s *SQLXBankTransactionStore) UpdateTagByIDs(ctx context.Context, ids []uuid.UUID, tag string) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET tag = ?, updated_at = ?
		WHERE id IN (?)
	`, s.tableName)

	// Build args for sqlx.In with a single IN-list expansion.
	args := make([]any, 0, len(ids)+2)
	args = append(args, tag, time.Now().UTC(), ids)
	expandedQuery, expandedArgs, err := sqlx.In(query, args...)
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to expand ids for update: %w", err)
	}
	expandedQuery = s.db.Rebind(expandedQuery)

	executor := s.db.GetExecutor(ctx)
	res, err := executor.ExecContext(ctx, expandedQuery, expandedArgs...)
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to update tag by ids: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to fetch rows affected: %w", err)
	}
	return int(affected), nil
}

func (s *SQLXBankTransactionStore) UpdateIgnoredByQuery(ctx context.Context, query CashflowTransactionQuery, ignored bool) (int, error) {
	whereClause, args := buildCashflowWhereClause(query)
	updateQuery := fmt.Sprintf(`
		UPDATE %s
		SET ignored = ?, updated_at = ?
		%s
	`, s.tableName, whereClause)
	updateQuery = s.db.Rebind(updateQuery)
	executor := s.db.GetExecutor(ctx)

	arguments := make([]any, 0, len(args)+2)
	arguments = append(arguments, ignored, time.Now().UTC())
	arguments = append(arguments, args...)

	res, err := executor.ExecContext(ctx, updateQuery, arguments...)
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to update ignored by query: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to fetch rows affected: %w", err)
	}
	return int(affected), nil
}

// UpdateIgnoredByFilter updates ignored-state for cashflow transactions matching application-level filters.
func (s *SQLXBankTransactionStore) UpdateIgnoredByFilter(ctx context.Context, filters cashflow.TransactionFilters, ignored bool) (int, error) {
	return s.UpdateIgnoredByQuery(ctx, cashflowTransactionQueryFromFilters(filters), ignored)
}

func (s *SQLXBankTransactionStore) UpdateIgnoredByIDs(ctx context.Context, ids []uuid.UUID, ignored bool) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	query := fmt.Sprintf(`
		UPDATE %s
		SET ignored = ?, updated_at = ?
		WHERE id IN (?)
	`, s.tableName)

	args := make([]any, 0, len(ids)+2)
	args = append(args, ignored, time.Now().UTC(), ids)
	expandedQuery, expandedArgs, err := sqlx.In(query, args...)
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to expand ids for ignored update: %w", err)
	}
	expandedQuery = s.db.Rebind(expandedQuery)

	executor := s.db.GetExecutor(ctx)
	res, err := executor.ExecContext(ctx, expandedQuery, expandedArgs...)
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to update ignored by ids: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("sqlx_transaction_store: failed to fetch rows affected: %w", err)
	}
	return int(affected), nil
}

func cashflowTransactionQueryFromFilters(filters cashflow.TransactionFilters) CashflowTransactionQuery {
	query := CashflowTransactionQuery{
		Q:           filters.Query,
		Description: filters.Description,
		Note:        filters.Note,
		Source:      filters.Source,
		Tags:        filters.Tags,
		From:        filters.From,
		To:          filters.To,
	}
	if filters.Direction != nil {
		query.Direction = string(*filters.Direction)
	}
	if filters.Untagged != nil {
		query.Untagged = *filters.Untagged
	}
	if filters.HideIgnored != nil {
		query.HideIgnored = *filters.HideIgnored
	}
	return query
}

func buildCashflowWhereClause(query CashflowTransactionQuery) (string, []any) {
	var (
		conditions []string
		args       []any
	)

	appendContains := func(column string, value string) {
		v := strings.TrimSpace(value)
		if v == "" {
			return
		}
		conditions = append(conditions, fmt.Sprintf("LOWER(%s) LIKE ?", column))
		args = append(args, "%"+strings.ToLower(v)+"%")
	}

	appendContains("description", query.Description)
	appendContains("note", query.Note)
	appendContains("source", query.Source)

	if direction := strings.TrimSpace(strings.ToLower(query.Direction)); direction != "" {
		conditions = append(conditions, "LOWER(direction) = ?")
		args = append(args, direction)
	}

	if len(query.Tags) > 0 {
		tagConditions := make([]string, 0, len(query.Tags))
		for _, tag := range query.Tags {
			t := strings.TrimSpace(tag)
			if t == "" {
				continue
			}
			tagConditions = append(tagConditions, "LOWER(COALESCE(tag, '')) = ?")
			args = append(args, strings.ToLower(t))
		}
		if len(tagConditions) > 0 {
			conditions = append(conditions, "("+strings.Join(tagConditions, " OR ")+")")
		}
	}

	if query.Untagged {
		conditions = append(conditions, "TRIM(COALESCE(tag, '')) = ''")
	}

	if query.HideIgnored {
		conditions = append(conditions, "ignored = ?")
		args = append(args, false)
	}

	if query.From != nil {
		conditions = append(conditions, "date >= ?")
		args = append(args, *query.From)
	}

	if query.To != nil {
		conditions = append(conditions, "date <= ?")
		args = append(args, *query.To)
	}

	if q := strings.TrimSpace(query.Q); q != "" {
		pattern := "%" + strings.ToLower(q) + "%"
		conditions = append(conditions, "(LOWER(description) LIKE ? OR LOWER(note) LIKE ? OR LOWER(COALESCE(tag, '')) LIKE ?)")
		args = append(args, pattern, pattern, pattern)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func normalizeCashflowSortBy(sortBy string) string {
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "description":
		return "description"
	case "note":
		return "note"
	case "tag":
		return "tag"
	case "source":
		return "source"
	case "amount":
		return "amount_cents"
	case "date":
		return "date"
	default:
		return "date"
	}
}

func normalizeSortOrder(sortOrder string) string {
	switch strings.ToLower(strings.TrimSpace(sortOrder)) {
	case "asc":
		return "ASC"
	case "desc":
		return "DESC"
	default:
		return "DESC"
	}
}

func (s *SQLXBankTransactionStore) Tag(ctx context.Context, id uuid.UUID, tag string) error {
	_, err := s.UpdateTagByIDs(ctx, []uuid.UUID{id}, tag)
	if err != nil {
		return fmt.Errorf("sqlx_transaction_store: failed to tag transaction: %w", err)
	}
	return nil
}

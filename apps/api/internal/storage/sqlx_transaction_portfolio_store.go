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
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lib/pq"
)

type SQLXPortfolioTransactionStore struct {
	db             *DB
	tableName      string
	positionsTable string
}

func NewSQLXPortfolioTransactionStore(db *DB) *SQLXPortfolioTransactionStore {
	return &SQLXPortfolioTransactionStore{
		db:             db,
		tableName:      qualifyPortfolioTransactionsTable(db),
		positionsTable: qualifyTable(db, SchemaPortfolio, TablePositions),
	}
}

var _ portfolio.TransactionStore = (*SQLXPortfolioTransactionStore)(nil)

func qualifyPortfolioTransactionsTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "portfolio_transactions"
	}
	return fmt.Sprintf("%s.%s", SchemaPortfolio, TableTransactions)
}

func (s *SQLXPortfolioTransactionStore) Create(ctx context.Context, tx *portfolio.Transaction) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, import_id, origin, source, occurred_at, position_id, isin, symbol, description,
			type, quantity, unit_price, amount_cents, checksum, row_number, created_at, updated_at
		) VALUES (
			:id, :account_id, :import_id, :origin, :source, :occurred_at, :position_id, :isin, :symbol, :description,
			:type, :quantity, :unit_price, :amount_cents, :checksum, :row_number, :created_at, :updated_at
		)
	`, s.tableName)
	executor := s.db.GetExecutor(ctx)
	namedQuery, args, err := sqlx.Named(query, tx)
	if err != nil {
		return fmt.Errorf("sqlx_portfolio_transaction_store: bind named params: %w", err)
	}
	namedQuery = sqlx.Rebind(sqlx.BindType(s.db.DriverName()), namedQuery)
	_, err = executor.ExecContext(ctx, namedQuery, args...)
	if err != nil {
		if isPortfolioDuplicate(err) {
			return portfolio.ErrDuplicateTransaction
		}
		return fmt.Errorf("sqlx_portfolio_transaction_store: create transaction: %w", err)
	}
	return nil
}

func (s *SQLXPortfolioTransactionStore) GetASC(ctx context.Context, accID uuid.UUID) ([]portfolio.Transaction, error) {
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE account_id = ?
		ORDER BY occurred_at ASC, row_number ASC
	`, s.tableName)
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)
	var txs []portfolio.Transaction
	if err := sqlx.SelectContext(ctx, executor, &txs, query, accID); err != nil {
		return nil, fmt.Errorf("sqlx_portfolio_transaction_store: get asc: %w", err)
	}
	return txs, nil
}

func (s *SQLXPortfolioTransactionStore) UpdatePositions(ctx context.Context, transactions []portfolio.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
		UPDATE %s
		SET position_id = ?, updated_at = ?
		WHERE id = ?
	`, s.tableName)
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)
	now := time.Now().UTC()
	for i := range transactions {
		tx := transactions[i]
		if tx.PositionID == nil {
			continue
		}
		if _, err := executor.ExecContext(ctx, query, *tx.PositionID, now, tx.ID); err != nil {
			return fmt.Errorf("sqlx_portfolio_transaction_store: update position mapping: %w", err)
		}
	}
	return nil
}

func (s *SQLXPortfolioTransactionStore) GetByPositionID(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]portfolio.Transaction, error) {
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE position_id = ?
	`, s.tableName)
	args := []any{positionID}
	if from != nil {
		query += " AND occurred_at >= ?"
		args = append(args, *from)
	}
	query += " ORDER BY occurred_at ASC, row_number ASC"
	query = s.db.Rebind(query)
	executor := s.db.GetExecutor(ctx)
	var txs []portfolio.Transaction
	if err := sqlx.SelectContext(ctx, executor, &txs, query, args...); err != nil {
		return nil, fmt.Errorf("sqlx_portfolio_transaction_store: get by position: %w", err)
	}
	return txs, nil
}

func (s *SQLXPortfolioTransactionStore) FetchForAccount(
	ctx context.Context,
	query portfolio.TransactionListQuery,
) (*portfolio.TransactionListResult, error) {
	whereClause, whereArgs := buildPortfolioTransactionWhereClause(query)
	orderClause := buildPortfolioTransactionOrderClause(query.SortBy, query.SortOrder)

	limit := query.Limit
	if limit <= 0 {
		limit = 25
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}

	executor := s.db.GetExecutor(ctx)

	countQuery := fmt.Sprintf(`
		SELECT COUNT(1)
		FROM %s t
		%s
	`, s.tableName, whereClause)
	countQuery = s.db.Rebind(strings.TrimSpace(countQuery))
	var total int
	if err := sqlx.GetContext(ctx, executor, &total, countQuery, whereArgs...); err != nil {
		return nil, fmt.Errorf("sqlx_portfolio_transaction_store: count transactions: %w", err)
	}

	dataQuery := fmt.Sprintf(`
		SELECT
			t.*,
			p.listing_id AS listing_id
		FROM %s t
		LEFT JOIN %s p ON p.id = t.position_id
		%s
		%s
		LIMIT ?
		OFFSET ?
	`, s.tableName, s.positionsTable, whereClause, orderClause)
	dataArgs := append(append([]any{}, whereArgs...), limit, offset)
	dataQuery = s.db.Rebind(strings.TrimSpace(dataQuery))
	rows, err := executor.QueryxContext(ctx, dataQuery, dataArgs...)
	if err == sql.ErrNoRows {
		return &portfolio.TransactionListResult{
			Total:        total,
			Transactions: []portfolio.TransactionWithListingID{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("sqlx_portfolio_transaction_store: fetch transactions: %w", err)
	}
	defer rows.Close()

	txs := make([]portfolio.TransactionWithListingID, 0)
	for rows.Next() {
		var tx portfolio.TransactionWithListingID
		if err := rows.StructScan(&tx); err != nil {
			return nil, fmt.Errorf("sqlx_portfolio_transaction_store: scan transactions: %w", err)
		}
		txs = append(txs, tx)
	}
	return &portfolio.TransactionListResult{
		Total:        total,
		Transactions: txs,
	}, nil
}

func buildPortfolioTransactionWhereClause(query portfolio.TransactionListQuery) (string, []any) {
	conditions := []string{"t.account_id = ?"}
	args := []any{query.AccountID}

	appendContains := func(column string, value string) {
		v := strings.ToLower(strings.TrimSpace(value))
		if v == "" {
			return
		}
		conditions = append(conditions, fmt.Sprintf("LOWER(COALESCE(%s, '')) LIKE ?", column))
		args = append(args, "%"+v+"%")
	}

	if query.From != nil {
		conditions = append(conditions, "t.occurred_at >= ?")
		args = append(args, *query.From)
	}
	if query.To != nil {
		conditions = append(conditions, "t.occurred_at <= ?")
		args = append(args, *query.To)
	}
	if query.Type != nil {
		conditions = append(conditions, "t.type = ?")
		args = append(args, string(*query.Type))
	}
	if query.Origin != nil {
		conditions = append(conditions, "t.origin = ?")
		args = append(args, string(*query.Origin))
	}
	appendContains("t.source", query.Source)

	listing := strings.ToLower(strings.TrimSpace(query.Listing))
	if listing != "" {
		conditions = append(conditions, "(LOWER(COALESCE(t.symbol, '')) LIKE ? OR LOWER(COALESCE(t.isin, '')) LIKE ?)")
		args = append(args, "%"+listing+"%", "%"+listing+"%")
	}

	q := strings.ToLower(strings.TrimSpace(query.Q))
	if q != "" {
		pattern := "%" + q + "%"
		conditions = append(conditions, "(LOWER(COALESCE(t.description, '')) LIKE ? OR LOWER(COALESCE(t.source, '')) LIKE ? OR LOWER(COALESCE(t.symbol, '')) LIKE ? OR LOWER(COALESCE(t.isin, '')) LIKE ?)")
		args = append(args, pattern, pattern, pattern, pattern)
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func buildPortfolioTransactionOrderClause(sortBy portfolio.TransactionSortBy, sortOrder portfolio.TransactionSortOrder) string {
	order := "DESC"
	if portfolio.NormalizeTransactionSortOrder(string(sortOrder)) == portfolio.TransactionSortOrderAsc {
		order = "ASC"
	}
	switch portfolio.NormalizeTransactionSortBy(string(sortBy)) {
	case portfolio.TransactionSortByDate:
		return fmt.Sprintf("ORDER BY t.occurred_at %s, t.created_at %s, t.id %s", order, order, order)
	default:
		return fmt.Sprintf("ORDER BY t.occurred_at %s, t.created_at %s, t.id %s", order, order, order)
	}
}

func isPortfolioDuplicate(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && strings.Contains(strings.ToLower(pqErr.Constraint), "checksum")
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") && strings.Contains(msg, "checksum")
}

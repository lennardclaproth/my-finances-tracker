package storage

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
)

// SQLXCashflowStore persists and reads cashflow transactions, including the
// analytics read models. It satisfies the cashflow command and query contracts.
type SQLXCashflowStore struct {
	db        *DB
	tableName string
}

var (
	_ cashflow.CommandStore = (*SQLXCashflowStore)(nil)
	_ cashflow.QueryStore   = (*SQLXCashflowStore)(nil)
)

// NewSQLXCashflowStore creates a cashflow transaction store backed by SQLX.
func NewSQLXCashflowStore(db *DB) *SQLXCashflowStore {
	return &SQLXCashflowStore{
		db:        db,
		tableName: qualifyTable(db, SchemaCashflow, TableTransactions),
	}
}

// CreateTransactions persists a batch of cashflow transactions in a single insert,
// skipping rows whose checksum already exists, and returns the number inserted.
func (s *SQLXCashflowStore) CreateTransactions(ctx context.Context, txs []*cashflow.Transaction) (int, error) {
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
	res, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, txs)
	if err != nil {
		return 0, fmt.Errorf("cashflow store: bulk insert transactions: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("cashflow store: rows affected: %w", err)
	}
	return int(affected), nil
}

// ListTransactions returns a filtered, sorted, paginated page of cashflow transactions
// with the total match count.
func (s *SQLXCashflowStore) ListTransactions(ctx context.Context, query cashflow.TransactionListQuery) (*cashflow.TransactionListResult, error) {
	q := cashflowQueryFromList(query)
	limit := q.Limit
	if limit == 0 {
		limit = 100
	}

	whereClause, args := buildCashflowWhereClause(q)
	executor := s.db.GetExecutor(ctx)

	totalQuery := s.db.Rebind(fmt.Sprintf("SELECT COUNT(1) FROM %s%s", s.tableName, whereClause))
	total := 0
	if err := sqlx.GetContext(ctx, executor, &total, totalQuery, args...); err != nil {
		return nil, fmt.Errorf("cashflow store: count transactions: %w", err)
	}

	dataQuery := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		%s
		ORDER BY %s %s
		LIMIT ?
		OFFSET ?
	`, s.tableName, whereClause, normalizeCashflowSortBy(q.SortBy), normalizeSortOrder(q.SortOrder)))
	dataArgs := append(append([]any{}, args...), limit, q.Offset)

	rows, err := executor.QueryxContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, fmt.Errorf("cashflow store: fetch transactions: %w", err)
	}
	transactions, err := parseCashflowRows(rows)
	closeErr := rows.Close()
	if err != nil {
		return nil, fmt.Errorf("cashflow store: parse transaction rows: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("cashflow store: close transaction rows: %w", closeErr)
	}

	return &cashflow.TransactionListResult{Total: total, Transactions: transactions}, nil
}

// CountByFilter counts cashflow transactions matching application-level filters.
func (s *SQLXCashflowStore) CountByFilter(ctx context.Context, filters cashflow.TransactionFilters) (int, error) {
	whereClause, args := buildCashflowWhereClause(cashflowQueryFromFilters(filters))
	query := s.db.Rebind(fmt.Sprintf("SELECT COUNT(1) FROM %s%s", s.tableName, whereClause))
	total := 0
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &total, query, args...); err != nil {
		return 0, fmt.Errorf("cashflow store: count by filter: %w", err)
	}
	return total, nil
}

// UpdateTagByIDs sets the tag for the given transaction IDs and returns the count updated.
func (s *SQLXCashflowStore) UpdateTagByIDs(ctx context.Context, ids []uuid.UUID, tag string) (int, error) {
	return s.updateByIDs(ctx, "tag", tag, ids)
}

// UpdateIgnoredByIDs sets the ignored flag for the given transaction IDs and returns the count updated.
func (s *SQLXCashflowStore) UpdateIgnoredByIDs(ctx context.Context, ids []uuid.UUID, ignored bool) (int, error) {
	return s.updateByIDs(ctx, "ignored", ignored, ids)
}

// UpdateTagByFilter sets the tag for transactions matching filters and returns the count updated.
func (s *SQLXCashflowStore) UpdateTagByFilter(ctx context.Context, filters cashflow.TransactionFilters, tag string) (int, error) {
	return s.updateByFilter(ctx, "tag", tag, filters)
}

// UpdateIgnoredByFilter sets the ignored flag for transactions matching filters and returns the count updated.
func (s *SQLXCashflowStore) UpdateIgnoredByFilter(ctx context.Context, filters cashflow.TransactionFilters, ignored bool) (int, error) {
	return s.updateByFilter(ctx, "ignored", ignored, filters)
}

func (s *SQLXCashflowStore) updateByIDs(ctx context.Context, column string, value any, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	query := fmt.Sprintf(`UPDATE %s SET %s = ?, updated_at = ? WHERE id IN (?)`, s.tableName, column)
	expanded, args, err := sqlx.In(query, value, time.Now().UTC(), ids)
	if err != nil {
		return 0, fmt.Errorf("cashflow store: expand ids: %w", err)
	}
	return s.exec(ctx, s.db.Rebind(expanded), args...)
}

func (s *SQLXCashflowStore) updateByFilter(ctx context.Context, column string, value any, filters cashflow.TransactionFilters) (int, error) {
	whereClause, whereArgs := buildCashflowWhereClause(cashflowQueryFromFilters(filters))
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET %s = ?, updated_at = ?%s`, s.tableName, column, whereClause))
	args := append([]any{value, time.Now().UTC()}, whereArgs...)
	return s.exec(ctx, query, args...)
}

func (s *SQLXCashflowStore) exec(ctx context.Context, query string, args ...any) (int, error) {
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("cashflow store: update transactions: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("cashflow store: rows affected: %w", err)
	}
	return int(affected), nil
}

// GetMonthlyAnalytics returns incoming/outgoing/net cashflow totals grouped by month.
func (s *SQLXCashflowStore) GetMonthlyAnalytics(ctx context.Context, filter cashflow.AnalyticsFilter) ([]cashflow.MonthlyAnalyticsPoint, error) {
	monthExpr := "TO_CHAR(DATE_TRUNC('month', date), 'YYYY-MM-01')"
	if s.db.DriverName() == string(Sqlite) {
		monthExpr = "COALESCE(STRFTIME('%Y-%m-01', date), SUBSTR(CAST(date AS TEXT), 1, 7) || '-01')"
	}

	whereClause, args := buildCashflowAnalyticsWhereClause(filter)
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT
			%s AS month_start,
			COALESCE(SUM(CASE WHEN direction = 'in' THEN amount_cents ELSE 0 END), 0) AS incoming_cents,
			COALESCE(SUM(CASE WHEN direction = 'out' THEN amount_cents ELSE 0 END), 0) AS outgoing_cents
		FROM %s
		%s
		GROUP BY 1
		ORDER BY 1 ASC
	`, monthExpr, s.tableName, whereClause))

	type row struct {
		MonthStart    string `db:"month_start"`
		IncomingCents int64  `db:"incoming_cents"`
		OutgoingCents int64  `db:"outgoing_cents"`
	}
	var rows []row
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &rows, query, args...); err != nil {
		return nil, fmt.Errorf("cashflow store: fetch monthly analytics: %w", err)
	}

	points := make([]cashflow.MonthlyAnalyticsPoint, 0, len(rows))
	for _, r := range rows {
		month, err := time.Parse("2006-01-02", r.MonthStart)
		if err != nil {
			return nil, fmt.Errorf("cashflow store: parse month %q: %w", r.MonthStart, err)
		}
		points = append(points, cashflow.MonthlyAnalyticsPoint{
			Month:         month.UTC(),
			IncomingCents: r.IncomingCents,
			OutgoingCents: r.OutgoingCents,
			NetCents:      r.IncomingCents - r.OutgoingCents,
		})
	}
	return points, nil
}

// GetTagDistribution returns cashflow tag totals grouped by direction (combined/incoming/outgoing).
func (s *SQLXCashflowStore) GetTagDistribution(ctx context.Context, filter cashflow.AnalyticsFilter) (*cashflow.TagDistribution, error) {
	whereClause, args := buildCashflowAnalyticsWhereClause(filter)
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT
			COALESCE(NULLIF(TRIM(tag), ''), 'untagged') AS tag_name,
			direction,
			COALESCE(SUM(amount_cents), 0) AS total_cents
		FROM %s
		%s
		GROUP BY tag_name, direction
	`, s.tableName, whereClause))

	type row struct {
		TagName    string `db:"tag_name"`
		Direction  string `db:"direction"`
		TotalCents int64  `db:"total_cents"`
	}
	var rows []row
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &rows, query, args...); err != nil {
		return nil, fmt.Errorf("cashflow store: fetch tag distribution: %w", err)
	}

	combined := make(map[string]int64)
	incoming := make(map[string]int64)
	outgoing := make(map[string]int64)
	for _, r := range rows {
		tag := strings.TrimSpace(r.TagName)
		if tag == "" {
			tag = "untagged"
		}
		combined[tag] += r.TotalCents
		switch r.Direction {
		case "in":
			incoming[tag] += r.TotalCents
		case "out":
			outgoing[tag] += r.TotalCents
		}
	}

	return &cashflow.TagDistribution{
		Combined: sortedTagEntries(combined),
		Incoming: sortedTagEntries(incoming),
		Outgoing: sortedTagEntries(outgoing),
	}, nil
}

// cashflowQuery is the storage-internal filter/sort/paginate shape shared by the
// list and bulk-update paths.
type cashflowQuery struct {
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

func cashflowQueryFromList(query cashflow.TransactionListQuery) cashflowQuery {
	direction := ""
	if query.Direction != nil {
		direction = string(*query.Direction)
	}
	return cashflowQuery{
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
	}
}

func cashflowQueryFromFilters(filters cashflow.TransactionFilters) cashflowQuery {
	query := cashflowQuery{
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

func buildCashflowWhereClause(query cashflowQuery) (string, []any) {
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

func buildCashflowAnalyticsWhereClause(filter cashflow.AnalyticsFilter) (string, []any) {
	var (
		conditions []string
		args       []any
	)
	if !filter.IncludeIgnored {
		conditions = append(conditions, "ignored = ?")
		args = append(args, false)
	}
	if filter.From != nil {
		conditions = append(conditions, "date >= ?")
		args = append(args, *filter.From)
	}
	if filter.To != nil {
		conditions = append(conditions, "date <= ?")
		args = append(args, *filter.To)
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
	default:
		return "date"
	}
}

func normalizeSortOrder(sortOrder string) string {
	if strings.ToLower(strings.TrimSpace(sortOrder)) == "asc" {
		return "ASC"
	}
	return "DESC"
}

func sortedTagEntries(input map[string]int64) []cashflow.TagDistributionEntry {
	out := make([]cashflow.TagDistributionEntry, 0, len(input))
	for tag, total := range input {
		out = append(out, cashflow.TagDistributionEntry{Tag: tag, TotalCents: total})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TotalCents == out[j].TotalCents {
			return out[i].Tag < out[j].Tag
		}
		return out[i].TotalCents > out[j].TotalCents
	})
	return out
}

func parseCashflowRows(rows *sqlx.Rows) ([]*cashflow.Transaction, error) {
	transactions := []*cashflow.Transaction{}
	for rows.Next() {
		var tx cashflow.Transaction
		if err := rows.StructScan(&tx); err != nil {
			return nil, fmt.Errorf("cashflow store: scan transaction: %w", err)
		}
		transactions = append(transactions, &tx)
	}
	return transactions, rows.Err()
}

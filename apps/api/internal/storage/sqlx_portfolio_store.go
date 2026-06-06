package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

// SQLXPortfolioStore persists and reads the portfolio aggregate: account projection,
// transactions, positions, position snapshots, and portfolio snapshots. It satisfies
// the portfolio command, query, builder, and lock contracts.
type SQLXPortfolioStore struct {
	db                 *DB
	accountsTable      string
	transactionsTable  string
	positionsTable     string
	posSnapshotsTable  string
	portSnapshotsTable string
}

var (
	_ portfolio.CommandStore     = (*SQLXPortfolioStore)(nil)
	_ portfolio.QueryStore       = (*SQLXPortfolioStore)(nil)
	_ portfolio.PositionStore    = (*SQLXPortfolioStore)(nil)
	_ portfolio.PortfolioStore   = (*SQLXPortfolioStore)(nil)
	_ portfolio.TransactionStore = (*SQLXPortfolioStore)(nil)
	_ portfolio.Locker           = (*SQLXPortfolioStore)(nil)
)

// NewSQLXPortfolioStore creates a portfolio store backed by SQLX.
func NewSQLXPortfolioStore(db *DB) *SQLXPortfolioStore {
	return &SQLXPortfolioStore{
		db:                 db,
		accountsTable:      qualifyTableAs(db, SchemaPortfolio, TableAccounts, "portfolio_accounts"),
		transactionsTable:  qualifyTableAs(db, SchemaPortfolio, TableTransactions, "portfolio_transactions"),
		positionsTable:     qualifyTable(db, SchemaPortfolio, TablePositions),
		posSnapshotsTable:  qualifyTable(db, SchemaPortfolio, TablePosSnapshots),
		portSnapshotsTable: qualifyTable(db, SchemaPortfolio, TablePortSnapshots),
	}
}

// --- Account projection -----------------------------------------------------

// CreateAccount inserts the portfolio account projection, ignoring an existing row.
func (s *SQLXPortfolioStore) CreateAccount(ctx context.Context, acc *portfolio.Account) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, building, created_at, updated_at)
		VALUES (:id, :account_id, :building, :created_at, :updated_at)
		ON CONFLICT (account_id) DO NOTHING
	`, s.accountsTable)
	_, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, acc)
	return err
}

// TryAcquireBuildLock atomically claims the account's build lock, returning true on
// success or portfolio.ErrAccountNotFound when the projection does not exist.
func (s *SQLXPortfolioStore) TryAcquireBuildLock(ctx context.Context, id uuid.UUID) (bool, error) {
	executor := s.db.GetExecutor(ctx)
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s SET building = ?, updated_at = ?
		WHERE account_id = ? AND building = ?
	`, s.accountsTable))
	res, err := executor.ExecContext(ctx, query, true, time.Now().UTC(), id, false)
	if err != nil {
		return false, fmt.Errorf("portfolio store: acquire build lock: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("portfolio store: acquire build lock rows: %w", err)
	}
	if n == 1 {
		return true, nil
	}

	var exists int
	checkQuery := s.db.Rebind(fmt.Sprintf("SELECT 1 FROM %s WHERE account_id = ?", s.accountsTable))
	if err := sqlx.GetContext(ctx, executor, &exists, checkQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, portfolio.ErrAccountNotFound
		}
		return false, fmt.Errorf("portfolio store: check account: %w", err)
	}
	return false, nil
}

// ReleaseBuildLock clears the account's build lock.
func (s *SQLXPortfolioStore) ReleaseBuildLock(ctx context.Context, id uuid.UUID) error {
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET building = ?, updated_at = ? WHERE account_id = ?`, s.accountsTable))
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, false, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("portfolio store: release build lock: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("portfolio store: release build lock rows: %w", err)
	}
	if n == 0 {
		return portfolio.ErrAccountNotFound
	}
	return nil
}

// --- Transactions -----------------------------------------------------------

const portfolioTxColumns = `
	id, account_id, import_id, origin, source, occurred_at, position_id, isin, symbol,
	description, type, quantity, unit_price, amount_cents, checksum, row_number,
	created_at, updated_at
`

const portfolioTxValues = `
	:id, :account_id, :import_id, :origin, :source, :occurred_at, :position_id, :isin, :symbol,
	:description, :type, :quantity, :unit_price, :amount_cents, :checksum, :row_number,
	:created_at, :updated_at
`

// CreateTransaction persists one portfolio transaction, mapping a duplicate checksum
// to portfolio.ErrDuplicateTransaction.
func (s *SQLXPortfolioStore) CreateTransaction(ctx context.Context, tx *portfolio.Transaction) error {
	query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, s.transactionsTable, portfolioTxColumns, portfolioTxValues)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, tx); err != nil {
		if isUniqueViolation(err) {
			return portfolio.ErrDuplicateTransaction
		}
		return fmt.Errorf("portfolio store: create transaction: %w", err)
	}
	return nil
}

// CreateTransactions persists a batch of portfolio transactions in a single insert,
// skipping rows whose checksum already exists, and returns the number inserted.
func (s *SQLXPortfolioStore) CreateTransactions(ctx context.Context, txs []*portfolio.Transaction) (int, error) {
	if len(txs) == 0 {
		return 0, nil
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (checksum) DO NOTHING
	`, s.transactionsTable, portfolioTxColumns, portfolioTxValues)
	res, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, txs)
	if err != nil {
		return 0, fmt.Errorf("portfolio store: bulk insert transactions: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("portfolio store: rows affected: %w", err)
	}
	return int(affected), nil
}

// TransactionsForAccount returns an account's transactions ordered by occurrence
// (sort "asc"/"desc").
func (s *SQLXPortfolioStore) TransactionsForAccount(ctx context.Context, accID uuid.UUID, sort string) ([]portfolio.Transaction, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		WHERE account_id = ?
		ORDER BY occurred_at %s, row_number %s
	`, s.transactionsTable, normalizeSortOrder(sort), normalizeSortOrder(sort)))
	var txs []portfolio.Transaction
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &txs, query, accID); err != nil {
		return nil, fmt.Errorf("portfolio store: transactions for account: %w", err)
	}
	return txs, nil
}

// TransactionsForPosition returns a position's transactions on or after the optional
// from date, ordered by occurrence.
func (s *SQLXPortfolioStore) TransactionsForPosition(ctx context.Context, positionID uuid.UUID, from *time.Time) ([]portfolio.Transaction, error) {
	query := fmt.Sprintf(`SELECT * FROM %s WHERE position_id = ?`, s.transactionsTable)
	args := []any{positionID}
	if from != nil {
		query += " AND occurred_at >= ?"
		args = append(args, *from)
	}
	query += " ORDER BY occurred_at ASC, row_number ASC"
	var txs []portfolio.Transaction
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &txs, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("portfolio store: transactions for position: %w", err)
	}
	return txs, nil
}

// UpdatePositions maps each transaction to its resolved position.
func (s *SQLXPortfolioStore) UpdatePositions(ctx context.Context, transactions []portfolio.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET position_id = ?, updated_at = ? WHERE id = ?`, s.transactionsTable))
	executor := s.db.GetExecutor(ctx)
	now := time.Now().UTC()
	for i := range transactions {
		tx := transactions[i]
		if tx.PositionID == nil {
			continue
		}
		if _, err := executor.ExecContext(ctx, query, *tx.PositionID, now, tx.ID); err != nil {
			return fmt.Errorf("portfolio store: update position mapping: %w", err)
		}
	}
	return nil
}

// --- Positions --------------------------------------------------------------

// CreateMany persists a batch of positions.
func (s *SQLXPortfolioStore) CreateMany(ctx context.Context, positions []*portfolio.Position) error {
	if len(positions) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, isin, symbol, listing_id,
			open_date, close_date, quantity, cost_basis,
			fees, income, taxes, realized_pnl,
			created_at, updated_at
		) VALUES (
			:id, :account_id, :isin, :symbol, :listing_id,
			:open_date, :close_date, :quantity, :cost_basis,
			:fees, :income, :taxes, :realized_pnl,
			:created_at, :updated_at
		)
	`, s.positionsTable)
	executor := s.db.GetExecutor(ctx)
	for _, pos := range positions {
		if pos.CreatedAt.IsZero() {
			pos.CreatedAt = time.Now().UTC()
		}
		if pos.UpdatedAt.IsZero() {
			pos.UpdatedAt = time.Now().UTC()
		}
		if _, err := sqlx.NamedExecContext(ctx, executor, query, pos); err != nil {
			return fmt.Errorf("portfolio store: create position: %w", err)
		}
	}
	return nil
}

// GetLastSnapshot returns the most recent position snapshot, or (nil, nil) when none.
func (s *SQLXPortfolioStore) GetLastSnapshot(ctx context.Context, positionID uuid.UUID) (*portfolio.PositionSnapshot, error) {
	var snap portfolio.PositionSnapshot
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s WHERE position_id = ? ORDER BY occurred_at DESC LIMIT 1
	`, s.posSnapshotsTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &snap, query, positionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("portfolio store: get last snapshot: %w", err)
	}
	return &snap, nil
}

// CreatePositionSnapshot persists one position snapshot, mapping the per-day unique
// violation to portfolio.ErrPositionSnapshotAlreadyExists.
func (s *SQLXPortfolioStore) CreatePositionSnapshot(ctx context.Context, snap *portfolio.PositionSnapshot) error {
	if snap.CreatedAt.IsZero() {
		snap.CreatedAt = time.Now().UTC()
	}
	snap.OccurredAt = date.StartOfDayUTC(snap.OccurredAt)
	snap.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, position_id, symbol, name, listing_id, occurred_at,
			quantity, unit_price, average_price, market_value,
			cost_basis, income, fees, taxes,
			total_pnl, total_pnl_pct, realized_pnl, unrealized_pnl, unrealized_pnl_pct,
			daily_delta_pnl, daily_delta_pnl_pct,
			created_at, updated_at
		) VALUES (
			:id, :account_id, :position_id, :symbol, :name, :listing_id, :occurred_at,
			:quantity, :unit_price, :average_price, :market_value,
			:cost_basis, :income, :fees, :taxes,
			:total_pnl, :total_pnl_pct, :realized_pnl, :unrealized_pnl, :unrealized_pnl_pct,
			:daily_delta_pnl, :daily_delta_pnl_pct,
			:created_at, :updated_at
		)
	`, s.posSnapshotsTable)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, snap); err != nil {
		if isUniqueViolation(err) {
			return portfolio.ErrPositionSnapshotAlreadyExists
		}
		return fmt.Errorf("portfolio store: create position snapshot: %w", err)
	}
	return nil
}

// --- Portfolio snapshots ----------------------------------------------------

// CreatePortfolioSnapshot upserts one portfolio snapshot keyed by (account, day).
func (s *SQLXPortfolioStore) CreatePortfolioSnapshot(ctx context.Context, snapshot *portfolio.PortfolioSnapshot) error {
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Now().UTC()
	}
	snapshot.OccurredAt = date.StartOfDayUTC(snapshot.OccurredAt)
	snapshot.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, account_id, occurred_at,
			market_value, cost_basis,
			realized_pnl, unrealized_pnl, unrealized_pnl_pct, income, fees, taxes, cash_balance,
			total_pnl, total_pnl_pct,
			daily_delta_pnl, daily_delta_pnl_pct, time_weighted_return_pct,
			net_cashflow, cumulative_net_cashflow,
			created_at, updated_at
		) VALUES (
			:id, :account_id, :occurred_at,
			:market_value, :cost_basis,
			:realized_pnl, :unrealized_pnl, :unrealized_pnl_pct, :income, :fees, :taxes, :cash_balance,
			:total_pnl, :total_pnl_pct,
			:daily_delta_pnl, :daily_delta_pnl_pct, :time_weighted_return_pct,
			:net_cashflow, :cumulative_net_cashflow,
			:created_at, :updated_at
		)
		ON CONFLICT (account_id, occurred_at) DO UPDATE SET
			market_value = EXCLUDED.market_value,
			cost_basis = EXCLUDED.cost_basis,
			realized_pnl = EXCLUDED.realized_pnl,
			unrealized_pnl = EXCLUDED.unrealized_pnl,
			unrealized_pnl_pct = EXCLUDED.unrealized_pnl_pct,
			income = EXCLUDED.income,
			fees = EXCLUDED.fees,
			taxes = EXCLUDED.taxes,
			cash_balance = EXCLUDED.cash_balance,
			total_pnl = EXCLUDED.total_pnl,
			total_pnl_pct = EXCLUDED.total_pnl_pct,
			daily_delta_pnl = EXCLUDED.daily_delta_pnl,
			daily_delta_pnl_pct = EXCLUDED.daily_delta_pnl_pct,
			time_weighted_return_pct = EXCLUDED.time_weighted_return_pct,
			net_cashflow = EXCLUDED.net_cashflow,
			cumulative_net_cashflow = EXCLUDED.cumulative_net_cashflow,
			updated_at = EXCLUDED.updated_at
	`, s.portSnapshotsTable)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, snapshot); err != nil {
		return fmt.Errorf("portfolio store: create portfolio snapshot: %w", err)
	}
	return nil
}

// Clean removes all built portfolio data for an account ahead of a full rebuild:
// position snapshots, portfolio snapshots, and positions.
func (s *SQLXPortfolioStore) Clean(ctx context.Context, accID uuid.UUID) error {
	executor := s.db.GetExecutor(ctx)
	for _, table := range []string{s.posSnapshotsTable, s.portSnapshotsTable, s.positionsTable} {
		query := s.db.Rebind(fmt.Sprintf("DELETE FROM %s WHERE account_id = ?", table))
		if _, err := executor.ExecContext(ctx, query, accID); err != nil {
			return fmt.Errorf("portfolio store: clean %s: %w", table, err)
		}
	}
	return nil
}

// --- Queries ----------------------------------------------------------------

// SnapshotsForAccount returns portfolio snapshots for an account within an optional
// date range, ordered by occurrence (default ascending), optionally paginated.
func (s *SQLXPortfolioStore) SnapshotsForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	limit, offset *int,
	from, to *time.Time,
	sort *sorting.Direction,
) ([]*portfolio.PortfolioSnapshot, error) {
	dir := "ASC"
	if sort != nil {
		dir = normalizeSortOrder(string(*sort))
	}
	query := fmt.Sprintf(`SELECT * FROM %s WHERE account_id = ?`, s.portSnapshotsTable)
	args := []any{accountID}
	if from != nil {
		query += " AND occurred_at >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND occurred_at <= ?"
		args = append(args, *to)
	}
	query += fmt.Sprintf(" ORDER BY occurred_at %s, id ASC", dir)
	if limit != nil {
		query += " LIMIT ?"
		args = append(args, *limit)
		if offset != nil {
			query += " OFFSET ?"
			args = append(args, *offset)
		}
	}

	var snapshots []*portfolio.PortfolioSnapshot
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &snapshots, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("portfolio store: snapshots for account: %w", err)
	}
	return snapshots, nil
}

// PositionsWithLatestSnapshot returns account positions joined with their latest
// position snapshot, optionally including closed positions.
func (s *SQLXPortfolioStore) PositionsWithLatestSnapshot(
	ctx context.Context,
	accID uuid.UUID,
	includeClosed bool,
) ([]*portfolio.PositionWithLatestSnapshot, error) {
	query := fmt.Sprintf(`
		SELECT
			p.id,
			p.symbol,
			ls.name,
			p.quantity,
			p.cost_basis,
			p.realized_pnl,
			ls.market_value,
			ls.unrealized_pnl_pct,
			ls.occurred_at AS last_snapshot_at,
			p.open_date,
			p.close_date,
			(p.close_date IS NOT NULL) AS is_closed
		FROM %s p
		LEFT JOIN %s ls
			ON ls.id = (
				SELECT ps.id
				FROM %s ps
				WHERE ps.position_id = p.id
				ORDER BY ps.occurred_at DESC, ps.id DESC
				LIMIT 1
			)
		WHERE p.account_id = ?
	`, s.positionsTable, s.posSnapshotsTable, s.posSnapshotsTable)
	args := []any{accID}
	if !includeClosed {
		query += " AND p.close_date IS NULL"
	}
	query += " ORDER BY p.open_date ASC, p.id ASC"

	var rows []*portfolio.PositionWithLatestSnapshot
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &rows, s.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("portfolio store: positions with latest snapshot: %w", err)
	}
	return rows, nil
}

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

type SQLXPositionStore struct {
	db             *DB
	positionsTable string
	snapshotsTable string
}

func NewSQLXPositionStore(db *DB) *SQLXPositionStore {
	return &SQLXPositionStore{
		db:             db,
		positionsTable: qualifyTable(db, SchemaPortfolio, TablePositions),
		snapshotsTable: qualifyTable(db, SchemaPortfolio, TablePosSnapshots),
	}
}

var _ portfolio.PositionStore = (*SQLXPositionStore)(nil)
var _ portfolio.PositionFetcher = (*SQLXPositionStore)(nil)

func (s *SQLXPositionStore) Clean(ctx context.Context, accID uuid.UUID) error {
	executor := s.db.GetExecutor(ctx)
	snapQuery := s.db.Rebind(fmt.Sprintf("DELETE FROM %s WHERE account_id = ?", s.snapshotsTable))
	if _, err := executor.ExecContext(ctx, snapQuery, accID); err != nil {
		return fmt.Errorf("sqlx_position_store: clean snapshots: %w", err)
	}
	posQuery := s.db.Rebind(fmt.Sprintf("DELETE FROM %s WHERE account_id = ?", s.positionsTable))
	if _, err := executor.ExecContext(ctx, posQuery, accID); err != nil {
		return fmt.Errorf("sqlx_position_store: clean positions: %w", err)
	}
	return nil
}

func (s *SQLXPositionStore) CreateMany(ctx context.Context, positions []*portfolio.Position) error {
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
			return fmt.Errorf("sqlx_position_store: create position: %w", err)
		}
	}
	return nil
}

func (s *SQLXPositionStore) GetByID(ctx context.Context, id uuid.UUID) (*portfolio.Position, error) {
	var pos portfolio.Position
	query := s.db.Rebind(fmt.Sprintf("SELECT * FROM %s WHERE id = ?", s.positionsTable))
	executor := s.db.GetExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &pos, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("sqlx_position_store: get by id: %w", err)
	}
	return &pos, nil
}

func (s *SQLXPositionStore) GetLastSnapshot(ctx context.Context, positionID uuid.UUID) (*portfolio.PositionSnapshot, error) {
	var snap portfolio.PositionSnapshot
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		WHERE position_id = ?
		ORDER BY occurred_at DESC
		LIMIT 1
	`, s.snapshotsTable))
	executor := s.db.GetExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &snap, query, positionID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("sqlx_position_store: get last snapshot: %w", err)
	}
	return &snap, nil
}

func (s *SQLXPositionStore) CreateSnapshot(ctx context.Context, snap *portfolio.PositionSnapshot) error {
	if snap.CreatedAt.IsZero() {
		snap.CreatedAt = time.Now().UTC()
	}
	snap.OccurredAt = time.Date(
		snap.OccurredAt.UTC().Year(),
		snap.OccurredAt.UTC().Month(),
		snap.OccurredAt.UTC().Day(),
		0, 0, 0, 0,
		time.UTC,
	)
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
	`, s.snapshotsTable)
	executor := s.db.GetExecutor(ctx)
	if _, err := sqlx.NamedExecContext(ctx, executor, query, snap); err != nil {
		if isPositionSnapshotDuplicate(err) {
			return portfolio.ErrPositionSnapshotAlreadyExists
		}
		return fmt.Errorf("sqlx_position_store: create snapshot: %w", err)
	}
	return nil
}

func isPositionSnapshotDuplicate(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" && strings.Contains(strings.ToLower(pqErr.Constraint), "position_snapshot")
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") &&
		strings.Contains(msg, "position_id") &&
		strings.Contains(msg, "occurred_at")
}

func (s *SQLXPositionStore) GetForAccount(ctx context.Context, accID uuid.UUID) ([]*portfolio.Position, error) {
	query := s.db.Rebind(fmt.Sprintf("SELECT * FROM %s WHERE account_id = ? ORDER BY open_date ASC", s.positionsTable))
	executor := s.db.GetExecutor(ctx)
	var positions []*portfolio.Position
	if err := sqlx.SelectContext(ctx, executor, &positions, query, accID); err != nil {
		return nil, fmt.Errorf("sqlx_position_store: get for account: %w", err)
	}
	return positions, nil
}

func (s *SQLXPositionStore) GetSnapshotsForAccount(ctx context.Context, accID uuid.UUID) ([]*portfolio.PositionSnapshot, error) {
	query := s.db.Rebind(fmt.Sprintf("SELECT * FROM %s WHERE account_id = ? ORDER BY occurred_at ASC, position_id ASC", s.snapshotsTable))
	executor := s.db.GetExecutor(ctx)
	var snaps []*portfolio.PositionSnapshot
	if err := sqlx.SelectContext(ctx, executor, &snaps, query, accID); err != nil {
		return nil, fmt.Errorf("sqlx_position_store: get snapshots for account: %w", err)
	}
	return snaps, nil
}

func (s *SQLXPositionStore) ListForAccountWithLatestSnapshot(
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
	`, s.positionsTable, s.snapshotsTable, s.snapshotsTable)

	args := []any{accID}
	if !includeClosed {
		query += " AND p.close_date IS NULL"
	}
	query += " ORDER BY p.open_date ASC, p.id ASC"
	query = s.db.Rebind(query)

	executor := s.db.GetExecutor(ctx)
	var rows []*portfolio.PositionWithLatestSnapshot
	if err := sqlx.SelectContext(ctx, executor, &rows, query, args...); err != nil {
		return nil, fmt.Errorf("sqlx_position_store: list for account with latest snapshot: %w", err)
	}
	return rows, nil
}

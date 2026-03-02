package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

type SQLXPortfolioSnapshotStore struct {
	db        *DB
	tableName string
}

func NewSQLXPortfolioSnapshotStore(db *DB) *SQLXPortfolioSnapshotStore {
	return &SQLXPortfolioSnapshotStore{
		db:        db,
		tableName: qualifyTable(db, SchemaPortfolio, TablePortSnapshots),
	}
}

var _ portfolio.PortfolioStore = (*SQLXPortfolioSnapshotStore)(nil)

func (s *SQLXPortfolioSnapshotStore) Clean(ctx context.Context, accID uuid.UUID) error {
	executor := s.db.GetExecutor(ctx)
	snapQuery := s.db.Rebind(fmt.Sprintf("DELETE FROM %s WHERE account_id = ?", s.tableName))
	if _, err := executor.ExecContext(ctx, snapQuery, accID); err != nil {
		return fmt.Errorf("sqlx_portfolio_snapshot_store: clean snapshots: %w", err)
	}
	return nil
}

func (s *SQLXPortfolioSnapshotStore) CreateSnapshot(ctx context.Context, snapshot *portfolio.PortfolioSnapshot) error {
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Now().UTC()
	}
	snapshot.OccurredAt = time.Date(
		snapshot.OccurredAt.UTC().Year(),
		snapshot.OccurredAt.UTC().Month(),
		snapshot.OccurredAt.UTC().Day(),
		0, 0, 0, 0,
		time.UTC,
	)
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
	`, s.tableName)
	executor := s.db.GetExecutor(ctx)
	if _, err := sqlx.NamedExecContext(ctx, executor, query, snapshot); err != nil {
		return fmt.Errorf("sqlx_portfolio_snapshot_store: create snapshot: %w", err)
	}
	return nil
}

func (s *SQLXPortfolioSnapshotStore) ListForAccount(
	ctx context.Context,
	accountID uuid.UUID,
	from, to *time.Time,
) ([]*portfolio.PortfolioSnapshot, error) {
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE account_id = ?
	`, s.tableName)
	args := []any{accountID}
	if from != nil {
		query += " AND occurred_at >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND occurred_at <= ?"
		args = append(args, *to)
	}
	query += " ORDER BY occurred_at ASC, id ASC"
	query = s.db.Rebind(query)

	var snapshots []*portfolio.PortfolioSnapshot
	executor := s.db.GetExecutor(ctx)
	if err := sqlx.SelectContext(ctx, executor, &snapshots, query, args...); err != nil {
		return nil, fmt.Errorf("sqlx_portfolio_snapshot_store: list for account: %w", err)
	}
	return snapshots, nil
}

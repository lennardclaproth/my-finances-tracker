package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

type SQLXPortfolioAccountStore struct {
	db        *DB
	tableName string
}

func NewSQLXPortfolioAccountStore(db *DB) *SQLXPortfolioAccountStore {
	return &SQLXPortfolioAccountStore{
		db:        db,
		tableName: qualifyPortfolioAccountsTable(db),
	}
}

var _ portfolio.AccountStore = (*SQLXPortfolioAccountStore)(nil)

func qualifyPortfolioAccountsTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "portfolio_accounts"
	}
	return fmt.Sprintf("%s.%s", SchemaPortfolio, TableAccounts)
}

func (s *SQLXPortfolioAccountStore) Create(ctx context.Context, acc *portfolio.Account) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, building, created_at, updated_at)
		VALUES (:id, :account_id, :building, :created_at, :updated_at)
		ON CONFLICT (account_id) DO NOTHING
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, acc)
	return err
}

func (s *SQLXPortfolioAccountStore) TryAcquireBuildLock(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET building = true, updated_at = $1
		WHERE account_id = $2 AND building = false
	`, s.tableName)
	res, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return false, fmt.Errorf("sqlx_portfolio_account_store: acquire build lock: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("sqlx_portfolio_account_store: acquire build lock rows: %w", err)
	}
	if n == 1 {
		return true, nil
	}
	executor := s.db.GetExecutor(ctx)
	var exists int
	checkQuery := fmt.Sprintf("SELECT 1 FROM %s WHERE account_id = $1", s.tableName)
	if err := sqlx.GetContext(ctx, executor, &exists, checkQuery, id); err != nil {
		if err == sql.ErrNoRows {
			return false, portfolio.ErrAccountNotFound
		}
		return false, fmt.Errorf("sqlx_portfolio_account_store: check account: %w", err)
	}
	return false, nil
}

func (s *SQLXPortfolioAccountStore) ReleaseBuildLock(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET building = false, updated_at = $1
		WHERE account_id = $2
	`, s.tableName)
	res, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("sqlx_portfolio_account_store: release build lock: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlx_portfolio_account_store: release build lock rows: %w", err)
	}
	if n == 0 {
		return portfolio.ErrAccountNotFound
	}
	return nil
}

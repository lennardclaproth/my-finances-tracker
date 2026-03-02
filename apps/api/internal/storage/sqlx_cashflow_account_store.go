package storage

import (
	"context"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
)

type SQLXCashflowAccountStore struct {
	db        *DB
	tableName string
}

func NewSQLXCashflowAccountStore(db *DB) *SQLXCashflowAccountStore {
	return &SQLXCashflowAccountStore{
		db:        db,
		tableName: qualifyCashflowAccountsTable(db),
	}
}

func qualifyCashflowAccountsTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "cashflow_accounts"
	}
	return fmt.Sprintf("%s.%s", SchemaCashflow, TableAccounts)
}

func (s *SQLXCashflowAccountStore) Create(ctx context.Context, acc *cashflow.Account) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, created_at, updated_at)
		VALUES (:id, :account_id, :created_at, :updated_at)
		ON CONFLICT (account_id) DO NOTHING
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, acc)
	return err
}

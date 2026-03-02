package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
)

type SQLXImportAccountStore struct {
	db        *DB
	tableName string
}

func NewSQLXImportAccountStore(db *DB) *SQLXImportAccountStore {
	return &SQLXImportAccountStore{
		db:        db,
		tableName: qualifyImportAccountsTable(db),
	}
}

func qualifyImportAccountsTable(db *DB) string {
	if db == nil || db.DriverName() == string(Sqlite) {
		return "import_accounts"
	}
	return fmt.Sprintf("%s.%s", SchemaImports, TableAccounts)
}

func (s *SQLXImportAccountStore) Create(ctx context.Context, acc *importer.Account) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, account_id, created_at, updated_at)
		VALUES (:id, :account_id, :created_at, :updated_at)
		ON CONFLICT (account_id) DO NOTHING
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, acc)
	return err
}

func (s *SQLXImportAccountStore) FetchByID(ctx context.Context, id uuid.UUID) (*importer.Account, error) {
	var acc importer.Account
	query := fmt.Sprintf(`SELECT id, account_id, created_at, updated_at FROM %s WHERE account_id = $1`, s.tableName)
	err := s.db.GetContext(ctx, &acc, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, importer.ErrImportAccountNotFound
		}
		return nil, err
	}
	return &acc, nil
}

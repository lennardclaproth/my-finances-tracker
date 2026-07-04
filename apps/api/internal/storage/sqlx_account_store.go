package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/my-finances-tracker/internal/account"
)

// SQLXAccountStore persists and reads account records. It satisfies the account
// feature's storage contracts (create, and query existence/lookup).
type SQLXAccountStore struct {
	db        *DB
	tableName string
}

var (
	_ account.CommandStore = (*SQLXAccountStore)(nil)
	_ account.QueryStore   = (*SQLXAccountStore)(nil)
)

// NewSQLXAccountStore creates an account store backed by SQLX.
func NewSQLXAccountStore(db *DB) *SQLXAccountStore {
	return &SQLXAccountStore{
		db:        db,
		tableName: qualifyTable(db, SchemaAccount, TableAccounts),
	}
}

// Create inserts a new account row, mapping a unique-constraint violation to
// account.ErrAccountAlreadyExists.
func (s *SQLXAccountStore) Create(ctx context.Context, acc *account.Account) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, external_id, name, created_at, updated_at)
		VALUES (:id, :external_id, :name, :created_at, :updated_at)
	`, s.tableName)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, acc); err != nil {
		if isUniqueViolation(err) {
			return account.ErrAccountAlreadyExists
		}
		return err
	}
	return nil
}

// GetByID returns one account by ID, or account.ErrAccountNotFound when absent.
func (s *SQLXAccountStore) GetByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	var acc account.Account
	query := s.db.Rebind(fmt.Sprintf(`SELECT id, external_id, name, created_at, updated_at FROM %s WHERE id = ?`, s.tableName))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &acc, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	return &acc, nil
}

// List returns all accounts ordered by creation time (oldest first).
func (s *SQLXAccountStore) List(ctx context.Context) ([]*account.Account, error) {
	var accounts []*account.Account
	query := fmt.Sprintf(`SELECT id, external_id, name, created_at, updated_at FROM %s ORDER BY created_at ASC`, s.tableName)
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &accounts, query); err != nil {
		return nil, err
	}
	return accounts, nil
}

// Exists reports whether an account with the given ID exists.
func (s *SQLXAccountStore) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	query := s.db.Rebind(fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE id = ?)`, s.tableName))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &exists, query, id); err != nil {
		return false, err
	}
	return exists, nil
}

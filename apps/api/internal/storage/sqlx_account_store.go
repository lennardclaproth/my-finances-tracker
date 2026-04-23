package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lib/pq"
)

// SQLXAccountStore persists and reads account records.
type SQLXAccountStore struct {
	db        *DB
	tableName string
}

// NewSQLXAccountStore creates an account store backed by SQLX.
func NewSQLXAccountStore(db *DB) *SQLXAccountStore {
	return &SQLXAccountStore{
		db:        db,
		tableName: qualifyTable(db, SchemaAccount, TableAccounts),
	}
}

// Create inserts a new account row.
func (s *SQLXAccountStore) Create(ctx context.Context, acc *account.Account) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, external_id, name, created_at, updated_at)
		VALUES (:id, :external_id, :name, :created_at, :updated_at)
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, acc)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return account.ErrAccountAlreadyExists
		}
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "unique") && strings.Contains(msg, "name") {
			return account.ErrAccountAlreadyExists
		}
		return err
	}
	return nil
}

// List returns all accounts sorted by name.
func (s *SQLXAccountStore) List(ctx context.Context) ([]*account.Account, error) {
	var accounts []*account.Account
	query := fmt.Sprintf(`SELECT id, external_id, name, created_at, updated_at FROM %s ORDER BY name ASC`, s.tableName)
	if err := s.db.SelectContext(ctx, &accounts, query); err != nil {
		return nil, err
	}
	return accounts, nil
}

// FetchByID returns one account by ID.
func (s *SQLXAccountStore) FetchByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	var acc account.Account
	query := fmt.Sprintf(`SELECT id, external_id, name, created_at, updated_at FROM %s WHERE id = $1`, s.tableName)
	err := s.db.GetContext(ctx, &acc, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	return &acc, nil
}

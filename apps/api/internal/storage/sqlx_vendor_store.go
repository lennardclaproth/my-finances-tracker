package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
	"github.com/lib/pq"
)

type SQLXVendorStore struct {
	db        *DB
	tableName string
}

func NewSQLXVendorStore(db *DB) *SQLXVendorStore {
	return &SQLXVendorStore{
		db:        db,
		tableName: qualifyTable(db, SchemaVendors, TableVendors),
	}
}

func (s *SQLXVendorStore) Create(ctx context.Context, v *vendor.Vendor) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, type, active, import_disabled, created_at, updated_at)
		VALUES (:id, :name, :type, :active, :import_disabled, :created_at, :updated_at)
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, v)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// 23505 = unique_violation
		if pqErr.Code == "23505" {
			// optional: check constraint name if you have multiple uniques
			if pqErr.Constraint == "vendors_name_key" {
				return vendor.ErrVendorAlreadyExists
			}
		}
	}
	return err
}

func (s *SQLXVendorStore) FetchByName(ctx context.Context, name vendor.VendorID) (*vendor.Vendor, error) {
	var v vendor.Vendor
	query := s.db.Rebind(fmt.Sprintf(`SELECT id, name, type, active, import_disabled, created_at, updated_at FROM %s WHERE name = ?`, s.tableName))
	executor := s.db.GetExecutor(ctx)
	err := sqlx.GetContext(ctx, executor, &v, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, vendor.ErrVendorNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (s *SQLXVendorStore) FetchById(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error) {
	var v vendor.Vendor
	query := s.db.Rebind(fmt.Sprintf(`SELECT id, name, type, active, import_disabled, created_at, updated_at FROM %s WHERE id = ?`, s.tableName))
	executor := s.db.GetExecutor(ctx)
	err := sqlx.GetContext(ctx, executor, &v, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, vendor.ErrVendorNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (s *SQLXVendorStore) ListActive(ctx context.Context) ([]*vendor.Vendor, error) {
	var vendors []*vendor.Vendor
	query := fmt.Sprintf(`
		SELECT id, name, type, active, import_disabled, created_at, updated_at
		FROM %s
		WHERE active = $1
		ORDER BY name ASC
	`, s.tableName)
	if err := s.db.SelectContext(ctx, &vendors, query, true); err != nil {
		return nil, err
	}
	return vendors, nil
}

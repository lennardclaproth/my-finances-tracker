package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
	"github.com/lib/pq"
)

type SQLXVendorStore struct {
	db *DB
}

func NewSQLXVendorStore(db *DB) *SQLXVendorStore {
	return &SQLXVendorStore{db: db}
}

func (s *SQLXVendorStore) Create(ctx context.Context, v *vendor.Vendor) error {
	query := `
		INSERT INTO vendors (id, name, created_at, updated_at)
		VALUES (:id, :name, :created_at, :updated_at)
	`
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
	query := `SELECT id, name, created_at, updated_at FROM vendors WHERE name = $1`
	err := s.db.GetContext(ctx, &v, query, name)
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
	query := `SELECT id, name, created_at, updated_at FROM vendors WHERE id = $1`
	err := s.db.GetContext(ctx, &v, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, vendor.ErrVendorNotFound
		}
		return nil, err
	}
	return &v, nil
}

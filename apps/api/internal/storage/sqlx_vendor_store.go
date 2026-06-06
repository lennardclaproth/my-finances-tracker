package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/my-finances-tracker/internal/vendor"
)

// SQLXVendorStore persists and reads vendor records.
type SQLXVendorStore struct {
	db        *DB
	tableName string
}

var (
	_ vendor.VendorCreator      = (*SQLXVendorStore)(nil)
	_ vendor.ActiveVendorLister = (*SQLXVendorStore)(nil)
	_ vendor.QueryStore         = (*SQLXVendorStore)(nil)
)

// NewSQLXVendorStore creates a vendor store backed by SQLX.
func NewSQLXVendorStore(db *DB) *SQLXVendorStore {
	return &SQLXVendorStore{
		db:        db,
		tableName: qualifyTable(db, SchemaVendors, TableVendors),
	}
}

// Create inserts a new vendor row, mapping a unique-constraint violation to
// vendor.ErrVendorAlreadyExists.
func (s *SQLXVendorStore) Create(ctx context.Context, v *vendor.Vendor) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, type, active, import_disabled, created_at, updated_at)
		VALUES (:id, :name, :type, :active, :import_disabled, :created_at, :updated_at)
	`, s.tableName)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, v); err != nil {
		if isUniqueViolation(err) {
			return vendor.ErrVendorAlreadyExists
		}
		return err
	}
	return nil
}

// GetByID returns one vendor by ID, or vendor.ErrVendorNotFound when absent.
func (s *SQLXVendorStore) GetByID(ctx context.Context, id uuid.UUID) (*vendor.Vendor, error) {
	var v vendor.Vendor
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT id, name, type, active, import_disabled, created_at, updated_at
		FROM %s WHERE id = ?
	`, s.tableName))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &v, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, vendor.ErrVendorNotFound
		}
		return nil, err
	}
	return &v, nil
}

// ListActive returns active vendors ordered by name.
func (s *SQLXVendorStore) ListActive(ctx context.Context) ([]*vendor.Vendor, error) {
	var vendors []*vendor.Vendor
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT id, name, type, active, import_disabled, created_at, updated_at
		FROM %s WHERE active = ? ORDER BY name ASC
	`, s.tableName))
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &vendors, query, true); err != nil {
		return nil, err
	}
	return vendors, nil
}

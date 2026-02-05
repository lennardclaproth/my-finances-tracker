package storage

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/domain"
	"github.com/lib/pq"
)

type SQLXVendorStore struct {
	db *DB
}

func NewSQLXVendorStore(db *DB) *SQLXVendorStore {
	return &SQLXVendorStore{db: db}
}

func (s *SQLXVendorStore) Create(vendor *domain.Vendor) error {
	record := fromDomainVendor(vendor)
	query := `
		INSERT INTO vendors (id, name, created_at, updated_at)
		VALUES (:id, :name, :created_at, :updated_at)
	`
	_, err := s.db.NamedExec(query, &record)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		// 23505 = unique_violation
		if pqErr.Code == "23505" {
			// optional: check constraint name if you have multiple uniques
			if pqErr.Constraint == "vendors_name_key" {
				return domain.ErrVendorAlreadyExists
			}
		}
	}
	return err
}

func (s *SQLXVendorStore) FetchByName(name domain.VendorID) (*domain.Vendor, error) {
	var record VendorRecord
	query := `SELECT id, name, created_at, updated_at FROM vendors WHERE name = $1`
	err := s.db.Get(&record, query, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrVendorNotFound
		}
		return nil, err
	}
	vendor := toDomainVendor(record)
	return &vendor, nil
}

func (s *SQLXVendorStore) FetchById(id uuid.UUID) (*domain.Vendor, error) {
	var record VendorRecord
	query := `SELECT id, name, created_at, updated_at FROM vendors WHERE id = $1`
	err := s.db.Get(&record, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrVendorNotFound
		}
		return nil, err
	}
	vendor := toDomainVendor(record)
	return &vendor, nil
}

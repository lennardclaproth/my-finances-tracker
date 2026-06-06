package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
)

// SQLXImporterStore persists and reads import records.
type SQLXImporterStore struct {
	db        *DB
	tableName string
}

var _ importer.ImportStore = (*SQLXImporterStore)(nil)

// NewSQLXImporterStore creates an import store backed by SQLX.
func NewSQLXImporterStore(db *DB) *SQLXImporterStore {
	return &SQLXImporterStore{
		db:        db,
		tableName: qualifyTable(db, SchemaImports, TableImports),
	}
}

// Create inserts a new import row. It persists the full Import record, including
// type, source, and listing_id (the previous store dropped these despite the
// struct and schema modelling them).
func (s *SQLXImporterStore) Create(ctx context.Context, imp *importer.Import) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, vendor_id, account_id, listing_id, type, source, path,
			status, status_msg, duplicates, total_rows, imported, failed,
			created_at, updated_at
		) VALUES (
			:id, :vendor_id, :account_id, :listing_id, :type, :source, :path,
			:status, :status_msg, :duplicates, :total_rows, :imported, :failed,
			:created_at, :updated_at
		)
	`, s.tableName)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, imp); err != nil {
		return err
	}
	return nil
}

// FetchByID returns one import by ID, or (nil, nil) when it does not exist so
// callers can treat a missing import as "nothing to do".
func (s *SQLXImporterStore) FetchByID(ctx context.Context, id uuid.UUID) (*importer.Import, error) {
	var imp importer.Import
	query := s.db.Rebind(fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, s.tableName))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &imp, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &imp, nil
}

// UpdateState persists the mutable processing fields of an import.
func (s *SQLXImporterStore) UpdateState(ctx context.Context, imp *importer.Import) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = :status, status_msg = :status_msg, duplicates = :duplicates,
			total_rows = :total_rows, imported = :imported, failed = :failed,
			updated_at = :updated_at
		WHERE id = :id
	`, s.tableName)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, imp); err != nil {
		return err
	}
	return nil
}

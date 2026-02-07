package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
)

type SQLXImportStore struct {
	db *DB
}

func NewSQLXImportStore(db *DB) *SQLXImportStore {
	return &SQLXImportStore{db: db}
}

func (s *SQLXImportStore) Create(ctx context.Context, imp *importer.Import) error {
	query := fmt.Sprintf(`INSERT INTO %s (id, vendor_id, path, status, status_msg, duplicates, total_rows, imported, failed, created_at, updated_at)
		VALUES (:id, :vendor_id, :path, :status, :status_msg, :duplicates, :total_rows, :imported, :failed, :created_at, :updated_at)
	`, TableImports)
	_, err := s.db.NamedExec(query, imp)
	return err
}

func (s *SQLXImportStore) OldestPending() (*importer.Import, error) {
	var imp importer.Import
	query := fmt.Sprintf(`
		SELECT id, vendor_id, path, status, status_msg, duplicates, total_rows, imported, failed, created_at, updated_at
		FROM %s
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT 1
	`, TableImports)
	err := s.db.Get(&imp, query, importer.ImportStatusPending)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, importer.ErrNoImportsPending
		}
		return nil, err
	}
	return &imp, nil
}

func (s *SQLXImportStore) UpdateState(ctx context.Context, imp *importer.Import) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = :status, status_msg = :status_msg, duplicates = :duplicates, total_rows = :total_rows, imported = :imported, failed = :failed, updated_at = :updated_at
		WHERE id = :id
	`, TableImports)
	_, err := s.db.NamedExec(query, imp)
	return err
}

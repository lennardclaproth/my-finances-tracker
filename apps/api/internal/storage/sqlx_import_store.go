package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
)

type SQLXImportStore struct {
	db        *DB
	tableName string
}

func NewSQLXImportStore(db *DB) *SQLXImportStore {
	return &SQLXImportStore{
		db:        db,
		tableName: qualifyTable(db, SchemaImports, TableImports),
	}
}

func (s *SQLXImportStore) Create(ctx context.Context, imp *importer.Import) error {
	query := fmt.Sprintf(`INSERT INTO %s (id, vendor_id, path, status, status_msg, duplicates, total_rows, imported, failed, created_at, updated_at)
		VALUES (:id, :vendor_id, :path, :status, :status_msg, :duplicates, :total_rows, :imported, :failed, :created_at, :updated_at)
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, imp)
	return err
}

func (s *SQLXImportStore) OldestPending(ctx context.Context) (*importer.Import, error) {
	var imp importer.Import
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT 1
	`, s.tableName)
	err := s.db.GetContext(ctx, &imp, query, importer.ImportStatusPending)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, importer.ErrNoImportsPending
		}
		return nil, err
	}
	return &imp, nil
}

func (s *SQLXImportStore) FetchByID(ctx context.Context, id uuid.UUID) (*importer.Import, error) {
	var imp importer.Import
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, s.tableName)
	err := s.db.GetContext(ctx, &imp, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, importer.ErrNoImportsPending
		}
		return nil, err
	}
	return &imp, nil
}

func (s *SQLXImportStore) ListPending(ctx context.Context, limit int) ([]*importer.Import, error) {
	var imports []*importer.Import
	if limit <= 0 {
		limit = 100
	}
	query := fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`, s.tableName)
	if err := s.db.SelectContext(ctx, &imports, query, importer.ImportStatusPending, limit); err != nil {
		return nil, err
	}
	return imports, nil
}

func (s *SQLXImportStore) TryMarkInProgress(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status = $4
	`, s.tableName)
	res, err := s.db.ExecContext(ctx, query, importer.ImportStatusInProgress, time.Now().UTC(), id, importer.ImportStatusPending)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (s *SQLXImportStore) UpdateState(ctx context.Context, imp *importer.Import) error {
	query := fmt.Sprintf(`
		UPDATE %s
		SET status = :status, status_msg = :status_msg, duplicates = :duplicates, total_rows = :total_rows, imported = :imported, failed = :failed, updated_at = :updated_at
		WHERE id = :id
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, imp)
	return err
}

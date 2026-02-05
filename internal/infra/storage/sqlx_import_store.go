package storage

import "github.com/lennardclaproth/my-finances-tracker/internal/domain"

type SQLXImportStore struct {
	db *DB
}

func NewSQLXImportStore(db *DB) *SQLXImportStore {
	return &SQLXImportStore{db: db}
}

func (s *SQLXImportStore) Create(imp *domain.Import) error {
	record := fromDomainImport(imp)
	query := `
		INSERT INTO imports (id, vendor_id, path, status, status_msg, duplicates, total_rows, imported, failed, created_at, updated_at)
		VALUES (:id, :vendor_id, :path, :status, :status_msg, :duplicates, :total_rows, :imported, :failed, :created_at, :updated_at)
	`
	_, err := s.db.NamedExec(query, &record)
	return err
}

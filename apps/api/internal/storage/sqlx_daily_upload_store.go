package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

type SQLXDailyUploadStore struct {
	db        *DB
	tableName string
}

func NewSQLXDailyUploadStore(db *DB) *SQLXDailyUploadStore {
	return &SQLXDailyUploadStore{
		db:        db,
		tableName: qualifyTable(db, SchemaMarketData, TableDailyUploads),
	}
}

func (s *SQLXDailyUploadStore) Create(ctx context.Context, upload *marketdata.DailyUpload) error {
	if upload.RowErrorsJSON == "" {
		if err := upload.SetRowErrors(upload.RowErrors, 0); err != nil {
			return err
		}
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id,
			listing_id,
			source,
			status,
			stored_filename,
			original_filename,
			status_message,
			total_rows,
			inserted_rows,
			duplicate_rows,
			error_rows,
			row_errors_json,
			started_at,
			finished_at,
			created_at,
			updated_at
		) VALUES (
			:id,
			:listing_id,
			:source,
			:status,
			:stored_filename,
			:original_filename,
			:status_message,
			:total_rows,
			:inserted_rows,
			:duplicate_rows,
			:error_rows,
			:row_errors_json,
			:started_at,
			:finished_at,
			:created_at,
			:updated_at
		)
	`, s.tableName)
	executor := s.db.GetExecutor(ctx)
	namedQuery, args, err := sqlx.Named(query, upload)
	if err != nil {
		return err
	}
	namedQuery = sqlx.Rebind(sqlx.BindType(s.db.DriverName()), namedQuery)
	_, err = executor.ExecContext(ctx, namedQuery, args...)
	return err
}

func (s *SQLXDailyUploadStore) FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.DailyUpload, error) {
	var upload marketdata.DailyUpload
	query := s.db.Rebind(fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, s.tableName))
	executor := s.db.GetExecutor(ctx)
	if err := sqlx.GetContext(ctx, executor, &upload, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, marketdata.ErrDailyUploadNotFound
		}
		return nil, err
	}
	if err := upload.DecodeRowErrors(); err != nil {
		return nil, err
	}
	return &upload, nil
}

func (s *SQLXDailyUploadStore) ListPending(ctx context.Context, limit int) ([]*marketdata.DailyUpload, error) {
	if limit <= 0 {
		limit = 100
	}
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE status = ?
		ORDER BY created_at ASC
		LIMIT ?
	`, s.tableName))
	executor := s.db.GetExecutor(ctx)

	rows := make([]*marketdata.DailyUpload, 0)
	if err := sqlx.SelectContext(ctx, executor, &rows, query, marketdata.DailyUploadStatusPending, limit); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		if err := row.DecodeRowErrors(); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

func (s *SQLXDailyUploadStore) TryMarkProcessing(ctx context.Context, id uuid.UUID) (bool, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s
		SET status = ?, started_at = ?, updated_at = ?
		WHERE id = ? AND status = ?
	`, s.tableName))
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, query, marketdata.DailyUploadStatusProcessing, now, now, id, marketdata.DailyUploadStatusPending)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (s *SQLXDailyUploadStore) UpdateState(ctx context.Context, upload *marketdata.DailyUpload) error {
	if upload.RowErrorsJSON == "" {
		if err := upload.SetRowErrors(upload.RowErrors, 0); err != nil {
			return err
		}
	}
	query := fmt.Sprintf(`
		UPDATE %s
		SET
			status = :status,
			status_message = :status_message,
			total_rows = :total_rows,
			inserted_rows = :inserted_rows,
			duplicate_rows = :duplicate_rows,
			error_rows = :error_rows,
			row_errors_json = :row_errors_json,
			started_at = :started_at,
			finished_at = :finished_at,
			updated_at = :updated_at
		WHERE id = :id
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, upload)
	return err
}

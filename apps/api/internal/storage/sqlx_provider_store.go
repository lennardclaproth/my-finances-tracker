package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

type SQLXProviderStore struct {
	db        *DB
	tableName string
}

func NewSQLXProviderStore(db *DB) *SQLXProviderStore {
	return &SQLXProviderStore{
		db:        db,
		tableName: qualifyTable(db, SchemaMarketData, TableProviders),
	}
}

func (s *SQLXProviderStore) Create(ctx context.Context, provider *marketdata.Provider) error {
	if err := provider.Validate(); err != nil {
		return err
	}
	if provider.ID == uuid.Nil {
		provider.ID = uuid.New()
	}

	existing, err := s.getExisting(ctx, provider)
	if err != nil {
		return err
	}
	if existing {
		return nil
	}

	query := fmt.Sprintf(`INSERT INTO %s (
			id,
			name,
			ingestion_mode,
			api_key,
			base_uri,
			remaining,
			used,
			total,
			resets_at
		)
		VALUES (
			:id,
			:name,
			:ingestion_mode,
			:api_key,
			:base_uri,
			:remaining,
			:used,
			:total,
			:resets_at
		)
	`, s.tableName)
	_, err = s.db.NamedExecContext(ctx, query, provider)
	return err
}

func (s *SQLXProviderStore) getExisting(ctx context.Context, provider *marketdata.Provider) (bool, error) {
	if provider == nil {
		return false, nil
	}
	var count int
	var query string
	var args []any
	if provider.IngestionMode == marketdata.ProviderIngestionModeManual {
		query = s.db.Rebind(fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE name = ? AND ingestion_mode = ?`, s.tableName))
		args = []any{provider.Name, provider.IngestionMode}
	} else {
		apiKey := ""
		if provider.ApiKey != nil {
			apiKey = *provider.ApiKey
		}
		query = s.db.Rebind(fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE name = ? AND ingestion_mode = ? AND api_key = ?`, s.tableName))
		args = []any{provider.Name, provider.IngestionMode, apiKey}
	}
	if err := s.db.GetContext(ctx, &count, query, args...); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *SQLXProviderStore) GetByName(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error) {
	var provider marketdata.Provider
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE name = ?
		  AND (
		    ingestion_mode = 'MANUAL'
		    OR ingestion_mode = 'API'
		  )
		ORDER BY
			CASE WHEN ingestion_mode = 'MANUAL' THEN 0 ELSE 1 END ASC,
			CASE WHEN total > 0 AND remaining <= 0 THEN 1 ELSE 0 END ASC,
			remaining DESC,
			used ASC,
			api_key ASC
		LIMIT 1
	`, s.tableName))
	err := s.db.GetContext(ctx, &provider, query, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, marketdata.ErrProviderNotFound
		}
		return nil, err
	}
	return &provider, nil
}

func (s *SQLXProviderStore) UpdateAPIKey(ctx context.Context, name marketdata.ProviderName, newAPIKey string) error {
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s
		SET api_key = ?
		WHERE name = ?
		  AND ingestion_mode = ?
		  AND api_key = (
			SELECT api_key
			FROM %s
			WHERE name = ?
			  AND ingestion_mode = ?
			ORDER BY used ASC, api_key ASC
			LIMIT 1
		  )
	`, s.tableName, s.tableName))
	res, err := s.db.ExecContext(ctx, query, newAPIKey, name, marketdata.ProviderIngestionModeAPI, name, marketdata.ProviderIngestionModeAPI)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return marketdata.ErrProviderNotFound
	}
	return nil
}

func (s *SQLXProviderStore) DeductTokens(ctx context.Context, name marketdata.ProviderName, count int32) error {
	if count <= 0 {
		return nil
	}

	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s
		SET
			used = used + ?,
			remaining = CASE
				WHEN total > 0 THEN
					CASE
						WHEN total - (used + ?) < 0 THEN 0
						ELSE total - (used + ?)
					END
				ELSE
					CASE
						WHEN remaining - ? < 0 THEN 0
						ELSE remaining - ?
					END
			END
		WHERE name = ?
		  AND ingestion_mode = ?
		  AND api_key = (
			SELECT api_key
			FROM %s
			WHERE name = ?
			  AND ingestion_mode = ?
			ORDER BY
				CASE WHEN total > 0 AND remaining <= 0 THEN 1 ELSE 0 END ASC,
				remaining DESC,
				used ASC,
				api_key ASC
			LIMIT 1
		  )
	`, s.tableName, s.tableName))
	res, err := s.db.ExecContext(ctx, query, count, count, count, count, count, name, marketdata.ProviderIngestionModeAPI, name, marketdata.ProviderIngestionModeAPI)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return marketdata.ErrProviderNotFound
	}
	return nil
}

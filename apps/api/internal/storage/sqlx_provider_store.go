package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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
	query := fmt.Sprintf(`INSERT INTO %s (
			name,
			api_key,
			base_uri,
			remaining,
			used,
			total,
			resets_at
		)
		VALUES (
			:name,
			:api_key,
			:base_uri,
			:remaining,
			:used,
			:total,
			:resets_at
		)
		ON CONFLICT (name, api_key) DO NOTHING
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, provider)
	return err
}

func (s *SQLXProviderStore) GetByName(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error) {
	var provider marketdata.Provider
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE name = ?
		ORDER BY
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
		  AND api_key = (
			SELECT api_key
			FROM %s
			WHERE name = ?
			ORDER BY used ASC, api_key ASC
			LIMIT 1
		  )
	`, s.tableName, s.tableName))
	res, err := s.db.ExecContext(ctx, query, newAPIKey, name, name)
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
		  AND api_key = (
			SELECT api_key
			FROM %s
			WHERE name = ?
			ORDER BY
				CASE WHEN total > 0 AND remaining <= 0 THEN 1 ELSE 0 END ASC,
				remaining DESC,
				used ASC,
				api_key ASC
			LIMIT 1
		  )
	`, s.tableName, s.tableName))
	res, err := s.db.ExecContext(ctx, query, count, count, count, count, count, name, name)
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

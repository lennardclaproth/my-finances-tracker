package storage

import (
	"context"
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
			base_uri
		)
		VALUES (
			:name,
			:api_key,
			:base_uri
		)
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, provider)
	return err
}

func (s *SQLXProviderStore) GetByName(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error) {
	var provider marketdata.Provider
	query := fmt.Sprintf(`SELECT * FROM %s WHERE name = $1`, s.tableName)
	err := s.db.GetContext(ctx, &provider, query, name)
	return &provider, err
}

func (s *SQLXProviderStore) UpdateAPIKey(ctx context.Context, name marketdata.ProviderName, newAPIKey string) error {
	query := fmt.Sprintf(`UPDATE %s SET api_key = $1 WHERE name = $2`, s.tableName)
	_, err := s.db.ExecContext(ctx, query, newAPIKey, name)
	return err
}

func (s *SQLXProviderStore) DeductTokens(ctx context.Context, name marketdata.ProviderName, count int32) error {
	// TODO: Implement quota tracking once provider quotas are persisted.
	_ = name
	_ = count
	return nil
}

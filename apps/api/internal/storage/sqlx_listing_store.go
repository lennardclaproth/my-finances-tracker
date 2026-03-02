package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
	"github.com/lib/pq"
)

type SQLXListingStore struct {
	db        *DB
	tableName string
}

func NewSQLXListingStore(db *DB) *SQLXListingStore {
	return &SQLXListingStore{
		db:        db,
		tableName: qualifyTable(db, SchemaMarketData, TableListings),
	}
}

var _ portfolio.ListingStore = (*SQLXListingStore)(nil)

func (s *SQLXListingStore) Create(ctx context.Context, listing *marketdata.Listing) error {
	query := fmt.Sprintf(`INSERT INTO %s (
			id, 
			symbol, 
			name, 
			source,
			isin,
			currency,
			region,
			type,
			ticker,
			description,
			exchange,
			should_accumulate
		)
		VALUES (
			:id, 
			:symbol, 
			:name,
			:source,
			:isin,
			:currency, 
			:region, 
			:type, 
			:ticker, 
			:description, 
			:exchange,
			:should_accumulate
		)
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, listing)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return marketdata.ErrListingAlreadyExists
		}
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "unique") && strings.Contains(msg, "symbol") && strings.Contains(msg, "source") {
			return marketdata.ErrListingAlreadyExists
		}
		return err
	}
	return nil
}

func (s *SQLXListingStore) FetchBySymbol(ctx context.Context, symbol string) (*marketdata.Listing, error) {
	var listing marketdata.Listing
	query := fmt.Sprintf(`SELECT * FROM %s WHERE symbol = $1`, s.tableName)
	err := s.db.GetContext(ctx, &listing, query, symbol)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &listing, nil
}

func (s *SQLXListingStore) List(ctx context.Context) ([]*marketdata.Listing, error) {
	listings := make([]*marketdata.Listing, 0)
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY symbol ASC, id ASC`, s.tableName)
	if err := s.db.SelectContext(ctx, &listings, query); err != nil {
		return nil, err
	}
	return listings, nil
}

func (s *SQLXListingStore) UpdateFields(ctx context.Context, listing *marketdata.Listing) error {
	listing.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf(`UPDATE %s
		SET name = :name, 
			isin = :isin,
			updated_at = :updated_at, 
			currency = :currency, 
			region = :region, 
			type = :type, 
			ticker = :ticker, 
			description = :description, 
			exchange = :exchange
		WHERE id = :id
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, listing)
	return err
}

func (s *SQLXListingStore) FetchByID(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
	var listing marketdata.Listing
	query := fmt.Sprintf(`SELECT * FROM %s WHERE id = $1`, s.tableName)
	err := s.db.GetContext(ctx, &listing, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &listing, nil
}

func (s *SQLXListingStore) FetchBySymbolOrISIN(ctx context.Context, val string) (*marketdata.Listing, error) {
	var listing marketdata.Listing
	query := fmt.Sprintf(`SELECT * FROM %s WHERE symbol = $1 OR isin = $1 LIMIT 1`, s.tableName)
	if err := s.db.GetContext(ctx, &listing, query, val); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &listing, nil
}

func (s *SQLXListingStore) TryAcquireSyncLock(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`
		UPDATE %s
		SET syncing = true, updated_at = $1
		WHERE id = $2
		  AND syncing = false
		  AND should_accumulate = true
	`, s.tableName)

	res, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

func (s *SQLXListingStore) ReleaseSyncLock(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`UPDATE %s 
		SET syncing = false, updated_at = $1 
		WHERE id = $2
	`, s.tableName)
	_, err := s.db.ExecContext(ctx, query, time.Now().UTC(), id)
	return err
}

func (s *SQLXListingStore) UpdateShouldAccumulate(ctx context.Context, id uuid.UUID, shouldAccumulate bool) error {
	query := fmt.Sprintf(`UPDATE %s SET should_accumulate = $1, updated_at = $2 WHERE id = $3`, s.tableName)
	_, err := s.db.ExecContext(ctx, query, shouldAccumulate, time.Now().UTC(), id)
	return err
}

func (s *SQLXListingStore) UpdateAccumulatedRange(ctx context.Context, id uuid.UUID, accumulatedStart, accumulatedEnd *time.Time) error {
	query := fmt.Sprintf(`UPDATE %s SET accumulated_start = $1, accumulated_end = $2, updated_at = $3 WHERE id = $4`, s.tableName)
	_, err := s.db.ExecContext(ctx, query, accumulatedStart, accumulatedEnd, time.Now().UTC(), id)
	return err
}

package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

type SQLXDailyStore struct {
	db        *DB
	tableName string
}

func NewSQLXDailyStore(db *DB) *SQLXDailyStore {
	return &SQLXDailyStore{
		db:        db,
		tableName: qualifyTable(db, SchemaMarketData, TableHistories),
	}
}

func (s *SQLXDailyStore) Create(ctx context.Context, daily *marketdata.Daily) error {
	if daily.ListingID == uuid.Nil {
		return marketdata.ErrDailyListingIDEmpty
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id,
			listing_id,
			symbol,
			date,
			open_cents,
			high_cents,
			low_cents,
			close_cents,
			volume,
			created_at,
			updated_at
		) VALUES (
			:id,
			:listing_id,
			:symbol,
			:date,
			:open,
			:high,
			:low,
			:close,
			:volume,
			:created_at,
			:updated_at
		)
		ON CONFLICT(listing_id, date) DO NOTHING
	`, s.tableName)
	_, err := s.db.NamedExecContext(ctx, query, daily)
	return err
}

func (s *SQLXDailyStore) FetchByListingID(ctx context.Context, listingID uuid.UUID, from, to *time.Time, limit, offset int) (*[]marketdata.Daily, error) {
	query := fmt.Sprintf(`
		SELECT
			id,
			listing_id,
			symbol,
			date,
			open_cents AS open,
			high_cents AS high,
			low_cents AS low,
			close_cents AS close,
			volume,
			created_at,
			updated_at
		FROM %s
		WHERE listing_id = ?
	`, s.tableName)

	args := []any{listingID}
	if from != nil {
		query += " AND date >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND date <= ?"
		args = append(args, *to)
	}
	query += " ORDER BY date ASC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	query = s.db.Rebind(strings.TrimSpace(query))
	var dailies []marketdata.Daily
	if err := s.db.SelectContext(ctx, &dailies, query, args...); err != nil {
		return nil, err
	}
	return &dailies, nil
}

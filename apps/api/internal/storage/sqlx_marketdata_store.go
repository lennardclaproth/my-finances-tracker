package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata/marketstack"
)

// SQLXMarketDataStore persists and reads market-data listings, EOD datapoints, and
// providers. It satisfies the marketdata command/query/sync contracts as well as
// the provider lookup the marketstack client and bootstrap depend on.
type SQLXMarketDataStore struct {
	db             *DB
	providersTable string
	listingsTable  string
	dailiesTable   string
}

var (
	_ marketdata.CommandStore   = (*SQLXMarketDataStore)(nil)
	_ marketdata.QueryStore     = (*SQLXMarketDataStore)(nil)
	_ marketdata.SyncStore      = (*SQLXMarketDataStore)(nil)
	_ marketstack.ProviderStore = (*SQLXMarketDataStore)(nil)
)

// NewSQLXMarketDataStore creates a market-data store backed by SQLX.
func NewSQLXMarketDataStore(db *DB) *SQLXMarketDataStore {
	return &SQLXMarketDataStore{
		db:             db,
		providersTable: qualifyTable(db, SchemaMarketData, TableProviders),
		listingsTable:  qualifyTable(db, SchemaMarketData, TableListings),
		dailiesTable:   qualifyTable(db, SchemaMarketData, TableHistories),
	}
}

// --- Listings ---------------------------------------------------------------

// Create inserts a new listing, mapping a unique-constraint violation (symbol/source)
// to marketdata.ErrListingAlreadyExists.
func (s *SQLXMarketDataStore) Create(ctx context.Context, listing *marketdata.Listing) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, symbol, name, source, isin, currency, region, type,
			ticker, description, exchange, should_accumulate
		) VALUES (
			:id, :symbol, :name, :source, :isin, :currency, :region, :type,
			:ticker, :description, :exchange, :should_accumulate
		)
	`, s.listingsTable)
	if _, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, listing); err != nil {
		if isUniqueViolation(err) {
			return marketdata.ErrListingAlreadyExists
		}
		return err
	}
	return nil
}

// Update persists the editable metadata fields of a listing.
func (s *SQLXMarketDataStore) Update(ctx context.Context, listing *marketdata.Listing) error {
	listing.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf(`
		UPDATE %s
		SET name = :name, isin = :isin, updated_at = :updated_at, currency = :currency,
			region = :region, type = :type, ticker = :ticker, description = :description,
			exchange = :exchange
		WHERE id = :id
	`, s.listingsTable)
	_, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, listing)
	return err
}

// Get returns one listing by ID, or (nil, nil) when it does not exist.
func (s *SQLXMarketDataStore) Get(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
	return s.selectListing(ctx, fmt.Sprintf(`SELECT * FROM %s WHERE id = ?`, s.listingsTable), id)
}

// Listing returns one listing by ID, or marketdata.ErrListingNotFound when absent.
// It is the sync path's lookup, which requires a present listing.
func (s *SQLXMarketDataStore) Listing(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error) {
	listing, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if listing == nil {
		return nil, marketdata.ErrListingNotFound
	}
	return listing, nil
}

// GetBySymbol returns one listing by symbol, or (nil, nil) when it does not exist.
func (s *SQLXMarketDataStore) GetBySymbol(ctx context.Context, symbol string) (*marketdata.Listing, error) {
	return s.selectListing(ctx, fmt.Sprintf(`SELECT * FROM %s WHERE symbol = ?`, s.listingsTable), symbol)
}

func (s *SQLXMarketDataStore) selectListing(ctx context.Context, query string, arg any) (*marketdata.Listing, error) {
	var listing marketdata.Listing
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &listing, s.db.Rebind(query), arg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &listing, nil
}

// List returns listings ordered by symbol, optionally paginated.
func (s *SQLXMarketDataStore) List(ctx context.Context, limit, offset *int) ([]*marketdata.Listing, error) {
	query := fmt.Sprintf(`SELECT * FROM %s ORDER BY symbol ASC, id ASC`, s.listingsTable)
	args := []any{}
	if limit != nil {
		query += " LIMIT ?"
		args = append(args, *limit)
		if offset != nil {
			query += " OFFSET ?"
			args = append(args, *offset)
		}
	}
	listings := make([]*marketdata.Listing, 0)
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &listings, s.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	return listings, nil
}

// Search returns listings matching q (symbol/name/ISIN) and the total match count.
func (s *SQLXMarketDataStore) Search(ctx context.Context, q string, limit, offset *int) ([]*marketdata.Listing, int, error) {
	lim := 25
	if limit != nil {
		lim = *limit
	}
	if lim <= 0 {
		lim = 25
	}
	if lim > 100 {
		lim = 100
	}
	off := 0
	if offset != nil && *offset > 0 {
		off = *offset
	}

	pattern := "%" + strings.ToLower(strings.TrimSpace(q)) + "%"
	where := `WHERE LOWER(symbol) LIKE ? OR LOWER(name) LIKE ? OR LOWER(COALESCE(isin, '')) LIKE ?`
	executor := s.db.GetExecutor(ctx)

	countQuery := s.db.Rebind(fmt.Sprintf(`SELECT COUNT(1) FROM %s %s`, s.listingsTable, where))
	var total int
	if err := sqlx.GetContext(ctx, executor, &total, countQuery, pattern, pattern, pattern); err != nil {
		return nil, 0, err
	}

	dataQuery := s.db.Rebind(fmt.Sprintf(`
		SELECT * FROM %s
		%s
		ORDER BY symbol ASC, id ASC
		LIMIT ? OFFSET ?
	`, s.listingsTable, where))
	listings := make([]*marketdata.Listing, 0)
	if err := sqlx.SelectContext(ctx, executor, &listings, dataQuery, pattern, pattern, pattern, lim, off); err != nil {
		return nil, 0, err
	}
	return listings, total, nil
}

// ShouldAccumulate sets the listing's should_accumulate flag.
func (s *SQLXMarketDataStore) ShouldAccumulate(ctx context.Context, lsID uuid.UUID, val bool) error {
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET should_accumulate = ?, updated_at = ? WHERE id = ?`, s.listingsTable))
	_, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, val, time.Now().UTC(), lsID)
	return err
}

// SetAccumulatedRange records the accumulated EOD window and clears the accumulate flag.
func (s *SQLXMarketDataStore) SetAccumulatedRange(ctx context.Context, id uuid.UUID, from, to *time.Time) error {
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s
		SET accumulated_start = ?, accumulated_end = ?, should_accumulate = ?, updated_at = ?
		WHERE id = ?
	`, s.listingsTable))
	_, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, from, to, false, time.Now().UTC(), id)
	return err
}

// TryAcquireSyncLock atomically claims the sync lock for an accumulating listing,
// returning true when the lock was acquired.
func (s *SQLXMarketDataStore) TryAcquireSyncLock(ctx context.Context, id uuid.UUID) (bool, error) {
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s
		SET syncing = ?, updated_at = ?
		WHERE id = ? AND syncing = ? AND should_accumulate = ?
	`, s.listingsTable))
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, true, time.Now().UTC(), id, false, true)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// ReleaseSyncLock clears the sync lock for a listing.
func (s *SQLXMarketDataStore) ReleaseSyncLock(ctx context.Context, id uuid.UUID) error {
	query := s.db.Rebind(fmt.Sprintf(`UPDATE %s SET syncing = ?, updated_at = ? WHERE id = ?`, s.listingsTable))
	_, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, false, time.Now().UTC(), id)
	return err
}

// --- EOD datapoints ---------------------------------------------------------

const eodInsertColumns = `
	id, listing_id, symbol, date,
	open_cents, high_cents, low_cents, close_cents,
	volume, created_at, updated_at
`

const eodInsertValues = `
	:id, :listing_id, :symbol, :date,
	:open, :high, :low, :close,
	:volume, :created_at, :updated_at
`

// CreateEODs persists a batch of EOD datapoints in a single insert, skipping rows
// that already exist for the (listing, date) pair, and returns the number inserted.
func (s *SQLXMarketDataStore) CreateEODs(ctx context.Context, eods []*marketdata.EOD) (int, error) {
	if len(eods) == 0 {
		return 0, nil
	}
	for _, eod := range eods {
		if eod.ListingID == uuid.Nil {
			return 0, marketdata.ErrDailyListingIDEmpty
		}
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (listing_id, date) DO NOTHING
	`, s.dailiesTable, eodInsertColumns, eodInsertValues)
	res, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, eods)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}

// InsertEOD persists one EOD datapoint, ignoring an existing (listing, date) pair.
func (s *SQLXMarketDataStore) InsertEOD(ctx context.Context, eod *marketdata.EOD) error {
	if eod.ListingID == uuid.Nil {
		return marketdata.ErrDailyListingIDEmpty
	}
	query := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT (listing_id, date) DO NOTHING
	`, s.dailiesTable, eodInsertColumns, eodInsertValues)
	_, err := sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, eod)
	return err
}

// CountEODByListing counts EOD datapoints for a listing within an optional date range.
func (s *SQLXMarketDataStore) CountEODByListing(ctx context.Context, lsID uuid.UUID, from, to *time.Time) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE listing_id = ?`, s.dailiesTable)
	args := []any{lsID}
	if from != nil {
		query += " AND date >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND date <= ?"
		args = append(args, *to)
	}
	var total int
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &total, s.db.Rebind(query), args...); err != nil {
		return 0, err
	}
	return total, nil
}

// GetEODForListing returns EOD datapoints for a listing within an optional date range,
// ordered by date (sort "asc"/"desc"), optionally paginated.
func (s *SQLXMarketDataStore) GetEODForListing(ctx context.Context, lsID uuid.UUID, from, to *time.Time, limit, offset *int, sort string) ([]*marketdata.EOD, error) {
	query := fmt.Sprintf(`
		SELECT
			id, listing_id, symbol, date,
			open_cents AS open, high_cents AS high, low_cents AS low, close_cents AS close,
			volume, created_at, updated_at
		FROM %s
		WHERE listing_id = ?
	`, s.dailiesTable)
	args := []any{lsID}
	if from != nil {
		query += " AND date >= ?"
		args = append(args, *from)
	}
	if to != nil {
		query += " AND date <= ?"
		args = append(args, *to)
	}
	query += fmt.Sprintf(" ORDER BY date %s", normalizeSortOrder(sort))
	if limit != nil && *limit > 0 {
		query += " LIMIT ?"
		args = append(args, *limit)
		if offset != nil && *offset > 0 {
			query += " OFFSET ?"
			args = append(args, *offset)
		}
	}

	eods := make([]*marketdata.EOD, 0)
	if err := sqlx.SelectContext(ctx, s.db.GetExecutor(ctx), &eods, s.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	return eods, nil
}

// --- Providers --------------------------------------------------------------

// CreateProvider inserts a provider unless an equivalent row already exists.
func (s *SQLXMarketDataStore) CreateProvider(ctx context.Context, provider *marketdata.Provider) error {
	if err := provider.Validate(); err != nil {
		return err
	}
	if provider.ID == uuid.Nil {
		provider.ID = uuid.New()
	}

	exists, err := s.providerExists(ctx, provider)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, name, ingestion_mode, api_key, base_uri, remaining, used, total, resets_at
		) VALUES (
			:id, :name, :ingestion_mode, :api_key, :base_uri, :remaining, :used, :total, :resets_at
		)
	`, s.providersTable)
	_, err = sqlx.NamedExecContext(ctx, s.db.GetExecutor(ctx), query, provider)
	return err
}

func (s *SQLXMarketDataStore) providerExists(ctx context.Context, provider *marketdata.Provider) (bool, error) {
	if provider == nil {
		return false, nil
	}
	var (
		query string
		args  []any
	)
	if provider.IngestionMode == marketdata.ProviderIngestionModeManual {
		query = fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE name = ? AND ingestion_mode = ?`, s.providersTable)
		args = []any{provider.Name, provider.IngestionMode}
	} else {
		apiKey := ""
		if provider.ApiKey != nil {
			apiKey = *provider.ApiKey
		}
		query = fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE name = ? AND ingestion_mode = ? AND api_key = ?`, s.providersTable)
		args = []any{provider.Name, provider.IngestionMode, apiKey}
	}
	var count int
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &count, s.db.Rebind(query), args...); err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetProviderByName returns the best provider candidate for the given name, or
// marketdata.ErrProviderNotFound when none exists.
func (s *SQLXMarketDataStore) GetProviderByName(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error) {
	var provider marketdata.Provider
	query := s.db.Rebind(fmt.Sprintf(`
		SELECT *
		FROM %s
		WHERE name = ? AND (ingestion_mode = 'MANUAL' OR ingestion_mode = 'API')
		ORDER BY
			CASE WHEN ingestion_mode = 'MANUAL' THEN 0 ELSE 1 END ASC,
			CASE WHEN total > 0 AND remaining <= 0 THEN 1 ELSE 0 END ASC,
			remaining DESC,
			used ASC,
			api_key ASC
		LIMIT 1
	`, s.providersTable))
	if err := sqlx.GetContext(ctx, s.db.GetExecutor(ctx), &provider, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, marketdata.ErrProviderNotFound
		}
		return nil, err
	}
	return &provider, nil
}

// GetByName returns the best provider candidate for the given name. It satisfies the
// marketstack client's provider lookup and is an alias of GetProviderByName.
func (s *SQLXMarketDataStore) GetByName(ctx context.Context, name marketdata.ProviderName) (*marketdata.Provider, error) {
	return s.GetProviderByName(ctx, name)
}

// DeductTokens decrements the quota counters of the active API provider key after usage.
func (s *SQLXMarketDataStore) DeductTokens(ctx context.Context, name marketdata.ProviderName, count int32) error {
	if count <= 0 {
		return nil
	}
	query := s.db.Rebind(fmt.Sprintf(`
		UPDATE %s
		SET
			used = used + ?,
			remaining = CASE
				WHEN total > 0 THEN
					CASE WHEN total - (used + ?) < 0 THEN 0 ELSE total - (used + ?) END
				ELSE
					CASE WHEN remaining - ? < 0 THEN 0 ELSE remaining - ? END
			END
		WHERE name = ?
		  AND ingestion_mode = ?
		  AND api_key = (
			SELECT api_key
			FROM %s
			WHERE name = ? AND ingestion_mode = ?
			ORDER BY
				CASE WHEN total > 0 AND remaining <= 0 THEN 1 ELSE 0 END ASC,
				remaining DESC,
				used ASC,
				api_key ASC
			LIMIT 1
		  )
	`, s.providersTable, s.providersTable))
	res, err := s.db.GetExecutor(ctx).ExecContext(ctx, query, count, count, count, count, count, name, marketdata.ProviderIngestionModeAPI, name, marketdata.ProviderIngestionModeAPI)
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

package marketdata

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
	"github.com/lennardclaproth/my-finances-tracker/internal/observability"
	"go.elastic.co/apm/v2"
)

type listingStore interface {
	Create(ctx context.Context, listing *Listing) error
	UpdateFields(ctx context.Context, listing *Listing) error
	List(ctx context.Context) ([]*Listing, error)
	Search(ctx context.Context, q string, limit, offset int) ([]*Listing, int, error)
	FetchBySymbol(ctx context.Context, symbol string) (*Listing, error)
	FetchByID(ctx context.Context, id uuid.UUID) (*Listing, error)
	TryAcquireSyncLock(ctx context.Context, id uuid.UUID) (bool, error)
	ReleaseSyncLock(ctx context.Context, id uuid.UUID) error
	UpdateShouldAccumulate(ctx context.Context, id uuid.UUID, shouldAccumulate bool) error
	UpdateAccumulatedRange(ctx context.Context, id uuid.UUID, accumulatedStart, accumulatedEnd *time.Time) error
}

type dailyStore interface {
	Create(ctx context.Context, daily *Daily) error
	FetchByListingID(ctx context.Context, listingID uuid.UUID, from, to *time.Time, limit, offset int) (*[]Daily, error)
	FetchByListingIDWithSort(
		ctx context.Context,
		listingID uuid.UUID,
		from, to *time.Time,
		limit, offset int,
		sortOrder DailyDateSortOrder,
	) (*[]Daily, error)
	CountByListingID(ctx context.Context, listingID uuid.UUID, from, to *time.Time) (int, error)
}

type providerStore interface {
	GetByName(ctx context.Context, name ProviderName) (*Provider, error)
}

type Service struct {
	listingStore  listingStore
	dailyStore    dailyStore
	providerStore providerStore
	client        eodClient
	log           logging.Logger
}

type eodClient interface {
	GetEOD(ctx context.Context, symbols []string, from, to *time.Time) iter.Seq2[Daily, error]
}

type Metadata struct {
	Message     string
	ResultCount int
	TotalCount  int
}

type DailyResponse struct {
	Data     []Daily
	Metadata Metadata
}

func NewService(listingStore listingStore, dailyStore dailyStore, client eodClient, log logging.Logger, providers ...providerStore) *Service {
	var providerStore providerStore
	if len(providers) > 0 {
		providerStore = providers[0]
	}
	return &Service{
		listingStore:  listingStore,
		dailyStore:    dailyStore,
		providerStore: providerStore,
		client:        client,
		log:           log,
	}
}

func (s *Service) sourceIsManual(ctx context.Context, source Source) (bool, error) {
	if s.providerStore == nil {
		return false, nil
	}
	providerName, err := ProviderNameFromSource(source)
	if err != nil {
		return false, nil
	}
	provider, err := s.providerStore.GetByName(ctx, providerName)
	if err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			return false, nil
		}
		return false, err
	}
	if provider == nil {
		return false, nil
	}
	return provider.IsManualIngestion(), nil
}

// GetDailies fetches daily data for a given symbol and date range.
// returns the data along with metadata about the request.
func (s *Service) GetDailies(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*DailyResponse, error) {
	return s.GetDailiesWithSort(ctx, symbol, from, to, limit, offset, DailyDateSortAsc)
}

func (s *Service) GetDailiesWithSort(
	ctx context.Context,
	symbol string,
	from, to *time.Time,
	limit, offset int,
	sortOrder DailyDateSortOrder,
) (*DailyResponse, error) {
	// Fetch listing to ensure it exists and is accumulated
	listing, err := s.listingStore.FetchBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("GetDaily failed to fetch listing: %w", err)
	}
	return s.getDailiesForListing(ctx, listing, from, to, limit, offset, sortOrder)
}

// GetDailiesByListingID fetches daily data for a listing id and date range.
func (s *Service) GetDailiesByListingID(
	ctx context.Context,
	listingID uuid.UUID,
	from, to *time.Time,
	limit, offset int,
) (*DailyResponse, error) {
	return s.GetDailiesByListingIDWithSort(ctx, listingID, from, to, limit, offset, DailyDateSortAsc)
}

func (s *Service) GetDailiesByListingIDWithSort(
	ctx context.Context,
	listingID uuid.UUID,
	from, to *time.Time,
	limit, offset int,
	sortOrder DailyDateSortOrder,
) (*DailyResponse, error) {
	listing, err := s.listingStore.FetchByID(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("GetDaily failed to fetch listing by id: %w", err)
	}
	return s.getDailiesForListing(ctx, listing, from, to, limit, offset, sortOrder)
}

func (s *Service) getDailiesForListing(
	ctx context.Context,
	listing *Listing,
	from, to *time.Time,
	limit, offset int,
	sortOrder DailyDateSortOrder,
) (*DailyResponse, error) {
	if listing == nil {
		return nil, fmt.Errorf("GetDaily failed, listing not found")
	}
	if !listing.Active {
		return nil, fmt.Errorf("GetDaily failed, listing with symbol %s is not active", listing.Symbol)
	}

	isManualProvider, err := s.sourceIsManual(ctx, listing.Source)
	if err != nil {
		s.log.Error(ctx, "GetDaily failed to resolve provider ingestion mode", err, "symbol", listing.Symbol, "source", listing.Source)
	}
	if isManualProvider {
		data, err := s.dailyStore.FetchByListingIDWithSort(
			ctx,
			listing.ID,
			from,
			to,
			limit,
			offset,
			NormalizeDailyDateSortOrder(string(sortOrder)),
		)
		if err != nil {
			return nil, fmt.Errorf("GetDaily failed to fetch daily data: %w", err)
		}
		totalCount, err := s.dailyStore.CountByListingID(ctx, listing.ID, from, to)
		if err != nil {
			return nil, fmt.Errorf("GetDaily failed to fetch total daily count: %w", err)
		}
		return &DailyResponse{
			Data: *data,
			Metadata: Metadata{
				Message:     "Manual provider configured; automatic sync disabled",
				ResultCount: len(*data),
				TotalCount:  totalCount,
			},
		}, nil
	}

	// If the listing is currently syncing it means that there is already a sync in progress.
	// In this case we can just fetch the existing data from the database, we don't need to trigger another sync.
	if listing.Syncing {
		data, err := s.dailyStore.FetchByListingIDWithSort(
			ctx,
			listing.ID,
			from,
			to,
			limit,
			offset,
			NormalizeDailyDateSortOrder(string(sortOrder)),
		)
		if err != nil {
			return nil, fmt.Errorf("GetDaily failed to fetch daily data: %w", err)
		}
		totalCount, err := s.dailyStore.CountByListingID(ctx, listing.ID, from, to)
		if err != nil {
			return nil, fmt.Errorf("GetDaily failed to fetch total daily count: %w", err)
		}
		return &DailyResponse{
			Data: *data,
			Metadata: Metadata{
				Message:     "Data may be stale, listing is currently syncing",
				ResultCount: len(*data),
				TotalCount:  totalCount,
			},
		}, nil
	}
	// If the accumulated end date is nil or before the day before the current data, it means the data is not
	// up to date, so we should trigger a sync to fetch the latest data.
	targetEnd := date.LatestBusinessDate(time.Now(), time.Local)
	if listing.AccumulatedEnd == nil || date.DateOnly(*listing.AccumulatedEnd, time.Local).Before(targetEnd) {
		err := s.listingStore.UpdateShouldAccumulate(ctx, listing.ID, true)
		if err != nil {
			s.log.Error(ctx, "GetDaily failed to update listing should_accumulate flag", err, "listing_id", listing.ID, "symbol", listing.Symbol)
		}
		if err == nil {
			s.startAsyncDailySync(ctx, listing, "GetDaily failed to sync daily data for listing")
		}
	}
	// Fetch daily data from the database, this will return the existing data if a sync is in progress, or the up to date data if not.
	data, err := s.dailyStore.FetchByListingIDWithSort(
		ctx,
		listing.ID,
		from,
		to,
		limit,
		offset,
		NormalizeDailyDateSortOrder(string(sortOrder)),
	)
	if err != nil {
		return nil, fmt.Errorf("GetDaily failed to fetch daily data: %w", err)
	}
	totalCount, err := s.dailyStore.CountByListingID(ctx, listing.ID, from, to)
	if err != nil {
		return nil, fmt.Errorf("GetDaily failed to fetch total daily count: %w", err)
	}
	return &DailyResponse{
		Data: *data,
		Metadata: Metadata{
			Message:     "Sync triggered; data may be stale until sync is complete",
			ResultCount: len(*data),
			TotalCount:  totalCount,
		},
	}, nil
}

// syncDailyData fetches daily data for a given listing and date range from the client and persists it to the database.
func (s *Service) syncDailyData(ctx context.Context, listingID uuid.UUID, from, to *time.Time) error {
	// Acquire sync lock, this ensures that only one sync can be in progress.
	// The acquiring of the lock is an atomic operation, returning an error if the lock is
	// already held, which means another sync is in progress.
	acquired, err := s.listingStore.TryAcquireSyncLock(ctx, listingID)
	if err != nil {
		return fmt.Errorf("syncDailyData failed to acquire sync lock: %w", err)
	}
	if !acquired {
		return ErrSyncInProgress
	}
	defer func() {
		// Ensure we always clear the syncing flag, even if the request context is canceled.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.listingStore.ReleaseSyncLock(cleanupCtx, listingID); err != nil {
			s.log.Error(cleanupCtx, "syncDailyData failed to release sync lock", err, "listing_id", listingID)
		}
	}()
	// Get most up to date listing data
	listing, err := s.listingStore.FetchByID(ctx, listingID)
	if err != nil {
		return fmt.Errorf("syncDailyData failed to fetch listing: %w", err)
	}
	// If from is nil it means there is no override on the accumulation date
	// so we should continue accumulating from the last accumulated date, which is the accumulated end date.
	if from == nil {
		// Accumulated end defaults to nil so we can safely pass this
		// it is important to note that if accumulated end is nil, the MarketStackClient will not include the date_from parameter in the request, which will cause MarketStack to return all available data for the symbol.
		// This is desired.
		from = listing.AccumulatedEnd
	}
	// If to is nil, accumulate only through the latest completed business date (exclude today).
	if to == nil {
		latestCompletedBusinessDate := date.LatestBusinessDate(time.Now(), time.Local)
		to = &latestCompletedBusinessDate
	}
	for d, err := range s.client.GetEOD(ctx, []string{listing.Symbol}, from, to) {
		if err != nil {
			s.log.Error(ctx, "syncDailyData failed to fetch daily data", err, "listing_id", listing.ID, "symbol", listing.Symbol)
			continue
		}
		d.ListingID = listing.ID
		if err := s.dailyStore.Create(ctx, &d); err != nil {
			s.log.Error(ctx, "syncDailyData failed to persist daily data", err, "listing_id", listing.ID, "symbol", listing.Symbol, "date", d.Date)
			continue
		}
	}
	if err := s.listingStore.UpdateAccumulatedRange(ctx, listingID, from, to); err != nil {
		s.log.Warn(ctx, "syncDailyData failed to update accumulated range", "listing_id", listingID.String(), "error", err.Error())
	}
	return nil
}

// CreateListing creates a new listing and persists it.
func (s *Service) CreateListing(
	ctx context.Context,
	symbol, name string,
	source Source,
	options ...ListingOption,
) (*Listing, error) {
	// Check if listing with source and symbol already exists
	existing, err := s.listingStore.FetchBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("CreateListing failed to fetch listing: %w", err)
	}
	if existing != nil && existing.Source == source {
		return nil, ErrListingAlreadyExists
	}
	// Create new listing and persist
	listing, err := NewListing(symbol, name, source, options...)
	if err != nil {
		return nil, fmt.Errorf("CreateListing failed to create listing: %w", err)
	}
	if err := s.listingStore.Create(ctx, listing); err != nil {
		if errors.Is(err, ErrListingAlreadyExists) {
			return nil, ErrListingAlreadyExists
		}
		return nil, fmt.Errorf("CreateListing failed to persist listing: %w", err)
	}

	isManualProvider, err := s.sourceIsManual(ctx, source)
	if err != nil {
		s.log.Error(ctx, "CreateListing failed to resolve provider ingestion mode", err, "symbol", symbol, "source", source)
	}
	if isManualProvider {
		return listing, nil
	}

	s.startAsyncDailySync(ctx, listing, "CreateListing failed to sync daily data for listing")
	return listing, nil
}

func (s *Service) startAsyncDailySync(ctx context.Context, listing *Listing, errorMessage string) {
	if listing == nil {
		return
	}
	headers := observability.PropagationHeadersFromContext(ctx)
	listingID := listing.ID
	symbol := listing.Symbol
	source := listing.Source

	go func() {
		fnCtx := observability.ContextWithPropagationHeaders(context.Background(), headers)
		fnCtx, _, _ = observability.EnsureRequestAndCorrelationIDs(
			fnCtx,
			observability.RequestIDFromContext(fnCtx),
			observability.CorrelationIDFromContext(fnCtx),
		)
		fnCtx, cancel := context.WithTimeout(fnCtx, 10*time.Minute)
		defer cancel()

		operation := observability.JobOperation("marketdata_sync_daily")
		tx, txCtx, traceErr := observability.StartTransactionFromHeaders(
			fnCtx,
			operation,
			"job",
			headers,
		)
		if traceErr != nil {
			s.log.Error(txCtx, "failed to parse trace context for async marketdata sync", traceErr,
				"listing_id", listingID,
				"symbol", symbol,
				"source", source,
				"operation", operation,
				"component", "service",
				"outcome", "failure",
				"error_class", "internal",
			)
		}

		tx.Result = "success"
		tx.Outcome = "success"
		observability.SetSafeTransactionLabels(tx, map[string]any{
			"operation":  operation,
			"component":  "service",
			"listing_id": listingID.String(),
			"symbol":     symbol,
			"source":     source,
			"stage":      "sync_daily_data",
		})

		var syncErr error
		defer func() {
			if syncErr != nil {
				tx.Result = "error"
				tx.Outcome = "failure"
				apm.CaptureError(txCtx, syncErr).Send()
			}
			tx.End()
		}()

		span, spanCtx := apm.StartSpan(txCtx, "sync_daily_data", "service")
		syncErr = s.syncDailyData(spanCtx, listingID, nil, nil)
		span.End()
		if syncErr != nil {
			s.log.Error(spanCtx, errorMessage, syncErr,
				"listing_id", listingID,
				"symbol", symbol,
				"source", source,
				"operation", operation,
				"component", "service",
				"outcome", "failure",
			)
		}
	}()
}

// UpdateListingFields updates only the provided listing fields.
func (s *Service) UpdateListingFields(
	ctx context.Context,
	id uuid.UUID,
	description, exchange, region, currency, isin, ticker, typ *string,
) (*Listing, error) {
	if description == nil && exchange == nil && region == nil && currency == nil && isin == nil && ticker == nil && typ == nil {
		return nil, ErrNoListingFieldsToUpdate
	}
	listing, err := s.listingStore.FetchByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("UpdateListingFields failed to fetch listing: %w", err)
	}
	if listing == nil {
		return nil, ErrListingNotFound
	}

	if description != nil {
		listing.Description = description
	}
	if exchange != nil {
		listing.Exchange = exchange
	}
	if region != nil {
		listing.Region = region
	}
	if isin != nil {
		listing.ISIN = isin
	}
	if ticker != nil {
		listing.Ticker = ticker
	}
	if typ != nil {
		listing.Type = typ
	}
	if currency != nil {
		cur := money.Currency(*currency)
		if !cur.IsValid() {
			return nil, ErrInvalidListingCurrency
		}
		listing.Currency = &cur
	}

	if err := s.listingStore.UpdateFields(ctx, listing); err != nil {
		return nil, fmt.Errorf("UpdateListingFields failed to persist listing: %w", err)
	}
	return listing, nil
}

// ListListings returns all listings in deterministic order for UI presentation.
func (s *Service) ListListings(ctx context.Context) ([]*Listing, error) {
	listings, err := s.listingStore.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("ListListings failed to fetch listings: %w", err)
	}
	return listings, nil
}

func (s *Service) SearchListings(ctx context.Context, q string, limit, offset int) ([]*Listing, int, error) {
	listings, total, err := s.listingStore.Search(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("SearchListings failed to fetch listings: %w", err)
	}
	return listings, total, nil
}

package marketdata

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

type listingStore interface {
	Create(ctx context.Context, listing *Listing) error
	UpdateFields(ctx context.Context, listing *Listing) error
	FetchBySymbol(ctx context.Context, symbol string) (*Listing, error)
	FetchByID(ctx context.Context, id uuid.UUID) (*Listing, error)
	TryAcquireSyncLock(ctx context.Context, id uuid.UUID) (bool, error)
	ReleaseSyncLock(ctx context.Context, id uuid.UUID) error
	UpdateShouldAccumulate(ctx context.Context, id uuid.UUID, shouldAccumulate bool) error
	UpdateAccumulatedRange(ctx context.Context, id uuid.UUID, accumulatedStart, accumulatedEnd *time.Time) error
}

type dailyStore interface {
	Create(ctx context.Context, daily *Daily) error
	FetchByListingID(ctx context.Context, listingID string, from, to *time.Time, limit, offset int) (*[]Daily, error)
}

type Service struct {
	listingStore listingStore
	dailyStore   dailyStore
	client       eodClient
	log          logging.Logger
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

func NewService(listingStore listingStore, dailyStore dailyStore, client eodClient, log logging.Logger) *Service {
	return &Service{
		listingStore: listingStore,
		dailyStore:   dailyStore,
		client:       client,
		log:          log,
	}
}

func latestBusinessDate(now time.Time, loc *time.Location) time.Time {
	n := now.In(loc)
	// Target is yesterday (one-day lag).
	target := n.AddDate(0, 0, -1)
	// If target falls on weekend, roll back to Friday.
	switch target.Weekday() {
	case time.Saturday:
		target = target.AddDate(0, 0, -1)
	case time.Sunday:
		target = target.AddDate(0, 0, -2)
	}
	// Normalize to midnight local time (date-only semantics).
	y, m, d := target.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

func dateOnly(t time.Time, loc *time.Location) time.Time {
	tt := t.In(loc)
	y, m, d := tt.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, loc)
}

// GetDailies fetches daily data for a given symbol and date range.
// returns the data along with metadata about the request.
func (s *Service) GetDailies(ctx context.Context, symbol string, from, to *time.Time, limit, offset int) (*DailyResponse, error) {
	// Fetch listing to ensure it exists and is accumulated
	listing, err := s.listingStore.FetchBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("GetDaily failed to fetch listing: %w", err)
	}
	if listing == nil {
		return nil, fmt.Errorf("GetDaily failed, listing with symbol %s not found", symbol)
	}
	if listing.Active == false {
		return nil, fmt.Errorf("GetDaily failed, listing with symbol %s is not active", symbol)
	}
	// If the listing is currently syncing it means that there is already a sync in progress.
	// In this case we can just fetch the existing data from the database, we don't need to trigger another sync.
	if listing.Syncing == true {
		data, err := s.dailyStore.FetchByListingID(ctx, symbol, from, to, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("GetDaily failed to fetch daily data: %w", err)
		}
		return &DailyResponse{
			Data: *data,
			Metadata: Metadata{
				Message:     "Data may be stale, listing is currently syncing",
				ResultCount: len(*data),
				TotalCount:  len(*data), // TODO: To implement
			},
		}, nil
	}
	// If the accumulated end date is nil or before the day before the current data, it means the data is not
	// up to date, so we should trigger a sync to fetch the latest data.
	targetEnd := latestBusinessDate(time.Now(), time.Local)
	if listing.AccumulatedEnd == nil || dateOnly(*listing.AccumulatedEnd, time.Local).Before(targetEnd) {
		err := s.listingStore.UpdateShouldAccumulate(ctx, listing.ID, true)
		if err != nil {
			s.log.Error(ctx, "GetDaily failed to update listing should_accumulate flag: %w", err)
		}
		if err == nil {
			// Trigger sync in a separate goroutine to avoid blocking the request, errors are logged but not returned
			go func() {
				fnCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
				defer cancel()
				err := s.syncDailyData(fnCtx, listing.ID, nil, nil)
				if err != nil {
					s.log.Error(fnCtx, "GetDaily failed to sync daily data for listing %s: %w", err, listing.Symbol)
				}
			}()
		}
	}
	// Fetch daily data from the database, this will return the existing data if a sync is in progress, or the up to date data if not.
	data, err := s.dailyStore.FetchByListingID(ctx, symbol, from, to, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("GetDaily failed to fetch daily data: %w", err)
	}
	return &DailyResponse{
		Data: *data,
		Metadata: Metadata{
			Message:     "Sync triggered; data may be stale until sync is complete",
			ResultCount: len(*data),
			TotalCount:  len(*data), // TODO: To implement
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
			s.log.Error(cleanupCtx, "syncDailyData failed to release sync lock: %w", err)
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
	// If to is nil, we want to accumulate up to the current date, so we can safely set it to now.
	if to == nil {
		now := time.Now().UTC()
		to = &now
	}
	for d, err := range s.client.GetEOD(ctx, []string{listing.Symbol}, from, to) {
		if err != nil {
			s.log.Error(ctx, "syncDailyData failed to fetch daily data for symbol %s: %w", err, listing.Symbol)
			continue
		}
		if err := s.dailyStore.Create(ctx, &d); err != nil {
			s.log.Error(ctx, "syncDailyData failed to persist daily data for symbol %s: %w", err, listing.Symbol)
			continue
		}
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
	if err == nil && existing != nil && existing.Source == source {
		return nil, fmt.Errorf("CreateListing failed, listing with symbol %s already exists", symbol)
	}
	// Create new listing and persist
	listing, err := NewListing(symbol, name, source, options...)
	if err != nil {
		return nil, fmt.Errorf("CreateListing failed to create listing: %w", err)
	}
	if err := s.listingStore.Create(ctx, listing); err != nil {
		return nil, fmt.Errorf("CreateListing failed to persist listing: %w", err)
	}
	// Fetch daily data for the listing and persist it, this is done in a
	// separate goroutine to avoid blocking the request, errors are logged but not returned
	go func() {
		fnCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		err := s.syncDailyData(fnCtx, listing.ID, nil, nil)
		if err != nil {
			s.log.Error(fnCtx, "CreateListing failed to sync daily data for listing %s: %w", err, listing.Symbol)
		}
	}()
	return listing, nil
}

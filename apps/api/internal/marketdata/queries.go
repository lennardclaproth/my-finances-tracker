package marketdata

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/sorting"
)

// QueryStore reads market-data listings, EOD datapoints, and providers.
type QueryStore interface {
	Get(ctx context.Context, id uuid.UUID) (*Listing, error)
	GetBySymbol(ctx context.Context, symbol string) (*Listing, error)
	List(ctx context.Context, limit, offset *int) ([]*Listing, error)
	Search(ctx context.Context, q string, limit, offset *int) ([]*Listing, int, error)
	ShouldAccumulate(ctx context.Context, lsID uuid.UUID, val bool) error
	CountEODByListing(ctx context.Context, lsID uuid.UUID, from, to *time.Time) (int, error)
	GetEODForListing(ctx context.Context, lsID uuid.UUID, from, to *time.Time, limit, offset *int, sort string) ([]*EOD, error)
	GetProviderByName(ctx context.Context, name ProviderName) (*Provider, error)
}

type Queries struct {
	qs QueryStore
	s  *Syncer
}

func NewListingHandler() *Queries {
	return &Queries{}
}

func (q *Queries) Listing(ctx context.Context, id uuid.UUID) (*Listing, error) {
	ls, err := q.qs.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get listing: failed to execute query: %w", err)
	}
	return ls, nil
}

func (q *Queries) ListingBySymbol(ctx context.Context, symbol string) (*Listing, error) {
	ls, err := q.qs.GetBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("get listing by symbol: failed to execute query: %w", err)
	}
	return ls, nil
}

// Provider returns the best provider candidate for the given provider name,
// wrapping ErrProviderNotFound when no matching provider exists.
func (q *Queries) Provider(ctx context.Context, name ProviderName) (*Provider, error) {
	provider, err := q.qs.GetProviderByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get provider: failed to execute query: %w", err)
	}
	return provider, nil
}

// ListListings returns all listings in deterministic order for UI presentation.
func (q *Queries) ListListings(ctx context.Context) ([]*Listing, error) {
	listings, err := q.qs.List(ctx, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("list listings: failed to execute query: %w", err)
	}
	return listings, nil
}

func (q *Queries) SearchListings(ctx context.Context, qs string, limit, offset int) ([]*Listing, int, error) {
	listings, total, err := q.qs.Search(ctx, qs, &limit, &offset)
	if err != nil {
		return nil, 0, fmt.Errorf("search listings: failed to execute query: %w", err)
	}
	return listings, total, nil
}

type Metadata struct {
	Message     string
	ResultCount int
	TotalCount  int
}

type EODResult struct {
	Data     []EOD
	Metadata Metadata
}

func (q *Queries) GetEODByListing(
	ctx context.Context,
	listingID uuid.UUID,
	from, to *time.Time,
	limit, offset int,
	sortOrder sorting.Direction,
) (*EODResult, error) {
	ls, err := q.qs.Get(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("handlers: GetByListing failed to get listing: %w", err)
	}
	if ls == nil {
		return nil, fmt.Errorf("handlers: GetByListing: %w", ErrListingNotFound)
	}
	if !ls.Active {
		return nil, fmt.Errorf("handlers: GetByListing listing with symbol %s is not active", ls.Symbol)
	}
	var res *EODResult
	if ls.Provider.IsManualIngestion() {
		res, err = getEODResult(ctx, ls, from, to, &limit, &offset, sortOrder, "Manual provider configured; automatic sync disabled", q.qs)
		if err != nil {
			return nil, fmt.Errorf("handlers: GetByListing failed to get EOD result: %w", err)
		}
		return res, nil
	}

	// If the listing is currently syncing it means that there is already a sync in progress.
	// In this case we can just fetch the existing data from the database, we don't need to trigger another sync.
	if ls.Syncing {
		res, err = getEODResult(
			ctx,
			ls,
			from,
			to,
			&limit,
			&offset,
			sortOrder,
			"Data may be stale, listing is currently syncing",
			q.qs,
		)
		if err != nil {
			return nil, fmt.Errorf("handlers: GetByListing failed to get EOD result: %w", err)
		}
		return res, nil
	}
	// If the accumulated end date is nil or before the day before the current data, it means the data is not
	// up to date, so we should trigger a sync to fetch the latest data.
	targetEnd := date.LatestBusinessDate(time.Now(), time.Local)
	if ls.AccumulatedEnd == nil || date.DateOnly(*ls.AccumulatedEnd, time.Local).Before(targetEnd) {
		err := q.qs.ShouldAccumulate(ctx, ls.ID, true)
		if err != nil {
			return nil, fmt.Errorf("get eod by listing: %w: %w", ErrShouldAccumulateFailed, err)
		}
		if err == nil {
			q.s.SyncEOD(ctx, ls.ID, from, to)
		}
	}
	// Fetch daily data from the database, this will return the existing data if a sync is in progress, or the up to date data if not.
	res, err = getEODResult(
		ctx,
		ls,
		from,
		to,
		&limit,
		&offset,
		sortOrder,
		"Sync triggered; data may be stale until sync is complete",
		q.qs,
	)
	if err != nil {
		return nil, fmt.Errorf("handlers: GetByListing failed to get EOD result: %w", err)
	}
	return res, nil
}

func getEODResult(
	ctx context.Context,
	ls *Listing,
	from, to *time.Time,
	limit, offset *int,
	sortOrder sorting.Direction,
	msg string,
	qs QueryStore,
) (*EODResult, error) {
	data, err := qs.GetEODForListing(
		ctx,
		ls.ID,
		from,
		to,
		limit,
		offset,
		string(sortOrder),
	)
	if err != nil {
		return nil, fmt.Errorf("GetDaily failed to fetch daily data: %w", err)
	}
	totalCount, err := qs.CountEODByListing(ctx, ls.ID, from, to)
	if err != nil {
		return nil, fmt.Errorf("GetDaily failed to fetch total daily count: %w", err)
	}
	res := make([]EOD, 0, len(data))
	for _, item := range data {
		if item == nil {
			continue
		}

		res = append(res, *item)
	}
	return &EODResult{
		Data: res,
		Metadata: Metadata{
			Message:     msg,
			ResultCount: len(data),
			TotalCount:  totalCount,
		},
	}, nil
}

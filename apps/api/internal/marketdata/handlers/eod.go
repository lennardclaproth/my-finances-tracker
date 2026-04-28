package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

// EODSortOrder controls date ordering for daily retrieval.
type EODSortOrder string

type Metadata struct {
	Message     string
	ResultCount int
	TotalCount  int
}

type EODResult struct {
	Data     []marketdata.EOD
	Metadata Metadata
}

const (
	SortEODAsc  EODSortOrder = "asc"
	SortEODDesc EODSortOrder = "desc"
)

// eodGetter retrieves EOD data for a listing, supporting pagination and date range filtering.
type eodGetter interface {
	GetByListing(ctx context.Context, listingID uuid.UUID, from, to *time.Time, limit, offset int, sortOrder string) (*[]marketdata.EOD, error)
}

// eodCounter retrieves the total count of EOD records for a listing and date range, which is used for pagination metadata.
type eodCounter interface {
	CountByListing(ctx context.Context, listingID uuid.UUID, from, to *time.Time) (int, error)
}

type EODHandler struct {
	lf  listingGetter
	la  listingAccumulator
	ll  listingLocker
	sh  *SyncHandler
	eg  eodGetter
	ec  eodCounter
	log logging.Logger
}

func NewEODHandler(lf listingGetter, la listingAccumulator, ll listingLocker, sh *SyncHandler, eg eodGetter, ec eodCounter, log logging.Logger) *EODHandler {
	return &EODHandler{
		lf:  lf,
		la:  la,
		ll:  ll,
		sh:  sh,
		eg:  eg,
		ec:  ec,
		log: log,
	}
}

func (h *EODHandler) GetBySymbol(
	ctx context.Context,
	symbol string,
	from, to *time.Time,
	limit, offset int,
	sortOrder EODSortOrder,
) error {
	// GetListing for symbol
	// GetByListing
	return nil
}

func (h *EODHandler) GetByListing(
	ctx context.Context,
	listingID uuid.UUID,
	from, to *time.Time,
	limit, offset int,
	sortOrder EODSortOrder,
) (*EODResult, error) {
	ls, err := h.lf.GetIncludingProvider(ctx, listingID)
	if err != nil {
		return nil, fmt.Errorf("handlers: GetByListing failed to get listing: %w", err)
	}
	if ls == nil {
		return nil, fmt.Errorf("handlers: GetByListing listing not found")
	}
	if !ls.Active {
		return nil, fmt.Errorf("handlers: GetByListing listing with symbol %s is not active", ls.Symbol)
	}
	var res *EODResult
	if ls.Provider.IsManualIngestion() {
		res, err = h.getEODResult(ctx, ls, from, to, limit, offset, sortOrder, "Manual provider configured; automatic sync disabled")
		if err != nil {
			return nil, fmt.Errorf("handlers: GetByListing failed to get EOD result: %w", err)
		}
		return res, nil
	}

	// If the listing is currently syncing it means that there is already a sync in progress.
	// In this case we can just fetch the existing data from the database, we don't need to trigger another sync.
	if ls.Syncing {
		res, err = h.getEODResult(
			ctx,
			ls,
			from,
			to,
			limit,
			offset,
			sortOrder,
			"Data may be stale, listing is currently syncing",
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
		err := h.la.SetShouldAccumulate(ctx, ls.ID, true)
		if err != nil {
			h.log.Error(ctx, "GetByListing failed to update listing should_accumulate flag", err, "listing_id", ls.ID, "symbol", ls.Symbol)
		}
		if err == nil {
			h.sh.SyncEOD(ctx, ls.ID, from, to)
		}
	}
	// Fetch daily data from the database, this will return the existing data if a sync is in progress, or the up to date data if not.
	res, err = h.getEODResult(
		ctx,
		ls,
		from,
		to,
		limit,
		offset,
		sortOrder,
		"Sync triggered; data may be stale until sync is complete",
	)
	if err != nil {
		return nil, fmt.Errorf("handlers: GetByListing failed to get EOD result: %w", err)
	}
	return res, nil
}

func (h *EODHandler) getEODResult(
	ctx context.Context,
	ls *marketdata.Listing,
	from, to *time.Time,
	limit, offset int,
	sortOrder EODSortOrder,
	msg string,
) (*EODResult, error) {
	data, err := h.eg.GetByListing(
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
	totalCount, err := h.ec.CountByListing(ctx, ls.ID, from, to)
	if err != nil {
		return nil, fmt.Errorf("GetDaily failed to fetch total daily count: %w", err)
	}
	return &EODResult{
		Data: *data,
		Metadata: Metadata{
			Message:     "Data may be stale, listing is currently syncing",
			ResultCount: len(*data),
			TotalCount:  totalCount,
		},
	}, nil
}

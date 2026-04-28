package handlers

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

type EODFetcher interface {
	GetEOD(ctx context.Context, symbols []string, from, to *time.Time) iter.Seq2[marketdata.EOD, error]
}

type eodInserter interface {
	Insert(ctx context.Context, eod *marketdata.EOD) error
}

type SyncHandler struct {
	ll  listingLocker
	ls  listingGetter
	la  listingAccumulator
	efs map[marketdata.Source]EODFetcher
	ei  eodInserter
	log logging.Logger
}

func NewSyncHandler() *SyncHandler {
	return &SyncHandler{}
}

// TODO: Wrap in go routine, implement worker pool and queue
func (h *SyncHandler) SyncEOD(ctx context.Context, lsID uuid.UUID, from, to *time.Time) error {
	// Acquire sync lock, this ensures that only one sync can be in progress.
	// The acquiring of the lock is an atomic operation, returning an error if the lock is
	// already held, which means another sync is in progress.
	acquired, err := h.ll.TryAcquireSyncLock(ctx, lsID)
	if err != nil {
		return fmt.Errorf("handlers: SyncForListing failed to acquire sync lock: %w", err)
	}
	if !acquired {
		return ErrSyncInProgress
	}
	defer func() {
		// Ensure we always clear the syncing flag, even if the request context is canceled.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.ll.ReleaseSyncLock(cleanupCtx, lsID); err != nil {
			h.log.Error(cleanupCtx, "handlers: SyncForListing failed to release sync lock", err, "listing_id", lsID)
		}
	}()
	// Get most up to date listing data
	listing, err := h.ls.GetIncludingProvider(ctx, lsID)
	if err != nil {
		return fmt.Errorf("handlers: SyncForListing failed to fetch listing: %w", err)
	}
	// If from is nil it means there is no override on the accumulation date
	// so we should continue accumulating from the last accumulated date, which is the accumulated end date.
	if from == nil {
		// Accumulated end defaults to nil so we can safely pass this
		// it is important to note that if accumulated end is nil, the MarketStackClient will not include the
		// date_from parameter in the request, which will cause MarketStack to return all available data for the symbol.
		// This is desired.
		from = listing.AccumulatedEnd
	}
	// If to is nil, accumulate only through the latest completed business date (exclude today).
	if to == nil {
		latestCompletedBusinessDate := date.LatestBusinessDate(time.Now(), time.Local)
		to = &latestCompletedBusinessDate
	}
	// Find EOD fetcher for listing source
	c := h.efs[listing.Source]
	if c == nil {
		return fmt.Errorf("handlers: SyncForListing failed to find EOD fetcher for source %s", listing.Source)
	}
	// Fetch daily data from the EOD fetcher and persist it, this will return an error for each individual day that fails to
	// fetch or persist, but will continue processing the rest of the data.
	for eod, err := range c.GetEOD(ctx, []string{listing.Symbol}, from, to) {
		if err != nil {
			h.log.Error(ctx, "handlers: SyncForListing failed to fetch daily data", err, "listing_id", listing.ID, "symbol", listing.Symbol)
			continue
		}
		eod.ListingID = listing.ID
		// TODO: implement batch insert
		if err := h.ei.Insert(ctx, &eod); err != nil {
			h.log.Error(ctx, "handlers: SyncForListing failed to persist daily data", err, "listing_id", listing.ID, "symbol", listing.Symbol, "date", eod.Date)
			continue
		}
	}
	if err := h.la.SetAccumulatedRange(ctx, lsID, from, to); err != nil {
		h.log.Warn(ctx, "handlers: SyncForListing failed to update accumulated range", "listing_id", lsID.String(), "error", err.Error())
	}
	return nil
}

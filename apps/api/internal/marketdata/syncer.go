package marketdata

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/date"
)

type EODFetcher interface {
	GetEOD(ctx context.Context, symbols []string, from, to *time.Time) iter.Seq2[EOD, error]
}

// SyncStore reads a listing and persists EOD sync progress.
type SyncStore interface {
	// Listing returns the listing (including provider) for the given id.
	Listing(ctx context.Context, id uuid.UUID) (*Listing, error)
	SetAccumulatedRange(ctx context.Context, id uuid.UUID, from, to *time.Time) error
	TryAcquireSyncLock(ctx context.Context, id uuid.UUID) (bool, error)
	ReleaseSyncLock(ctx context.Context, id uuid.UUID) error
	InsertEOD(ctx context.Context, eod *EOD) error
}

type Syncer struct {
	ss  SyncStore
	efs map[Source]EODFetcher
}

type SyncEODResult struct {
	ListingID     uuid.UUID
	Symbol        string
	FetchedCount  int
	InsertedCount int
	FetchErrors   []error
	InsertErrors  []error
	AccumError    error
}

// TODO: Wrap in go routine, implement worker pool and queue
func (s *Syncer) SyncEOD(ctx context.Context, lsID uuid.UUID, from, to *time.Time) (result *SyncEODResult, err error) {
	// Acquire sync lock, this ensures that only one sync can be in progress.
	// The acquiring of the lock is an atomic operation, returning an error if the lock is
	// already held, which means another sync is in progress.
	acquired, err := s.ss.TryAcquireSyncLock(ctx, lsID)
	if err != nil {
		return nil, fmt.Errorf("sync eod: failed to execute sync lock: %w", err)
	}
	if !acquired {
		return nil, fmt.Errorf("sync eod: failed to acquire lock: %w", ErrSyncInProgress)
	}
	// Considering the defer func is registered here we can
	// assume that the lock has been acquired.
	defer func() {
		// Ensure we always clear the syncing flag, even if the request context is canceled.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if releaseErr := s.ss.ReleaseSyncLock(cleanupCtx, lsID); releaseErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf("sync eod: %w: %w", ErrReleaseSyncLock, releaseErr),
			)
		}
	}()
	// Get most up to date listing data
	listing, err := s.ss.Listing(ctx, lsID)
	if err != nil {
		return nil, fmt.Errorf("sync eod: failed to get listing: %w", err)
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
	c, ok := s.efs[listing.Source]
	if !ok {
		return nil, fmt.Errorf("sync eod: failed to find fetcher for listing source %s", listing.Source)
	}
	// Fetch EOD data from the EOD fetcher and persist it, this will return an error for each individual day that fails to
	// fetch or persist, but will continue processing the rest of the data.
	for eod, fetchErr := range c.GetEOD(ctx, []string{listing.Symbol}, from, to) {
		if err != nil {
			result.FetchErrors = append(result.FetchErrors, fetchErr)
			continue
		}
		result.FetchedCount++
		eod.ListingID = listing.ID

		// TODO: implement batch insert
		if insertErr := s.ss.InsertEOD(ctx, &eod); insertErr != nil {
			result.InsertErrors = append(result.InsertErrors, insertErr)
		}
	}
	if rangeErr := s.ss.SetAccumulatedRange(ctx, lsID, from, to); rangeErr != nil {
		result.AccumError = rangeErr
	}
	return result, nil
}

package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/marketdata"
)

// listingGetter retrieves listing details, including provider information, which is required for syncing.
type listingGetter interface {
	GetIncludingProvider(ctx context.Context, id uuid.UUID) (*marketdata.Listing, error)
	GetBySymbol(ctx context.Context, symbol string) (*marketdata.Listing, error)
}

// listingAccumulator controls accumulation state and range for a listing, which is used to prevent concurrent syncs and to track the date range of accumulated data.
type listingAccumulator interface {
	SetShouldAccumulate(ctx context.Context, id uuid.UUID, shouldAccumulate bool) error
	SetAccumulatedRange(ctx context.Context, id uuid.UUID, from, to *time.Time) error
}

// listingLocker provides distributed locking capabilities for a listing, ensuring that only one sync can be in progress for a listing at any time.
type listingLocker interface {
	TryAcquireSyncLock(ctx context.Context, id uuid.UUID) (bool, error)
	ReleaseSyncLock(ctx context.Context, id uuid.UUID) error
}

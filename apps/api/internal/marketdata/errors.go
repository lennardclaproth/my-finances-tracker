package marketdata

import (
	"errors"
)

// Sync errors
var (
	ErrReleaseSyncLock = errors.New("Failed to release lock for the listing")
	ErrSyncInProgress  = errors.New("sync is already in progress for this listing")
)

// Accumulate errors
var (
	ErrShouldAccumulateFailed = errors.New("Failed to set should accumulate")
)

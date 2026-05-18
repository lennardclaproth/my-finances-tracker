package assets

import "errors"

// Asset errors.
var (
	ErrAssetNotFound = errors.New("asset not found")
)

// Account errors
var (
	ErrAccountNotFound = errors.New("account not found")
)

// Class errors
var (
	ErrClassNotFound      = errors.New("class not found")
	ErrClassNotManual     = errors.New("asset class must be manual")
	ErrClassReserved      = errors.New("asset class cannot use a reserved name")
	ErrClassNameEmpty     = errors.New("asset class name cannot be empty")
	ErrClassAlreadyExists = errors.New("asset class already exists")
)

// Syncer errors
var (
	ErrSyncInProgress  = errors.New("sync already in progress")
	ErrReleaseSyncLock = errors.New("Failed to release lock for the listing")
)

package files

import "errors"

var (
	// ErrContentRequired indicates a write command was called without file content.
	ErrContentRequired = errors.New("file content is required")
	// ErrPathRequired indicates a read or remove command was called without a stored path.
	ErrPathRequired = errors.New("file path is required")
	// ErrStoreNotConfigured indicates file use cases were created without a backing store.
	ErrStoreNotConfigured = errors.New("file store is not configured")
)

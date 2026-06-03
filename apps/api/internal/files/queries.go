package files

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type queryStore interface {
	ReadCsv(path string) (io.ReadCloser, error)
}

// Queries exposes file-storage read-side use cases.
type Queries struct {
	store queryStore
}

// NewQueries creates file-storage read-side use cases.
func NewQueries(store queryStore) *Queries {
	return &Queries{store: store}
}

// ReadCsv opens a persisted CSV-compatible file for streaming.
func (q *Queries) ReadCsv(ctx context.Context, path string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if strings.TrimSpace(path) == "" {
		return nil, ErrPathRequired
	}
	if q.store == nil {
		return nil, ErrStoreNotConfigured
	}
	rc, err := q.store.ReadCsv(path)
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	return rc, nil
}

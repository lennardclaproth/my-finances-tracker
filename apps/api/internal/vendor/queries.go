package vendor

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Queries exposes vendor read-side use cases.
type Queries struct {
	qs QueryStore
}

// QueryStore reads vendor records for query use cases.
type QueryStore interface {
	GetByID(ctx context.Context, vID uuid.UUID) (*Vendor, error)
	ListActive(ctx context.Context) ([]*Vendor, error)
}

// NewQueries creates vendor read-side use cases.
func NewQueries(qs QueryStore) *Queries {
	return &Queries{qs: qs}
}

// GetById returns a vendor by ID.
func (q *Queries) GetById(ctx context.Context, vID uuid.UUID) (*Vendor, error) {
	v, err := q.qs.GetByID(ctx, vID)
	if err != nil {
		return nil, fmt.Errorf("get by id: failed to execute query: %w", err)
	}
	return v, nil
}

// ListActive returns active vendors in presentation order.
func (q *Queries) ListActive(ctx context.Context) ([]*Vendor, error) {
	vendors, err := q.qs.ListActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active vendors: failed to execute query: %w", err)
	}
	if vendors == nil {
		return []*Vendor{}, nil
	}
	return vendors, nil
}

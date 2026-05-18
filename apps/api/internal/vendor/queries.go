package vendor

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Queries struct {
	qs QueryStore
}

type QueryStore interface {
	ByID(ctx context.Context, vID uuid.UUID) (*Vendor, error)
}

func (q *Queries) GetById(ctx context.Context, vID uuid.UUID) (*Vendor, error) {
	v, err := q.qs.ByID(ctx, vID)
	if err != nil {
		return nil, fmt.Errorf("get by id: failed to execute query: %w", err)
	}
	return v, nil
}

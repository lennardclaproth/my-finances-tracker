package account

import (
	"context"

	"github.com/google/uuid"
)

type Queries struct {
	qs QueryStore
}

// QueryStore reads account records for query use cases.
type QueryStore interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
}

// NewQueries creates account read-side use cases.
func NewQueries(qs QueryStore) *Queries {
	return &Queries{
		qs: qs,
	}
}

func (q *Queries) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return q.qs.Exists(ctx, id)
}

func (q *Queries) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	return q.qs.GetByID(ctx, id)
}

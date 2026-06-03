package account

import (
	"context"

	"github.com/google/uuid"
)

type Queries struct {
	qs queryStore
}

type queryStore interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
}

func NewExistsHandler(qs queryStore) *Queries {
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

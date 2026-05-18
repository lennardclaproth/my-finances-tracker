package account

import (
	"context"

	"github.com/google/uuid"
)

type Queries struct {
	querier existsQuerier
}

type existsQuerier interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

func NewExistsHandler(querier existsQuerier) *Queries {
	return &Queries{
		querier: querier,
	}
}

func (q *Queries) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return q.querier.Exists(ctx, id)
}

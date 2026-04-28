package handlers

import (
	"context"

	"github.com/google/uuid"
)

type ExistsHandler struct {
	querier existsQuerier
}

type existsQuerier interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

func NewExistsHandler(querier existsQuerier) *ExistsHandler {
	return &ExistsHandler{
		querier: querier,
	}
}

func (h *ExistsHandler) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	return h.querier.Exists(ctx, id)
}
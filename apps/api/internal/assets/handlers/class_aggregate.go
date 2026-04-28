package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type ClassAggregateHandler struct {
	cs classSummer
}

// TODO: think of better num
type classSummer interface {
	SumWorth(ctx context.Context, accID, classID uuid.UUID) (money.Price, error)
}

func NewClassAggregateHandler() *ClassAggregateHandler {
	return &ClassAggregateHandler{}
}

func (h *ClassAggregateHandler) AggregateWorth(ctx context.Context, accID, classID uuid.UUID) (money.Price, error) {
	return h.cs.SumWorth(ctx, accID, classID)
}

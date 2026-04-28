package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type MutationCreationHandler struct {
	mc mutationCreator
}

type mutationCreator interface {
	Create(ctx context.Context, mutation *assets.Mutation) error
}

func NewMutationCreationHandler(mc mutationCreator) *MutationCreationHandler {
	return &MutationCreationHandler{
		mc: mc,
	}
}

func (h *MutationCreationHandler) Create(ctx context.Context, accID, classID, assetID uuid.UUID,
	changeType assets.ChangeType,
	direction *assets.ChangeDirection,
	amount, previousWorth, classTotalWorth money.Price,
	effectiveDate time.Time,
	note *string) (*assets.Mutation, error) {
	// TODO: add total worth calculation, this should be determined by the newworth - previousworth
	// the delta here should be added to the class total worth because as input here
	// we get the current class total worth.
	mutation, err := assets.NewMutation(
		accID, classID, assetID,
		changeType, direction,
		amount, previousWorth, classTotalWorth,
		effectiveDate, note,
	)
	if err != nil {
		return nil, fmt.Errorf("handlers: Create mutation failed: %w", err)
	}
	err = h.mc.Create(ctx, mutation)
	if err != nil {
		return nil, fmt.Errorf("handlers: Create mutation failed: %w", err)
	}
	return mutation, nil
}

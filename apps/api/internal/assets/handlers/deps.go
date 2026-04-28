package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
)

type accountExistenceChecker interface {
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type assetGetter interface {
	Get(ctx context.Context, id uuid.UUID) (*assets.Asset, error)
}

type unitOfWork interface {
	Do(ctx context.Context, fn func(txCtx context.Context) error) error
}

type assetUpdater interface {
	SetWorth(ctx context.Context, asset *assets.Asset) error
}

type classGetter interface {
	Get(ctx context.Context, accID, classID uuid.UUID) (*assets.Class, error)
}

type classAggregator interface {
	AggregateValue(ctx context.Context, accID, classID uuid.UUID) (money.Price, error)
}

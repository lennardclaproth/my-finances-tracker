package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type AssetCreator struct {
	aec accountExistsChecker
	cg  classGetter
	uow unitOfWork
	as  assetStorer
	ca  classAggregator
	mc  MutationCreationHandler
}

type assetStorer interface {
	Create(ctx context.Context, asset *assets.Asset) error
}

func NewAssetCreator(
	aec accountExistenceChecker,
	cg classGetter,
	uow unitOfWork,
	as assetStorer,
	ca classAggregator,
	mc MutationCreationHandler,
) *AssetCreator {
	return &AssetCreator{
		aec: aec,
		cg:  cg,
		uow: uow,
		as:  as,
		ca:  ca,
		mc:  mc,
	}
}

func (h *AssetCreator) Create(
	ctx context.Context,
	accID, classID uuid.UUID,
	name, string,
	initialWorth money.Price,
	date time.Time,
	note *string,
) (*assets.Asset, error) {
	// Check if account exists
	exists, err := h.aec.Exists(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("handlers: CreateAsset failed: %w", err)
	}
	if !exists {
		return nil, ErrAccountNotFound
	}
	// Get class and check if it exists
	class, err := h.cg.Get(ctx, accID, classID)
	if err != nil {
		return nil, fmt.Errorf("handlers: CreateAsset failed: %w", err)
	}
	if class == nil {
		return nil, ErrAssetClassNotFound
	}
	asset, err := assets.NewAsset(accID, classID, name, initialWorth, date, note)
	if err != nil {
		return nil, fmt.Errorf("handlers: CreateAsset failed: %w", err)
	}
	if err != nil {
		return nil, fmt.Errorf("handlers: CreateAsset failed: %w", err)
	}
	// Start a transaction to ensure atomicity
	err = h.uow.Do(ctx, func(txCtx context.Context) error {
		// Create the asset
		if err := h.as.Create(txCtx, asset); err != nil {
			return fmt.Errorf("handlers: CreateAsset failed: %w", err)
		}
		// Get current value of class
		classTotal, err := h.ca.AggregateValue(txCtx, accID, classID)
		if err != nil {
			return fmt.Errorf("handlers: CreateAsset failed: %w", err)
		}
		// Create mutation entry for initial worth with class total aggregated
		h.mc.Create(txCtx, accID, classID, asset.ID,
			assets.ChangeTypeSet, nil,
			initialWorth, 0, classTotal,
			date, note)
		return nil
	})
	if err != nil {
		return nil, err
	}
	// TODO: publish snapshot rebuild event for account.
	return asset, nil
}

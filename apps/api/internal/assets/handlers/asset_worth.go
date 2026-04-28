package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

type AssetWorthSetter struct {
	aec accountExistenceChecker
	cg  classGetter
	uow unitOfWork
	ag  assetGetter
	ca  classAggregator
	mc  MutationCreationHandler
	au  assetUpdater
}

func NewAssetWorthSetter(aec accountExistenceChecker, cg classGetter, uow unitOfWork, ag assetGetter, ca classAggregator, mc MutationCreationHandler, au assetUpdater) *AssetWorthSetter {
	return &AssetWorthSetter{
		aec: aec,
		cg:  cg,
		uow: uow,
		ag:  ag,
		ca:  ca,
		mc:  mc,
		au:  au,
	}
}

func (h *AssetWorthSetter) UpdateWorth(
	ctx context.Context,
	accID, classID, assetID uuid.UUID,
	worth money.Price,
	changeType assets.ChangeType,
	direction *assets.ChangeDirection,
	effectiveDate time.Time,
	note *string,
) error {
	// Check if account exists
	exists, err := h.aec.Exists(ctx, accID)
	if err != nil {
		return fmt.Errorf("handlers: updating asset worth failed: %w", err)
	}
	if !exists {
		return ErrAccountNotFound
	}
	// Check if class exists
	class, err := h.cg.Get(ctx, accID, classID)
	if err != nil {
		return fmt.Errorf("handlers: updating asset worth failed: %w", err)
	}
	if class == nil {
		return ErrClassNotFound
	}
	// guard against updating non manual asset classes
	if class.Source != assets.ClassSourceManual && changeType != assets.ChangeTypeSet {
		// Portfolio worth is set only by sync flow.
		return ErrAssetClassNotManual
	}
	if class.Source == assets.ClassSourcePortfolio {
		return ErrAssetClassNotManual
	}
	// Start a transaction to ensure atomicity
	err = h.uow.Do(ctx, func(txCtx context.Context) error {
		// Get the asset
		asset, err := h.ag.Get(txCtx, assetID)
		if err != nil {
			return fmt.Errorf("handlers: updating asset worth failed: %w", err)
		}
		if asset == nil {
			return ErrAssetNotFound
		}
		previousWorth := asset.CurrentWorth
		classTotal, err := h.ca.AggregateValue(txCtx, accID, classID)
		// Create mutation for change
		m, err := h.mc.Create(txCtx, accID, classID, asset.ID,
			changeType, direction,
			worth, previousWorth, classTotal,
			effectiveDate, note)
		if err != nil {
			return fmt.Errorf("handlers: update asset worth failed: %w", err)
		}
		asset.CurrentWorth = m.NewWorth
		err = h.au.SetWorth(txCtx, asset)
		if err != nil {
			return fmt.Errorf("handlers: update asset worth faild: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	// TODO: publish snapshot rebuild event for account.
	return nil
}

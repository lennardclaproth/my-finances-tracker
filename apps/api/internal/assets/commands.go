package assets

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

// TODO: fix interfaces

type commandStore interface {
	CreateAsset(ctx context.Context, asset *Asset) error
	CreateAccount(ctx context.Context, account *Account) error
	SetWorth(ctx context.Context, asset *Asset) error
	CreateClass(ctx context.Context, class *Class) error
	UpdateClass(ctx context.Context, class *Class) error
	CreateMutation(ctx context.Context, mut *Mutation) error
	DeleteClass(ctx context.Context, classID uuid.UUID) error
}

type commandGetter interface {
	Class(ctx context.Context, classID uuid.UUID) (*Class, error)
	Asset(ctx context.Context, assetID uuid.UUID) (*Asset, error)
}

type classAggregator interface {
	AggregateValue(ctx context.Context, accID, classID uuid.UUID) (money.Price, error)
	AggregateValues(ctx context.Context, accID, classIDs []uuid.UUID) (map[uuid.UUID]money.Price, error)
}

type unitOfWork interface {
	Do(ctx context.Context, fn func(txCtx context.Context) error) error
}

type Commands struct {
	cs  commandStore
	cg  commandGetter
	aq  account.Queries
	uow unitOfWork
	ca  classAggregator
}

func (c *Commands) CreateAsset(
	ctx context.Context,
	accID, classID uuid.UUID,
	name string,
	initialWorth money.Price,
	date time.Time,
	note *string,
) (*Asset, error) {
	// Check if account exists
	exists, err := c.aq.Exists(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("create asset: account existence checker failed: %w", err)
	}
	if !exists {
		return nil, ErrAccountNotFound
	}
	// Get class and check if it exists
	class, err := c.cg.Class(ctx, classID)
	if err != nil {
		return nil, fmt.Errorf("create asset: get class failed: %w", err)
	}
	if class == nil {
		return nil, ErrClassNotFound
	}
	asset, err := NewAsset(accID, classID, name, initialWorth, *note)
	if err != nil {
		return nil, fmt.Errorf("create asset: failed to create new asset: %w", err)
	}
	// Start a transaction to ensure atomicity
	err = c.uow.Do(ctx, func(txCtx context.Context) error {
		// Create the asset
		if err := c.cs.CreateAsset(txCtx, asset); err != nil {
			return fmt.Errorf("create asset: failed to store asset: %w", err)
		}
		// Get current value of class
		classTotal, err := c.ca.AggregateValue(txCtx, accID, classID)
		if err != nil {
			return fmt.Errorf("create asset: could not aggregate the class total : %w", err)
		}
		// Create mutation entry for initial worth with class total aggregated
		c.CreateMutation(txCtx,
			accID, classID, asset.ID,
			ChangeTypeSet, nil,
			initialWorth, 0, classTotal,
			date, note)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("create asset: failed to execute transaction: %w", err)
	}
	// TODO: publish snapshot rebuild event for account.
	return asset, nil
}

// CreateClass creates a manual class for an account.
func (c *Commands) CreateClass(
	ctx context.Context,
	accID uuid.UUID,
	name string,
) (*Class, error) {
	// Guard against account existence
	exists, err := c.aq.Exists(ctx, accID)
	if err != nil {
		return nil, fmt.Errorf("create class: failed to check account existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("create class: %w", ErrAccountNotFound)
	}
	// Create new class
	class, err := NewClass(accID, nil, name)
	if err != nil {
		return nil, fmt.Errorf("create class: failed to create domain model: %w", err)
	}
	// Store class
	err = c.cs.CreateClass(ctx, class)
	if err != nil {
		return nil, fmt.Errorf("create class: failed to store class: %w", err)
	}
	// TODO: publish rebuild event
	return class, nil
}

func (c *Commands) CreateMutation(
	ctx context.Context,
	accID, classID, assetID uuid.UUID,
	changeType ChangeType,
	direction *ChangeDirection,
	amount, previousWorth, classTotalWorth money.Price,
	effectiveDate time.Time,
	note *string) (*Mutation, error) {
	// TODO: add total worth calculation, this should be determined by the newworth - previousworth
	// the delta here should be added to the class total worth because as input here
	// we get the current class total worth.
	mutation, err := NewMutation(
		accID, classID, assetID,
		changeType, direction,
		amount, previousWorth, classTotalWorth,
		effectiveDate, note,
	)
	if err != nil {
		return nil, fmt.Errorf("create mutation: failed to create domain model: %w", err)
	}
	err = c.cs.CreateMutation(ctx, mutation)
	if err != nil {
		return nil, fmt.Errorf("create mutation: failed to store mutation: %w", err)
	}
	return mutation, nil
}

func (c *Commands) UpdateAssetWorth(
	ctx context.Context,
	accID, classID, assetID uuid.UUID,
	worth money.Price,
	changeType ChangeType,
	direction *ChangeDirection,
	effectiveDate time.Time,
	note *string,
) error {
	// Check if class exists
	class, err := c.cg.Class(ctx, classID)
	if err != nil {
		return fmt.Errorf("update asset worth: failed to get class: %w", err)
	}
	if class == nil {
		return fmt.Errorf("update asset worth: %w", ErrClassNotFound)
	}
	// guard against updating non manual asset classes
	if class.Source != ClassSourceManual && changeType != ChangeTypeSet {
		// Portfolio worth is set only by sync flow.
		return fmt.Errorf("update asset worth: %w", ErrClassNotManual)
	}
	if class.Source == ClassSourcePortfolio {
		return fmt.Errorf("update asset worth: %w", ErrClassReserved)
	}
	// Start a transaction to ensure atomicity
	err = c.uow.Do(ctx, func(txCtx context.Context) error {
		// Get the asset
		asset, err := c.cg.Asset(txCtx, assetID)
		if err != nil {
			return fmt.Errorf("update asset worth: failed to get asset: %w", err)
		}
		if asset == nil {
			return fmt.Errorf("update asset worth: %w", ErrAssetNotFound)
		}
		previousWorth := asset.CurrentWorth
		classTotal, err := c.ca.AggregateValue(txCtx, accID, classID)
		// Create mutation for change
		m, err := c.CreateMutation(txCtx, accID, classID, asset.ID,
			changeType, direction,
			worth, previousWorth, classTotal,
			effectiveDate, note)
		if err != nil {
			return fmt.Errorf("update asset worth: failed to create mutation: %w", err)
		}
		asset.CurrentWorth = m.NewWorth
		err = c.cs.SetWorth(txCtx, asset)
		if err != nil {
			return fmt.Errorf("update asset worth: setting the new worth failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("update asset worth: failed to execute transaction: %w", err)
	}
	// TODO: publish snapshot rebuild event for account.
	return nil
}

// UpdateClass mutates a manual class name/archive status.
func (c *Commands) UpdateClass(ctx context.Context, classID uuid.UUID, name string, archived bool) error {
	class, err := c.cg.Class(ctx, classID)
	if err != nil {
		return fmt.Errorf("update class: fetch class: %w", err)
	}
	if class == nil {
		return fmt.Errorf("update class: %w", ErrClassNotFound)
	}
	if class.Source != ClassSourceManual {
		return fmt.Errorf("update class: %w", ErrClassNotManual)
	}
	if err := class.Update(&name, &archived); err != nil {
		return fmt.Errorf("update class: failed to update domain model: %w", err)
	}
	if err := c.cs.UpdateClass(ctx, class); err != nil {
		return fmt.Errorf("update class: failed to store changes: %w", err)
	}
	// TODO: publish snapshot rebuild event
	return nil
}

// DeleteClass removes a manual class and related items/history.
func (c *Commands) DeleteClass(ctx context.Context, classID uuid.UUID) error {
	class, err := c.cg.Class(ctx, classID)
	if err != nil {
		return fmt.Errorf("delete asset: failed to get class: %w", err)
	}
	if class == nil {
		return ErrClassNotFound
	}
	if class.Source != ClassSourceManual {
		return ErrClassNotManual
	}
	if err := c.cs.DeleteClass(ctx, classID); err != nil {
		return fmt.Errorf("delete asset: failed to delete: %w", err)
	}

	// TODO: publish snapshot rebuild event
	return nil
}

func (c *Commands) CreateAccount(ctx context.Context, accountID uuid.UUID) (*Account, error) {
	acc := NewAccount(accountID)
	if err := c.cs.CreateAccount(ctx, acc); err != nil {
		return nil, fmt.Errorf("create account: failed to store account: %w", err)
	}
	return acc, nil
}

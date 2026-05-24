package assets

import (
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

// ChangeType identifies whether worth was replaced or adjusted.
type ChangeType string

const (
	// ChangeTypeSet replaces an item worth with a new value.
	ChangeTypeSet ChangeType = "SET"
	// ChangeTypeAdjust increments/decrements an item worth by a delta.
	ChangeTypeAdjust ChangeType = "ADJUST"
)

// ChangeDirection identifies adjust direction.
type ChangeDirection string

const (
	// ChangeDirectionIncrease increases worth by amount.
	ChangeDirectionIncrease ChangeDirection = "INCREASE"
	// ChangeDirectionDecrease decreases worth by amount.
	ChangeDirectionDecrease ChangeDirection = "DECREASE"
)

// Mutation captures one worth mutation for a tracked asset entry.
type Mutation struct {
	ID         uuid.UUID        `db:"id"`
	AccountID  uuid.UUID        `db:"account_id"`
	ClassID    uuid.UUID        `db:"class_id"`
	AssetID    uuid.UUID        `db:"asset_id"`
	ChangeType ChangeType       `db:"change_type"`
	Direction  *ChangeDirection `db:"direction"`
	// The amount of the change
	Amount money.Price `db:"amount"`
	// The previous worth of the Asset
	PreviousWorth money.Price `db:"previous_worth"`
	// The new worth of the Asset
	NewWorth        money.Price `db:"new_worth"`
	ClassTotalWorth money.Price `db:"class_total_worth"`
	EffectiveDate   time.Time   `db:"effective_date"`
	Note            *string     `db:"note"`
	CreatedAt       time.Time   `db:"created_at"`
}

func NewMutation(
	accID, classID, assetID uuid.UUID,
	changeType ChangeType,
	direction *ChangeDirection,
	amount, previousWorth, classTotalWorth money.Price,
	effectiveDate time.Time,
	note *string,
) (*Mutation, error) {
	// determine how to apply the change in worth, set nextWorth to worth, this
	// is the base case (ChangeType is "SET"). When ChangeType is "ADJUST" we need
	// to determine whether to decrease or increase the previousworth.
	newWorth := amount
	if changeType == ChangeTypeAdjust {
		if direction != nil && *direction == ChangeDirectionDecrease {
			newWorth = previousWorth - amount
		} else {
			amount = previousWorth + amount
		}
	}
	return &Mutation{
		ID:              uuid.New(),
		AccountID:       accID,
		ClassID:         classID,
		AssetID:         assetID,
		ChangeType:      changeType,
		Direction:       direction,
		Amount:          amount,
		PreviousWorth:   previousWorth,
		NewWorth:        newWorth,
		ClassTotalWorth: classTotalWorth,
		EffectiveDate:   effectiveDate,
		Note:            note,
		CreatedAt:       time.Now(),
	}, nil
}

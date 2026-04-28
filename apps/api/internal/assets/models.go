package assets

import (
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

// // Item is one concrete asset tracked inside a class.
// type Item struct {
// 	ID           uuid.UUID   `db:"id"`
// 	ClassID      uuid.UUID   `db:"class_id"`
// 	AccountID    uuid.UUID   `db:"account_id"`
// 	Name         string      `db:"name"`
// 	CurrentWorth money.Price `db:"current_worth"`
// 	Archived     bool        `db:"archived"`
// 	CreatedAt    time.Time   `db:"created_at"`
// 	UpdatedAt    time.Time   `db:"updated_at"`
// }

// // HistoryEntry captures one worth mutation for a tracked asset entry.
// type HistoryEntry struct {
// 	ID              uuid.UUID        `db:"id"`
// 	AccountID       uuid.UUID        `db:"account_id"`
// 	ClassID         uuid.UUID        `db:"class_id"`
// 	ItemID          uuid.UUID        `db:"item_id"`
// 	ChangeType      ChangeType       `db:"change_type"`
// 	Direction       *ChangeDirection `db:"direction"`
// 	Amount          money.Price      `db:"amount"`
// 	PreviousWorth   money.Price      `db:"previous_worth"`
// 	NewWorth        money.Price      `db:"new_worth"`
// 	ClassTotalWorth money.Price      `db:"class_total_worth"`
// 	EffectiveDate   time.Time        `db:"effective_date"`
// 	Note            string           `db:"note"`
// 	CreatedAt       time.Time        `db:"created_at"`
// }

// Snapshot stores one account-level total worth point for a specific day.
// type Snapshot struct {
// 	ID         uuid.UUID   `db:"id"`
// 	AccountID  uuid.UUID   `db:"account_id"`
// 	OccurredAt time.Time   `db:"occurred_at"`
// 	TotalWorth money.Price `db:"total_worth"`
// 	CreatedAt  time.Time   `db:"created_at"`
// 	UpdatedAt  time.Time   `db:"updated_at"`
// }

// ClassSummary is the high-level row shown in the classes table.
type ClassSummary struct {
	ID           uuid.UUID
	Name         string
	Source       ClassSource
	Archived     bool
	CurrentWorth money.Price
	LastChangeAt *time.Time
	GrowthPct    *float64
	UpdatedAt    time.Time
}

// ItemSummary represents one asset item inside a class.
type ItemSummary struct {
	ID           uuid.UUID
	Name         string
	CurrentWorth money.Price
	Archived     bool
	UpdatedAt    time.Time
}

// GrowthPoint represents class worth over time.
type GrowthPoint struct {
	Date       time.Time
	TotalWorth money.Price
}

// ClassDetails contains slider data for one class.
type ClassDetails struct {
	Class   ClassSummary
	Items   []ItemSummary
	Growth  []GrowthPoint
	History []Mutation
}

package assets

import (
	"time"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/money"
)

// Snapshot stores one account-level total worth point for a specific day.
type Snapshot struct {
	ID         uuid.UUID   `db:"id"`
	AccountID  uuid.UUID   `db:"account_id"`
	OccurredAt time.Time   `db:"occurred_at"`
	TotalWorth money.Price `db:"total_worth"`
	CreatedAt  time.Time   `db:"created_at"`
	UpdatedAt  time.Time   `db:"updated_at"`
}

package assets

import (
	"time"

	"github.com/google/uuid"
)

// Account is the portfolio-domain projection for a base account.
// It stores portfolio-specific state (e.g. build lock) and references account domain via AccountID.
type Account struct {
	ID        uuid.UUID `db:"id"`
	AccountID uuid.UUID `db:"account_id"`
	Building  bool      `db:"building"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewAccount(accountID uuid.UUID) *Account {
	return &Account{
		ID:        uuid.New(),
		AccountID: accountID,
		Building:  false,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

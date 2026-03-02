package cashflow

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Account is the cashflow-domain account projection.
type Account struct {
	ID        uuid.UUID `db:"id"`
	AccountID uuid.UUID `db:"account_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type AccountCreator interface {
	Create(ctx context.Context, acc *Account) error
}

func NewAccount(accountID uuid.UUID) *Account {
	return &Account{
		ID:        uuid.New(),
		AccountID: accountID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

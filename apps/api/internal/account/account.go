package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID         uuid.UUID  `db:"id"`
	ExternalID *uuid.UUID `db:"external_id"`
	Name       string     `db:"name"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

var (
	ErrAccountNotFound      = fmt.Errorf("account not found")
	ErrAccountAlreadyExists = fmt.Errorf("account already exists")
	ErrAccountNameRequired  = fmt.Errorf("account name is required")
)

type Creator interface {
	Create(ctx context.Context, acc *Account) error
}

type Lister interface {
	List(ctx context.Context) ([]*Account, error)
}

type Fetcher interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*Account, error)
}

func NewAccount(name string, externalID *uuid.UUID) (*Account, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, ErrAccountNameRequired
	}
	return &Account{
		ID:         uuid.New(),
		ExternalID: externalID,
		Name:       trimmed,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}, nil
}

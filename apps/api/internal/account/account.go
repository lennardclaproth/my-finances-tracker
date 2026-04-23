package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Account is the canonical account aggregate shared across domains.
type Account struct {
	ID         uuid.UUID  `db:"id"`
	ExternalID *uuid.UUID `db:"external_id"`
	Name       string     `db:"name"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

var (
	// ErrAccountNotFound indicates that the requested account does not exist.
	ErrAccountNotFound = fmt.Errorf("account not found")
	// ErrAccountAlreadyExists indicates that an account with the same unique key already exists.
	ErrAccountAlreadyExists = fmt.Errorf("account already exists")
	// ErrAccountNameRequired indicates that account name validation failed.
	ErrAccountNameRequired = fmt.Errorf("account name is required")
)

// Creator persists a new account.
type Creator interface {
	Create(ctx context.Context, acc *Account) error
}

// Lister returns all stored accounts.
type Lister interface {
	List(ctx context.Context) ([]*Account, error)
}

// Fetcher returns a single account by ID.
type Fetcher interface {
	FetchByID(ctx context.Context, id uuid.UUID) (*Account, error)
}

// NewAccount constructs a validated account instance with generated identity and timestamps.
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

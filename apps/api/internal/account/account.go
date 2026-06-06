package account

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Account is the canonical account aggregate shared across domains.
type Account struct {
	ID uuid.UUID `db:"id"`
	// ExternalID in the account domain is an optional identifier that can come from an external system (i.e. Google or EntraID)
	ExternalID *string   `db:"external_id"`
	Name       string    `db:"name"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// NewAccount constructs a validated account instance with generated identity and timestamps.
func NewAccount(name string, id *uuid.UUID, externalID *string) (*Account, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, ErrAccountNameRequired
	}

	resolvedID := uuid.New()
	if id != nil && *id != uuid.Nil {
		resolvedID = *id
	}

	now := time.Now().UTC()

	return &Account{
		ID:         resolvedID,
		ExternalID: externalID,
		Name:       trimmed,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

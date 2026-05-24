package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
)

type Commands struct {
	c creator
	b eventbus.Bus
}

type creator interface {
	Create(ctx context.Context, acc *Account) error
}

// Create persists an account and publishes an AccountCreated event when a
// publisher is configured.
//
// The id argument is optional. If id is nil or uuid.Nil, Create generates a new
// UUID for the account.
//
// The externalID argument is optional and can be used to store an identifier
// from an external system, such as Google or Entra ID.
//
// The name argument is required and must not be empty or whitespace.
func (c *Commands) Create(ctx context.Context, id *uuid.UUID, externalID *string, name string) (*uuid.UUID, error) {
	acc, err := NewAccount(name, id, externalID)
	if err != nil {
		return nil, fmt.Errorf("create: failed to create new account")
	}
	if err := c.c.Create(ctx, acc); err != nil {
		return nil, fmt.Errorf("create: failed to create account: %w", err)
	}
	c.b.Publish(ctx, TopicAccountCreated, AccountCreated{AccID: acc.ID})
	return &acc.ID, nil
}

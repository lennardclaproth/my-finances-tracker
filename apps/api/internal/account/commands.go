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

// Create persists the account and publishes an AccountCreated event when a publisher is configured.
func (c *Commands) Create(ctx context.Context, id, externalId *uuid.UUID, name string) error {
	acc, err := NewAccount(name, id, externalId)
	if err != nil {
		return fmt.Errorf("create: failed to create new account")
	}
	if err := c.c.Create(ctx, acc); err != nil {
		return fmt.Errorf("handlers: failed to create account: %w", err)
	}
	c.b.Publish(ctx, TopicAccountCreated, AccountCreated{AccID: acc.ID})
	return nil
}

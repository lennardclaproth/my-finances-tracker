package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	bus "github.com/lennardclaproth/my-finances-tracker/internal/messaging"
)

type CreateHandler struct {
	creator   creator
	publisher publisher
}

type publisher interface {
	Publish(ctx context.Context, env any) error
}

type creator interface {
	Create(ctx context.Context, acc *account.Account) error
}

// Create persists the account and publishes an AccountCreated event when a publisher is configured.
func (ch *CreateHandler) Create(ctx context.Context, id, externalId *uuid.UUID, name string) error {
	acc, err := account.NewAccount(name, id, externalId)
	if err := ch.creator.Create(ctx, acc); err != nil {
		return fmt.Errorf("handlers: failed to create account: %w", err)
	}
	if ch.publisher == nil {
		return nil
	}
	env, err := bus.NewJSONEnvelopeFromContext(ctx, api.AccountCreated{AccID: acc.ID})
	if err != nil {
		return fmt.Errorf("handlers: failed to encode account created event: %w", err)
	}
	if err := ch.publisher.Publish(ctx, env); err != nil {
		return fmt.Errorf("handlers: failed to publish account created event: %w", err)
	}
	return nil
}

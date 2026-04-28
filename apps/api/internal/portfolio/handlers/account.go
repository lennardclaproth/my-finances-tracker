package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

type AccountCreator interface {
	Create(ctx context.Context, acc *portfolio.Account) error
}

type AccountHandler struct {
	ac AccountCreator
}

func NewAccountHandler(ac AccountCreator) *AccountHandler {
	return &AccountHandler{
		ac: ac,
	}
}

func (h *AccountHandler) CreateAccount(ctx context.Context, accountID uuid.UUID) (*portfolio.Account, error) {
	acc := portfolio.NewAccount(accountID)
	if err := h.ac.Create(ctx, acc); err != nil {
		return nil, fmt.Errorf("handlers: CreateAccount failed to create account: %w", err)
	}
	return acc, nil
}

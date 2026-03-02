package cashflowhandlers

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/cashflow"
)

type AccountCreatedHandler struct {
	creator cashflow.AccountCreator
}

func NewAccountCreatedHandler(creator cashflow.AccountCreator) *AccountCreatedHandler {
	return &AccountCreatedHandler{
		creator: creator,
	}
}

func (h *AccountCreatedHandler) Handle(ctx context.Context, envelope bus.Envelope, e api.AccountCreated) error {
	acc := cashflow.NewAccount(e.AccID)
	return h.creator.Create(ctx, acc)
}

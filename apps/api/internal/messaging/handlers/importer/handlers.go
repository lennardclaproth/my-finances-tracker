package importerhandlers

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
)

type AccountCreatedHandler struct {
	creator importer.AccountCreator
}

func NewAccountCreatedHandler(creator importer.AccountCreator) *AccountCreatedHandler {
	return &AccountCreatedHandler{
		creator: creator,
	}
}

func (h *AccountCreatedHandler) Handle(ctx context.Context, envelope bus.Envelope, e api.AccountCreated) error {
	acc := importer.NewAccount(e.AccID)
	return h.creator.Create(ctx, acc)
}

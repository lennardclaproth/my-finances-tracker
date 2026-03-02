package portfoliohandlers

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

type TransactionsImportedHandler struct {
	pb portfolio.PortfolioBuilder
}

func NewTransactionsImportedHandler(pb portfolio.PortfolioBuilder) *TransactionsImportedHandler {
	return &TransactionsImportedHandler{
		pb: pb,
	}
}

func (h *TransactionsImportedHandler) Handle(ctx context.Context, envelope bus.Envelope, e api.TransactionsCreated) error {
	return h.pb.Build(ctx, e.AccID)
}

type PortfolioRebuildRequestedHandler struct {
	pb portfolio.PortfolioBuilder
}

func NewPortfolioRebuildRequestedHandler(pb portfolio.PortfolioBuilder) *PortfolioRebuildRequestedHandler {
	return &PortfolioRebuildRequestedHandler{
		pb: pb,
	}
}

func (h *PortfolioRebuildRequestedHandler) Handle(ctx context.Context, envelope bus.Envelope, e api.PortfolioRebuildRequested) error {
	return h.pb.Build(ctx, e.AccID)
}

type AccountCreatedHandler struct {
	creator portfolio.AccountCreator
}

func NewAccountCreatedHandler(creator portfolio.AccountCreator) *AccountCreatedHandler {
	return &AccountCreatedHandler{
		creator: creator,
	}
}

func (h *AccountCreatedHandler) Handle(ctx context.Context, envelope bus.Envelope, e api.AccountCreated) error {
	acc := portfolio.NewAccount(e.AccID)
	return h.creator.Create(ctx, acc)
}

package portfoliohandlers

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

type TransactionsImportedHandler struct {
	pb portfolio.PortfolioBuilder
	b  bus.Bus
	l  logging.Logger
}

func NewTransactionsImportedHandler(pb portfolio.PortfolioBuilder, b bus.Bus, l logging.Logger) *TransactionsImportedHandler {
	return &TransactionsImportedHandler{
		pb: pb,
		b:  b,
		l:  l,
	}
}

func (h *TransactionsImportedHandler) Handle(ctx context.Context, envelope bus.Envelope, e api.TransactionsCreated) error {
	if err := h.pb.Build(ctx, e.AccID); err != nil {
		return err
	}

	if h.b == nil {
		return nil
	}
	msg, err := bus.NewJSONEnvelopeFromContext(ctx, api.PortfolioRebuilt(e))
	if err != nil {
		h.l.Error(ctx, "failed to encode portfolio rebuilt event", err, "account_id", e.AccID.String())
		return nil
	}
	if err := h.b.Publish(ctx, msg); err != nil {
		h.l.Error(ctx, "failed to publish portfolio rebuilt event", err, "account_id", e.AccID.String())
	}
	return nil
}

type PortfolioRebuildRequestedHandler struct {
	pb portfolio.PortfolioBuilder
	b  bus.Bus
	l  logging.Logger
}

func NewPortfolioRebuildRequestedHandler(pb portfolio.PortfolioBuilder, b bus.Bus, l logging.Logger) *PortfolioRebuildRequestedHandler {
	return &PortfolioRebuildRequestedHandler{
		pb: pb,
		b:  b,
		l:  l,
	}
}

func (h *PortfolioRebuildRequestedHandler) Handle(ctx context.Context, envelope bus.Envelope, e api.PortfolioRebuildRequested) error {
	if err := h.pb.Build(ctx, e.AccID); err != nil {
		return err
	}

	if h.b == nil {
		return nil
	}
	msg, err := bus.NewJSONEnvelopeFromContext(ctx, api.PortfolioRebuilt(e))
	if err != nil {
		h.l.Error(ctx, "failed to encode portfolio rebuilt event", err, "account_id", e.AccID.String())
		return nil
	}
	if err := h.b.Publish(ctx, msg); err != nil {
		h.l.Error(ctx, "failed to publish portfolio rebuilt event", err, "account_id", e.AccID.String())
	}
	return nil
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

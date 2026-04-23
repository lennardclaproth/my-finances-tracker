package assetshandlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

type accountProjectionEnsurer interface {
	EnsureAccountProjection(ctx context.Context, accountID uuid.UUID) error
}

type portfolioSyncer interface {
	SyncPortfolioClassFromRebuild(ctx context.Context, accountID uuid.UUID) error
}

type snapshotRebuilder interface {
	RebuildTotalSnapshots(ctx context.Context, accountID uuid.UUID) error
}

// AccountCreatedHandler projects newly created accounts into assets accounts.
type AccountCreatedHandler struct {
	ensurer accountProjectionEnsurer
}

// NewAccountCreatedHandler constructs an account projection handler for assets.
func NewAccountCreatedHandler(ensurer accountProjectionEnsurer) *AccountCreatedHandler {
	return &AccountCreatedHandler{ensurer: ensurer}
}

// Handle creates the assets account projection for the event account id.
func (h *AccountCreatedHandler) Handle(ctx context.Context, _ bus.Envelope, e api.AccountCreated) error {
	return h.ensurer.EnsureAccountProjection(ctx, e.AccID)
}

// PortfolioRebuiltHandler syncs portfolio-linked asset worth after rebuild completion.
type PortfolioRebuiltHandler struct {
	syncer portfolioSyncer
}

// NewPortfolioRebuiltHandler constructs a portfolio-rebuilt sync handler.
func NewPortfolioRebuiltHandler(syncer portfolioSyncer) *PortfolioRebuiltHandler {
	return &PortfolioRebuiltHandler{syncer: syncer}
}

// Handle updates portfolio-linked class worth from the latest portfolio snapshot.
func (h *PortfolioRebuiltHandler) Handle(ctx context.Context, _ bus.Envelope, e api.PortfolioRebuilt) error {
	return h.syncer.SyncPortfolioClassFromRebuild(ctx, e.AccID)
}

// SnapshotsRebuildRequestedHandler rebuilds account-level assets snapshots.
type SnapshotsRebuildRequestedHandler struct {
	rebuilder snapshotRebuilder
	publisher bus.Bus
	log       logging.Logger
}

// NewSnapshotsRebuildRequestedHandler constructs a snapshots rebuild handler.
func NewSnapshotsRebuildRequestedHandler(rebuilder snapshotRebuilder, publisher bus.Bus, log logging.Logger) *SnapshotsRebuildRequestedHandler {
	return &SnapshotsRebuildRequestedHandler{
		rebuilder: rebuilder,
		publisher: publisher,
		log:       log,
	}
}

// Handle rebuilds account snapshots and publishes a rebuilt event.
func (h *SnapshotsRebuildRequestedHandler) Handle(ctx context.Context, _ bus.Envelope, e api.AssetsSnapshotsRebuildRequested) error {
	if err := h.rebuilder.RebuildTotalSnapshots(ctx, e.AccID); err != nil {
		return err
	}
	if h.publisher == nil {
		return nil
	}
	env, err := bus.NewJSONEnvelopeFromContext(ctx, api.AssetsSnapshotsRebuilt(e))
	if err != nil {
		h.log.Error(ctx, "failed to encode assets snapshots rebuilt event", err, "account_id", e.AccID.String())
		return nil
	}
	if err := h.publisher.Publish(ctx, env); err != nil {
		h.log.Error(ctx, "failed to publish assets snapshots rebuilt event", err, "account_id", e.AccID.String())
	}
	return nil
}

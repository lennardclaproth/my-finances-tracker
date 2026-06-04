// Package assets holds event handlers that keep account-level assets state in
// sync with the events that affect it.
package assets

import (
	"context"

	"github.com/google/uuid"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

// PortfolioRebuiltHandler reacts to a portfolio rebuild by re-syncing the
// portfolio-linked assets class from the fresh portfolio snapshots and then
// rebuilding account-level assets snapshots, announcing the result so the
// realtime layer can notify clients.
type PortfolioRebuiltHandler struct {
	syncer  *assets.Syncer
	builder *assets.Builder
	bus     eventbus.Bus
	logger  logging.Logger
}

// NewPortfolioRebuiltHandler constructs a PortfolioRebuiltHandler. The bus may
// be nil, in which case the SnapshotsRebuilt event is not published.
func NewPortfolioRebuiltHandler(
	syncer *assets.Syncer,
	builder *assets.Builder,
	bus eventbus.Bus,
	logger logging.Logger,
) *PortfolioRebuiltHandler {
	return &PortfolioRebuiltHandler{syncer: syncer, builder: builder, bus: bus, logger: logger}
}

// Handle syncs the portfolio-linked class and rebuilds assets snapshots for the
// account, then publishes assets.SnapshotsRebuilt on success.
func (h *PortfolioRebuiltHandler) Handle(ctx context.Context, evt portfolio.Rebuilt, _ eventbus.Metadata) error {
	if err := h.syncer.SyncPortfolio(ctx, evt.AccID); err != nil {
		if h.logger != nil {
			h.logger.Error(ctx, "assets portfolio sync failed", err, "account_id", evt.AccID.String())
		}
		return err
	}
	if err := h.builder.RebuildAll(ctx, evt.AccID); err != nil {
		if h.logger != nil {
			h.logger.Error(ctx, "assets snapshots rebuild failed", err, "account_id", evt.AccID.String())
		}
		return err
	}
	publishSnapshotsRebuilt(ctx, h.bus, evt.AccID)
	return nil
}

// publishSnapshotsRebuilt announces that account-level assets snapshots were
// rebuilt. It is a no-op when no bus is configured. Shared by the handlers that
// rebuild assets snapshots.
func publishSnapshotsRebuilt(ctx context.Context, bus eventbus.Bus, accID uuid.UUID) {
	if bus == nil {
		return
	}
	_ = bus.Publish(ctx, assets.TopicSnapshotsRebuilt, assets.SnapshotsRebuilt{AccID: accID})
}

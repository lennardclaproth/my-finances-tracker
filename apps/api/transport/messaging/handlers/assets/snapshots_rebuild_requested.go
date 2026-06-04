package assets

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

// SnapshotsRebuildRequestedHandler rebuilds account-level assets snapshots in
// response to a rebuild request (published by the assets write side after a
// class or asset mutation) and announces completion so the realtime layer can
// notify clients.
type SnapshotsRebuildRequestedHandler struct {
	builder *assets.Builder
	bus     eventbus.Bus
	logger  logging.Logger
}

// NewSnapshotsRebuildRequestedHandler constructs a SnapshotsRebuildRequestedHandler.
// The bus may be nil, in which case the SnapshotsRebuilt event is not published.
func NewSnapshotsRebuildRequestedHandler(
	builder *assets.Builder,
	bus eventbus.Bus,
	logger logging.Logger,
) *SnapshotsRebuildRequestedHandler {
	return &SnapshotsRebuildRequestedHandler{builder: builder, bus: bus, logger: logger}
}

// Handle rebuilds the account's assets snapshots, then publishes
// assets.SnapshotsRebuilt on success.
func (h *SnapshotsRebuildRequestedHandler) Handle(ctx context.Context, evt assets.SnapshotsRebuildRequested, _ eventbus.Metadata) error {
	if err := h.builder.RebuildAll(ctx, evt.AccID); err != nil {
		if h.logger != nil {
			h.logger.Error(ctx, "assets snapshots rebuild failed", err, "account_id", evt.AccID.String())
		}
		return err
	}
	publishSnapshotsRebuilt(ctx, h.bus, evt.AccID)
	return nil
}

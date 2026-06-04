package notify

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

const (
	// EventPortfolioRebuilt notifies clients that portfolio projections were rebuilt.
	EventPortfolioRebuilt = "portfolio.rebuilt"
	// EventImportCompleted notifies clients that an import finished successfully.
	EventImportCompleted = "import.completed"
	// EventAssetsRebuilt notifies clients that assets snapshots were rebuilt.
	EventAssetsRebuilt = "assets.rebuilt"
)

// ImportCompletedHandler forwards import completion events to websocket clients.
type ImportCompletedHandler struct {
	hub *Hub
}

// NewImportCompletedHandler creates an ImportCompletedHandler.
func NewImportCompletedHandler(hub *Hub) *ImportCompletedHandler {
	return &ImportCompletedHandler{hub: hub}
}

// Handle pushes an import-completed websocket notification for the target
// account. Imports without an account scope (such as EOD imports) are ignored.
func (h *ImportCompletedHandler) Handle(ctx context.Context, evt importer.Completed, _ eventbus.Metadata) error {
	if h.hub == nil || evt.AccountID == nil {
		return nil
	}
	h.hub.NotifyDataChanged(ctx, *evt.AccountID, EventImportCompleted)
	return nil
}

// PortfolioRebuiltHandler forwards portfolio rebuilt events to websocket clients.
type PortfolioRebuiltHandler struct {
	hub *Hub
}

// NewPortfolioRebuiltHandler creates a PortfolioRebuiltHandler.
func NewPortfolioRebuiltHandler(hub *Hub) *PortfolioRebuiltHandler {
	return &PortfolioRebuiltHandler{hub: hub}
}

// Handle pushes a portfolio-rebuilt websocket notification for the target account.
func (h *PortfolioRebuiltHandler) Handle(ctx context.Context, evt portfolio.Rebuilt, _ eventbus.Metadata) error {
	if h.hub == nil {
		return nil
	}
	h.hub.NotifyDataChanged(ctx, evt.AccID, EventPortfolioRebuilt)
	return nil
}

// AssetsSnapshotsRebuiltHandler forwards assets snapshot rebuild events to websocket clients.
type AssetsSnapshotsRebuiltHandler struct {
	hub *Hub
}

// NewAssetsSnapshotsRebuiltHandler creates an AssetsSnapshotsRebuiltHandler.
func NewAssetsSnapshotsRebuiltHandler(hub *Hub) *AssetsSnapshotsRebuiltHandler {
	return &AssetsSnapshotsRebuiltHandler{hub: hub}
}

// Handle pushes an assets-rebuilt websocket notification for the target account.
func (h *AssetsSnapshotsRebuiltHandler) Handle(ctx context.Context, evt assets.SnapshotsRebuilt, _ eventbus.Metadata) error {
	if h.hub == nil {
		return nil
	}
	h.hub.NotifyDataChanged(ctx, evt.AccID, EventAssetsRebuilt)
	return nil
}

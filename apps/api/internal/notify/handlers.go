package notify

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/api"
	"github.com/lennardclaproth/my-finances-tracker/internal/bus"
)

const (
	// EventPortfolioRebuilt notifies clients that portfolio projections were rebuilt.
	EventPortfolioRebuilt = "portfolio.rebuilt"
	// EventImportCompleted notifies clients that an import finished successfully.
	EventImportCompleted = "import.completed"
	// EventBulkTagCompleted notifies clients that async bulk tagging completed.
	EventBulkTagCompleted = "bulk_tag.completed"
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

// Handle pushes an import completed websocket notification for the target account.
func (h *ImportCompletedHandler) Handle(ctx context.Context, _ bus.Envelope, e api.ImportCompleted) error {
	if h.hub != nil {
		h.hub.NotifyDataChanged(ctx, e.AccID, EventImportCompleted)
	}
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

// Handle pushes a portfolio rebuilt websocket notification for the target account.
func (h *PortfolioRebuiltHandler) Handle(ctx context.Context, _ bus.Envelope, e api.PortfolioRebuilt) error {
	if h.hub != nil {
		h.hub.NotifyDataChanged(ctx, e.AccID, EventPortfolioRebuilt)
	}
	return nil
}

// BulkTagCompletedHandler forwards bulk-tag completion events to websocket clients.
type BulkTagCompletedHandler struct {
	hub *Hub
}

// NewBulkTagCompletedHandler creates a BulkTagCompletedHandler.
func NewBulkTagCompletedHandler(hub *Hub) *BulkTagCompletedHandler {
	return &BulkTagCompletedHandler{hub: hub}
}

// Handle pushes a bulk tag completed websocket notification for the target account.
func (h *BulkTagCompletedHandler) Handle(ctx context.Context, _ bus.Envelope, e api.BulkTagCompleted) error {
	if h.hub != nil {
		h.hub.NotifyDataChanged(ctx, e.AccID, EventBulkTagCompleted)
	}
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

// Handle pushes an assets rebuilt websocket notification for the target account.
func (h *AssetsSnapshotsRebuiltHandler) Handle(ctx context.Context, _ bus.Envelope, e api.AssetsSnapshotsRebuilt) error {
	if h.hub != nil {
		h.hub.NotifyDataChanged(ctx, e.AccID, EventAssetsRebuilt)
	}
	return nil
}

// Package importer holds event handlers that react to importer domain events.
package importer

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

// AcceptedHandler reacts to importer.Accepted by running the persisted import
// through the importer write side. It replaces the removed background import
// job: without a subscriber on importer.TopicAccepted, accepted uploads would
// never be processed and would remain pending forever.
type AcceptedHandler struct {
	commands *importer.Commands
	logger   logging.Logger
}

// NewAcceptedHandler constructs an AcceptedHandler that delegates to the
// importer commands.
func NewAcceptedHandler(commands *importer.Commands, logger logging.Logger) *AcceptedHandler {
	return &AcceptedHandler{commands: commands, logger: logger}
}

// Handle processes the accepted import. Processing advances the import
// lifecycle and publishes importer.Completed or importer.Failed.
func (h *AcceptedHandler) Handle(ctx context.Context, evt importer.Accepted, _ eventbus.Metadata) error {
	if err := h.commands.Process(ctx, evt.ImportID); err != nil {
		if h.logger != nil {
			h.logger.Error(ctx, "import processing failed", err, "import_id", evt.ImportID.String())
		}
		return err
	}
	return nil
}

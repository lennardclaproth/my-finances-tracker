package assets

import (
	"context"

	"github.com/lennardclaproth/my-finances-tracker/internal/account"
	"github.com/lennardclaproth/my-finances-tracker/internal/assets"
	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
)

// AccountCreatedHandler creates the assets projection for a new account. The
// projection holds assets-specific per-account state (the sync lock), so it
// must exist before assets snapshots can be built for the account.
type AccountCreatedHandler struct {
	commands *assets.Commands
	logger   logging.Logger
}

// NewAccountCreatedHandler constructs an AccountCreatedHandler.
func NewAccountCreatedHandler(commands *assets.Commands, logger logging.Logger) *AccountCreatedHandler {
	return &AccountCreatedHandler{commands: commands, logger: logger}
}

// Handle creates the assets projection for the newly created account.
func (h *AccountCreatedHandler) Handle(ctx context.Context, evt account.Created, _ eventbus.Metadata) error {
	if _, err := h.commands.CreateAccount(ctx, evt.AccID); err != nil {
		if h.logger != nil {
			h.logger.Error(ctx, "assets account projection creation failed", err, "account_id", evt.AccID.String())
		}
		return err
	}
	return nil
}

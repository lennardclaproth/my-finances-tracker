// Package portfolio holds event handlers that trigger portfolio rebuilds.
package portfolio

import (
	"context"
	"errors"

	"github.com/lennardclaproth/my-finances-tracker/internal/eventbus"
	"github.com/lennardclaproth/my-finances-tracker/internal/importer"
	"github.com/lennardclaproth/my-finances-tracker/internal/logging"
	"github.com/lennardclaproth/my-finances-tracker/internal/portfolio"
)

// ImportCompletedHandler rebuilds an account's portfolio after a portfolio
// import completes. A successful rebuild publishes portfolio.Rebuilt, which in
// turn drives the assets sync. Other import types are ignored here: cashflow
// transactions are queried directly, and EOD revaluation is not yet wired
// because import.Completed carries the listing, not the affected accounts.
type ImportCompletedHandler struct {
	builder *portfolio.Builder
	logger  logging.Logger
}

// NewImportCompletedHandler constructs an ImportCompletedHandler.
func NewImportCompletedHandler(builder *portfolio.Builder, logger logging.Logger) *ImportCompletedHandler {
	return &ImportCompletedHandler{builder: builder, logger: logger}
}

// Handle rebuilds the portfolio for the account referenced by a completed
// portfolio import.
func (h *ImportCompletedHandler) Handle(ctx context.Context, evt importer.Completed, _ eventbus.Metadata) error {
	if evt.Type != importer.ImportTypePortfolio || evt.AccountID == nil {
		return nil
	}
	// A portfolio import with no buildable positions yields no snapshots; that is
	// an expected outcome, not a failure, so it is not surfaced as an error.
	if err := h.builder.Build(ctx, *evt.AccountID); err != nil && !errors.Is(err, portfolio.ErrPortfolioNoSnapshots) {
		if h.logger != nil {
			h.logger.Error(ctx, "portfolio rebuild after import failed", err,
				"account_id", evt.AccountID.String(), "import_id", evt.ImportID.String())
		}
		return err
	}
	return nil
}

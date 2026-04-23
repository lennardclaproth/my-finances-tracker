## [025] Portfolio-Linked Asset Class Synchronization

### Summary
Adds a portfolio-linked, read-only asset class projection in the assets domain. Synchronization runs only on portfolio rebuild completion events and rebuilds class worth history from the portfolio snapshot timeline.
It now also triggers account-level assets snapshot projection rebuilds so the total-assets graph endpoint stays aligned after portfolio rebuilds.

### Why This Feature Exists
Users want total portfolio worth represented alongside manually tracked asset classes, without allowing manual edits to that portfolio-derived class.

### Synchronization Contract
1. Event trigger
- Source event: `PortfolioRebuilt.v1`
- On event handling, the assets domain reads portfolio snapshots for that account.

2. Projection behavior
- Ensures a fixed portfolio class exists (`source=PORTFOLIO`, name `Portfolio`).
- Ensures a fixed portfolio item exists (`Portfolio Worth`).
- Uses `portfolio_snapshots.market_value` as the source-of-truth worth and sets current portfolio item worth to the latest snapshot market value.
- Rebuilds class history entries from snapshots (one entry per snapshot day), so growth chart points match portfolio snapshot history.
- Replaces prior portfolio-class history during sync to keep the projection deterministic and aligned with source snapshots.
- After portfolio-class sync, publishes an assets snapshots rebuild request so account-level totals are recomputed asynchronously.

3. Read-only constraints
- Portfolio class is not mutable via manual class/item workflows.
- Manual update/delete APIs return validation errors when targeting non-manual classes.

### Event and Handler Wiring
1. Account projection
- `AccountCreated.v1` handler ensures assets account projection exists.

2. Portfolio sync
- `PortfolioRebuilt.v1` handler invokes assets sync service.
- Assets sync then publishes `AssetsSnapshotsRebuildRequested.v1`; a dedicated handler rebuilds snapshots and publishes `AssetsSnapshotsRebuilt.v1`.
- Realtime notifies web clients with `data_changed.event = "assets.rebuilt"` when the snapshot rebuild completes.

### Operational Rules
- Sync runs only on explicit portfolio rebuild completion; no background polling.
- If no snapshots exist, sync is a no-op.
- Sync remains account-scoped and does not cross accounts.

### Code References
- Assets event handlers:
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/messaging/handlers/assets/handlers.go)
- Assets sync service logic:
  - [service.go](C:/personal/git/my-finances-tracker/apps/api/internal/assets/service.go)
- Bus subscription wiring:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)

### Validation Coverage
- Integration test validates that publishing `PortfolioRebuilt` creates/syncs portfolio asset class worth.
  - [assets_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/assets_integration_test.go)
- Integration test also validates account-level assets snapshots reflect portfolio-linked totals after rebuild completion.
  - [assets_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/assets_integration_test.go)

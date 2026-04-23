## [001] WebSocket Client Updates

### Summary
The API provides account-scoped realtime notifications over WebSockets so connected clients can refresh cashflow and portfolio data when backend processing completes.

### Why This Feature Exists
Long-running operations (imports, portfolio rebuilds, async bulk tagging) finish asynchronously. Without push notifications, users must poll manually to know when fresh data is ready.

### Backend Contract
- WebSocket endpoint: `GET /ws/accounts/{account_id}`
- Connection scope: every socket is bound to exactly one `account_id`
- Fanout model: multiple simultaneous clients/tabs can connect for the same account and all receive notifications
- Message shape:

```json
{
  "type": "data_changed",
  "event": "portfolio.rebuilt | import.completed | bulk_tag.completed",
  "account_id": "<uuid>",
  "timestamp": "<RFC3339 timestamp>"
}
```

### Notification Sources (Event Bus -> WebSocket)
The websocket layer subscribes to internal completion events and maps them to websocket events:

- `ImportCompleted.v1` -> `import.completed`
- `PortfolioRebuilt.v1` -> `portfolio.rebuilt`
- `BulkTagCompleted.v1` -> `bulk_tag.completed`

### Completion Event Publishing
Notifications are emitted only after successful completion points:

- Import job publishes `ImportCompleted` after import state is persisted as completed.
- Portfolio messaging handlers publish `PortfolioRebuilt` after `pb.Build(...)` succeeds.
- Bulk tag async job publishes `BulkTagCompleted` after `UpdateTagByQuery(...)` succeeds.

### Connection Lifecycle and Staleness
The hub enforces two independent stale/dispose rules:

1. Ping/pong health
- Ping interval: 10 seconds (during no-update periods)
- If no pong is received for 3 consecutive ping attempts, the socket is closed.

2. No-update idle staleness
- If no `data_changed` update message is sent for 5 minutes, the socket is closed with:
  - close code: `4001`
  - reason: `idle_no_updates`
- This timer is reset by update messages, not by ping/pong traffic.

### Frontend Behavior
Frontend uses a singleton realtime client that binds to the current active account socket and dispatches parsed notifications to subscribers.

Reconnect behavior:
- Normal disconnects: automatic reconnect with bounded backoff (`500ms -> 1s -> 2s -> 5s -> 10s`).
- Idle stale close (`4001`): no automatic reconnect.
- Reconnect after `4001`: triggered on first user interaction (`click` or `pointerdown`).
- Unexpected websocket payloads and transport errors are ignored for UX continuity but logged with warnings for diagnosis.

### Page-Level Refresh Integration
- Cashflow page refreshes transactions + analytics on:
  - `import.completed`
  - `bulk_tag.completed`
- Portfolio page refreshes snapshots + positions + transactions on:
  - `portfolio.rebuilt`
  - `import.completed`
- Refreshes are coalesced with a short debounce to avoid burst refetch storms.

### Important Compatibility Notes
- Existing HTTP endpoints and successful response contracts remain unchanged.
- Async bulk tag notifications require `account_id` in tag-filter requests for routing; if missing, large operations fall back to synchronous processing.
- This implementation is process-local (in-memory hub), so cross-instance fanout is not provided.
- WebSocket upgrade remains compatible with HTTP middleware wrappers because the response writer used for request logging preserves connection hijacking.
- APM server transactions ignore websocket handshake requests on `/ws/accounts/{account_id}` to avoid high-cardinality trace noise.

### Code References
- API route wiring and subscriptions:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)
- WebSocket hub implementation:
  - [hub.go](C:/personal/git/my-finances-tracker/apps/api/internal/notify/hub.go)
- Bus->WebSocket event mapping:
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/notify/handlers.go)
- Event definitions:
  - [events.go](C:/personal/git/my-finances-tracker/apps/api/api/events.go)
- Completion publishers:
  - [import.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/import.go)
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/messaging/handlers/portfolio/handlers.go)
  - [bulk_tag.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/bulk_tag.go)
- Frontend realtime client and integrations:
  - [realtime.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/realtime.ts)
  - [CashflowTransactionsPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/CashflowTransactionsPage.vue)
  - [PortfolioPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/PortfolioPage.vue)

### Validation Coverage
Backend unit tests cover:
- idle close with `4001` after no updates,
- update-message timer reset,
- missed-pong disposal behavior.

See [hub_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/notify/hub_test.go).

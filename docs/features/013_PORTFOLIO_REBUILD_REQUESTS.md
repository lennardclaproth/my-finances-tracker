## [013] Portfolio Rebuild Requests via Event Bus

### Summary
`POST /portfolio/rebuild` is an asynchronous command endpoint that publishes a rebuild intent on the internal event bus, with actual portfolio recomputation handled by subscribed background/message handlers.

### Why This Feature Exists
Portfolio rebuilds can be expensive and should not block HTTP request lifecycles; event-driven execution keeps the API responsive while preserving traceable command intent.

### HTTP Contract
- Endpoint: `POST /portfolio/rebuild`
- Request body: `{ account_id: UUID }`
- Success: `202 Accepted` with `api.AsyncEventAcceptedResponse`
  - `message_id`, `topic`, `account_id`
- Error responses:
  - `400` for invalid payload (for example empty `account_id`)
  - `404` when account does not exist
  - `503` when event bus is unavailable or publish fails
  - `500` for unexpected server errors

### Request Handling Semantics
1. Preconditions
- Event bus publisher must be configured.
- Account existence is verified before publishing.

2. Publish behavior
- Handler publishes `api.PortfolioRebuildRequested` as JSON envelope.
- Topic is `PortfolioRebuildRequested.v1`.
- API returns accepted metadata from the created envelope.

### Asynchronous Execution Pipeline
1. Bus subscription
- Server wiring subscribes a `PortfolioRebuildRequestedHandler` to `PortfolioRebuildRequested.v1`.

2. Rebuild execution
- Handler calls `PortfolioBuilder.Build(account_id)`.
- Build flow uses account-level build lock (`TryAcquireBuildLock` / `ReleaseBuildLock`) to avoid overlapping rebuilds.
- Builder computes positions, position snapshots, and then portfolio snapshots.

3. Completion event
- On successful build, handler publishes `PortfolioRebuilt.v1`.
- Realtime notify bridge maps this to websocket event `portfolio.rebuilt` for clients.

### Frontend Interaction Pattern
- Client calls `requestPortfolioRebuild(accountId)` and immediately shows a queued-background message.
- Portfolio page listens for `portfolio.rebuilt` (and `import.completed`) and performs debounced refresh of snapshots, positions, and transactions.

### Error Behavior and Reliability Notes
- Missing bus dependency is surfaced as `503` instead of silently dropping requests.
- Publish failure returns `503` with explicit message.
- Build lock release is attempted with a background timeout context to reduce risk of stale lock on canceled request contexts.

### Important Notes
- Endpoint is command-oriented: it confirms enqueue/publication, not rebuild completion.
- Rebuild completion visibility is event-driven (`portfolio.rebuilt`), not immediate in HTTP response.
- Async rebuild request/accepted-response DTO contracts are documented in code.

### Code References
- Rebuild HTTP handler:
  - [portfolio.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/portfolio.go)
- Bus subscription and composition wiring:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)
- Rebuild message handler:
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/messaging/handlers/portfolio/handlers.go)
- Portfolio builder and lock behavior:
  - [portfolio_builder.go](C:/personal/git/my-finances-tracker/apps/api/internal/portfolio/portfolio_builder.go)
  - [portfolio.go](C:/personal/git/my-finances-tracker/apps/api/internal/portfolio/portfolio.go)
- Event definitions and realtime notify mapping:
  - [events.go](C:/personal/git/my-finances-tracker/apps/api/api/events.go)
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/notify/handlers.go)
- Frontend rebuild request + realtime refresh usage:
  - [portfolio.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/portfolio.ts)
  - [PortfolioPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/PortfolioPage.vue)

### Validation Coverage
- Handler tests cover accepted publish path, unknown account (`404`), publish failure (`503`), and invalid request validation (`400`).
  - [portfolio_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/portfolio_test.go)
- Portfolio builder tests cover lock-acquire failure and in-progress/no-snapshot failure paths.
  - [portfolio_builder_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/portfolio/portfolio_builder_test.go)

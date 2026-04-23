## [017] Event-Driven Messaging Handlers

### Summary
The application uses typed internal bus handlers to project account events, trigger portfolio rebuild flows, and bridge domain-completion events into realtime websocket notifications.

### Why This Feature Exists
Event-driven handlers decouple write commands from downstream module updates so account setup, portfolio recomputation, and realtime UI refresh can evolve independently.

### Bus Wiring and Delivery Model
- Composition root creates an in-memory bus with worker pool and bounded queue.
- Handlers are subscribed by topic in `setupBus(...)` and `setupRealtimeNotifications(...)`.
- Messages are decoded through `bus.DecodeHandler` using codec registry (`json`).

### Core Domain Handler Subscriptions
1. Account projection fanout (`AccountCreated.v1`)
- `portfoliohandlers.AccountCreatedHandler` -> creates `portfolio` account projection.
- `cashflowhandlers.AccountCreatedHandler` -> creates `cashflow` account projection.
- `importerhandlers.AccountCreatedHandler` -> creates `import` account projection.

2. Portfolio rebuild triggers
- `TransactionsCreated.v1` -> `portfoliohandlers.TransactionsImportedHandler`
  - builds portfolio
  - publishes `PortfolioRebuilt.v1` on success
- `PortfolioRebuildRequested.v1` -> `portfoliohandlers.PortfolioRebuildRequestedHandler`
  - builds portfolio
  - publishes `PortfolioRebuilt.v1` on success

### Realtime Bridge Subscriptions
- `ImportCompleted.v1` -> notify handler emits websocket `import.completed`
- `PortfolioRebuilt.v1` -> notify handler emits websocket `portfolio.rebuilt`
- `BulkTagCompleted.v1` -> notify handler emits websocket `bulk_tag.completed`

These messages are account-scoped fanout signals for frontend refresh logic.

### Envelope/Trace Propagation Behavior
- Publish path (`NewJSONEnvelopeFromContext`) propagates:
  - request/correlation IDs
  - W3C trace headers (`traceparent`/`tracestate`)
- Consume path (`DecodeHandler`) restores propagation headers into context and starts consumer APM transactions.
- If correlation/request IDs are absent in context, envelope IDs are used as fallback identifiers.

### Failure and Backpressure Notes
- Bus backpressure policy is configured as `BackpressureDrop` in composition root.
- Handler publish failures (for follow-up events) are logged; handler execution itself can still return success.
- Subscription setup failures abort startup path with explicit errors.

### Important Notes
- Messaging handlers in `internal/messaging/handlers/*` currently have no dedicated unit test files.
- Realtime notify handlers are separate from domain handlers but are wired from the same bus topics.
- Exported event payload/topic contracts are documented in code so topic-to-payload intent stays explicit as handlers evolve.

### Code References
- Bus interface, typed decode, and envelope context propagation:
  - [bus.go](C:/personal/git/my-finances-tracker/apps/api/internal/bus/bus.go)
  - [codec.go](C:/personal/git/my-finances-tracker/apps/api/internal/bus/codec.go)
  - [envelope.go](C:/personal/git/my-finances-tracker/apps/api/internal/bus/envelope.go)
- In-memory bus implementation:
  - [memory_bus.go](C:/personal/git/my-finances-tracker/apps/api/internal/bus/memory/memory_bus.go)
- Domain messaging handlers:
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/messaging/handlers/portfolio/handlers.go)
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/messaging/handlers/cashflow/handlers.go)
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/messaging/handlers/importer/handlers.go)
- Realtime notify handlers:
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/notify/handlers.go)
- Subscription wiring:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)
- Event types/topics:
  - [events.go](C:/personal/git/my-finances-tracker/apps/api/api/events.go)

### Validation Coverage
- Bus envelope/trace propagation behavior is covered by tests for header propagation and trace continuation across decode handler execution.
  - [envelope_trace_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/bus/envelope_trace_test.go)

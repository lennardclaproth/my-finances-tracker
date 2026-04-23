## [018] Structured Observability Foundation

### Summary
Observability is implemented through request/correlation ID propagation, allowlisted structured logging, and APM trace continuity across HTTP, bus, and background jobs.

### Why This Feature Exists
The system mixes synchronous HTTP handling with async bus/jobs; consistent identifiers and trace propagation are required for end-to-end debugging and incident analysis.

### Request and Correlation Identifiers
- `WithRequestIdentifiers()` middleware:
  - reads inbound `X-Request-ID` and `X-Correlation-ID`
  - generates missing values
  - writes both headers back on response
  - stores IDs in request context
- If both are missing, correlation defaults to request ID.

### Structured Logging Model
- Logger implementation (`SlogLogger`) appends context fields automatically.
- Log fields are filtered through an explicit allowlist to prevent accidental high-cardinality or sensitive labels.
- Request logging middleware emits normalized fields:
  - method, path, status, bytes, duration_ms
  - operation, component, outcome
  - request/correlation and trace IDs when available
- Request logging avoids duplicating context identifiers by emitting them exactly once per log entry.
- APM HTTP middleware skips websocket upgrade requests (`GET /ws/accounts/{account_id}`) to reduce low-value, per-connection transaction volume.

### APM Trace Integration
1. HTTP entry
- Server wraps mux with `apmhttp.Wrap`, creating HTTP transaction context.

2. Async propagation
- Context propagation headers are captured on enqueue/publish boundaries.
- Bus envelopes carry request/correlation and trace headers.

3. Consumer/job continuation
- Bus decode handler and jobs start transactions from incoming trace headers.
- Parse/persist/publish spans are created in long-running flows (import, daily upload, bulk tag, tagger).

### Operation Naming Convention
Operation helpers normalize names consistently:
- HTTP: `http.<method>.<route>`
- Jobs: `job.<job_name>.process`
- Bus consume/publish: `bus.consume.<topic>` / `bus.publish.<topic>`

### Safety and Redaction Controls
- `FilterFields(...)` removes non-allowlisted keys before log emission.
- `SetSafeTransactionLabels(...)` applies only allowlisted labels to APM transactions.
- Correlation IDs are convertible to UUID (fallback deterministic SHA1 UUID when non-UUID strings are used).

### Important Notes
- Observability fields and labels are intentionally strict; non-allowlisted diagnostic keys are dropped.
- Trace continuation from malformed incoming headers still creates a root transaction while surfacing parse errors.
- Request/correlation middleware contracts are documented in code, and `/health` plus `/swagger/` now run through request-logging middleware.

### Code References
- Context identifiers and header constants:
  - [context.go](C:/personal/git/my-finances-tracker/apps/api/internal/observability/context.go)
- Trace/operation helpers and transaction continuation:
  - [trace.go](C:/personal/git/my-finances-tracker/apps/api/internal/observability/trace.go)
- Field/label allowlist filtering:
  - [redaction.go](C:/personal/git/my-finances-tracker/apps/api/internal/observability/redaction.go)
- Request identifier and request logging middleware:
  - [middlewares.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/middlewares.go)
- Server APM wrapping:
  - [server.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/server.go)
- Logger implementation:
  - [slog_logger.go](C:/personal/git/my-finances-tracker/apps/api/internal/logging/slog_logger.go)
- Bus propagation integration:
  - [bus.go](C:/personal/git/my-finances-tracker/apps/api/internal/bus/bus.go)
  - [envelope.go](C:/personal/git/my-finances-tracker/apps/api/internal/bus/envelope.go)

### Validation Coverage
- Middleware tests verify request/correlation generation/preservation and request-log context fields.
  - [middlewares_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/middlewares_test.go)
- Context and redaction tests verify ID behavior and allowlist filtering.
  - [context_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/observability/context_test.go)
  - [redaction_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/observability/redaction_test.go)
- Envelope/decoder tests verify trace and correlation propagation across bus boundaries.
  - [envelope_trace_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/bus/envelope_trace_test.go)
- Server test verifies configured timeout hardening.
  - [server_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/server_test.go)

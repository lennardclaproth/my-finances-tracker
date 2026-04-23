## [007] Daily Market Data Retrieval

### Summary
The API exposes `GET /marketdata/dailies` to retrieve listing daily price history with optional date filters, pagination, and sort order, resolved by either `listing_id` or `symbol`.

### Why This Feature Exists
Portfolio and admin screens need a consistent way to read historical daily market data while the system may still be synchronizing or using manual providers.

### HTTP Contract
- Endpoint: `GET /marketdata/dailies`
- Query parameters:
  - `listing_id` (optional UUID, takes precedence when provided)
  - `symbol` (required only when `listing_id` is omitted)
  - `from`, `to` (optional `YYYY-MM-DD`)
  - `sort_order` (`asc` or `desc`, default `asc`)
  - `limit` (default `100` when omitted)
  - `offset`
- Success response: `200 OK` with `marketdata.DailyResponse`

### Lookup Resolution Rules
1. If `listing_id` is present and valid UUID, retrieval is by listing ID.
2. Otherwise retrieval is by `symbol`.
3. If both are supplied, `listing_id` path is used.

### Request Validation Behavior
- `symbol` is required when `listing_id` is not provided.
- `listing_id` must be a valid UUID when provided.
- `from` and `to` must parse as `YYYY-MM-DD`.
- `sort_order` accepts only `asc`/`desc`.
- `limit` and `offset` must be non-negative.

### Response Shape and Metadata Semantics
Current response shape is struct-driven (`DailyResponse`) with PascalCase keys:
- `Data`: daily rows
- `Metadata`:
  - `Message`
  - `ResultCount`
  - `TotalCount`

`Metadata.Message` indicates runtime freshness/context:
- `Manual provider configured; automatic sync disabled`
- `Data may be stale, listing is currently syncing`
- `Sync triggered; data may be stale until sync is complete`

### Runtime Behavior by Provider/Sync State
1. Manual provider source
- No automatic sync is triggered.
- Data is read directly from store and returned with manual-provider metadata message.

2. Listing currently syncing
- Existing stored data is returned immediately.
- Metadata indicates possible staleness.

3. Stale/non-accumulated listing
- Service marks listing for accumulation and starts async sync in background.
- Current stored data is returned while sync continues.

### Sorting and Pagination
- Sort order is normalized to `asc`/`desc` and applied in store query.
- `limit` and `offset` are passed through to store queries.
- Store also computes `TotalCount` for metadata/pagination awareness.

### Error Behavior
- Invalid query input maps to `400` with field-specific messages.
- Service/store failures currently map to `500` from handler.

### Important Notes
- Handler defaults `limit` to `100` when omitted.
- The web client maps PascalCase response DTO fields into camelCase frontend models.
- Date range ordering (`from <= to`) is not explicitly enforced in the dailies handler.
- Daily retrieval request/response DTO contracts are documented in code for maintainable API intent.

### Code References
- Dailies HTTP handler:
  - [dailies.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/dailies.go)
- Dailies request validation:
  - [requests.go](C:/personal/git/my-finances-tracker/apps/api/api/requests.go)
- Dailies service logic and metadata behavior:
  - [service.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/service.go)
- Daily sort normalization/domain model:
  - [daily.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/daily.go)
- Daily persistence/query with sort/paging:
  - [sqlx_daily_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_daily_store.go)
- Frontend DTO mapping:
  - [marketdata.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/marketdata.ts)
  - [marketdata.ts](C:/personal/git/my-finances-tracker/apps/web/src/types/marketdata.ts)

### Validation Coverage
- Integration test covers `/marketdata/dailies` happy path after listing create + async sync.
  - [endpoints_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/endpoints_integration_test.go)
- Service tests cover manual-provider behavior, syncing state messaging, and async sync triggering.
  - [service_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/service_test.go)

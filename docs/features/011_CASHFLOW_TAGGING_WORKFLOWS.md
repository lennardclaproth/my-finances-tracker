## [011] Cashflow Tagging Workflows

### Summary
Cashflow tagging supports three mutation patterns: single-transaction tagging, selected-ID batch tagging, and filter-based tagging that can execute synchronously or be queued as a background job.

### Why This Feature Exists
Users need fast manual categorization for small edits and scalable bulk updates for large filtered sets without blocking the API request path.

### HTTP Contract
1. Single transaction tag
- Endpoint: `POST /cashflow/transactions/tag`
- Body: `{ id, tag }`
- Success: `200 OK` with empty object payload `{}`

2. Selection-based tag update
- Endpoint: `POST /cashflow/transactions/tag/selection`
- Body: `{ tag, ids[] }`
- Success: `200 OK` with `updated_count` and status message

3. Filter-based tag update
- Endpoint: `POST /cashflow/transactions/tag/filter`
- Body: `{ tag, filters, account_id? }`
- Success:
  - `200 OK` for synchronous updates
  - `202 Accepted` when scheduled as background bulk-tag job

### Validation and Filter Semantics
- Selection endpoint requires non-empty `ids`.
- Filter endpoint builds a `CashflowTransactionQuery` from `filters` using the same semantics as transaction querying:
  - `q`, `description`, `note`, `source`, `direction`, `tags`, `untagged`, `hide_ignored`, `from`, `to`
- Filter date rules:
  - `from` / `to` must be `YYYY-MM-DD`
  - `from <= to`
- Invalid filter input returns `400` with a `filters` message.

### Sync vs Async Decision Rules (Filter Endpoint)
- The API first counts matches for the filter.
- Async path is used only when all conditions are true:
  - match count is greater than `1000`
  - a bulk-tag enqueuer is configured
  - `account_id` is provided and non-nil
- Async path response:
  - `202 Accepted`
  - `updated_count = 0`
  - status text indicates background scheduling
- Otherwise, update runs synchronously with `200 OK` and actual `updated_count`.

### Background Bulk Tag Job Behavior
- Queue-based worker job applies `UpdateTagByQuery` in background.
- Worker pool is capped (`maxBulkTagWorkers = 4`), queue is bounded.
- Queue overflow returns `ErrBulkTagQueueFull` to caller (surfaces as `500` on enqueue attempt).
- After successful background mutation, job publishes `BulkTagCompleted.v1` with `account_id` + updated count.

### Realtime and Frontend Effect
- `BulkTagCompleted.v1` is bridged to websocket event `bulk_tag.completed`.
- Cashflow page listens for `bulk_tag.completed` (and `import.completed`) and debounces refresh of transactions + analytics.
- UI mutation status is tone-aware: scheduled/background wording is shown as informational instead of success.

### Error Behavior
- Validation/filter parsing errors return `400`.
- Storage failures and enqueue failures return `500`.

### Important Notes
- `POST /cashflow/transactions/tag` does not enforce explicit field-level validation in handler logic; behavior relies on store update semantics.
- Empty filter objects are accepted for filter-based mutation; this can target a broad/all-row update depending on dataset and flags.
- Async scheduling requires `account_id`; without it, large filtered updates still execute synchronously.
- Tagging request/response DTO contracts are documented in code for clearer mutation API expectations.
- Filtered-tag async cutoff and enqueue-or-sync orchestration are implemented in a dedicated cashflow service, keeping handlers transport-focused.

### Code References
- Tagging/filtered tagging handlers:
  - [transactions.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/transactions.go)
- Cashflow store mutation methods:
  - [sqlx_transaction_cashflow_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_transaction_cashflow_store.go)
- Background bulk tag job and enqueue contract:
  - [bulk_tag.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/bulk_tag.go)
- Bulk tag completion event type:
  - [events.go](C:/personal/git/my-finances-tracker/apps/api/api/events.go)
- Realtime notify bridge:
  - [handlers.go](C:/personal/git/my-finances-tracker/apps/api/internal/notify/handlers.go)
- Frontend mutation calls and realtime refresh handling:
  - [cashflowTransactions.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/cashflowTransactions.ts)
  - [CashflowTransactionsPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/CashflowTransactionsPage.vue)

### Validation Coverage
- Integration tests cover single-tag updates, selection updates, and filter-based sync vs async behavior (`202` scheduling above threshold).
  - [endpoints_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/endpoints_integration_test.go)
- Job tests cover worker bound configuration, queue-full behavior, and queued-task processing.
  - [bulk_tag_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/bulk_tag_test.go)

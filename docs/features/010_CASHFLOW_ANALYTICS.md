## [010] Cashflow Analytics

### Summary
The API exposes two read-only analytics endpoints over cashflow transactions: monthly trend aggregation (`/cashflow/analytics/monthly`) and tag distribution aggregation (`/cashflow/analytics/tags`).

### Why This Feature Exists
Dashboard/reporting screens need pre-aggregated analytics views so clients can render trends and category breakdowns without downloading and reprocessing all transactions in-browser.

### HTTP Contract
1. Monthly trend endpoint
- Endpoint: `GET /cashflow/analytics/monthly`
- Query params:
  - `from` (optional `YYYY-MM-DD`)
  - `to` (optional `YYYY-MM-DD`)
  - `include_ignored` (optional bool)
- Response: `200 OK` with `api.CashflowMonthlyAnalyticsResponse`
  - `data[]` entries contain `month`, `incomingCents`, `outgoingCents`, `netCents`

2. Tag distribution endpoint
- Endpoint: `GET /cashflow/analytics/tags`
- Query params:
  - `from` (optional `YYYY-MM-DD`)
  - `to` (optional `YYYY-MM-DD`)
  - `include_ignored` (optional bool)
- Response: `200 OK` with `api.CashflowTagDistributionResponse`
  - `combined[]`, `incoming[]`, `outgoing[]` entries with `tag` and `totalCents`

### Date Range and Filtering Rules
- `from` and `to` are parsed as `YYYY-MM-DD`.
- Range is inclusive (`date >= from`, `date <= to`).
- If both dates are supplied, `from` must be before or equal to `to`.
- By default, ignored transactions are excluded.
- Setting `include_ignored=true` includes ignored transactions in aggregates.

### Monthly Aggregation Semantics
- Grouping key is month start (`YYYY-MM-01`).
- Groups are returned in ascending month order.
- Incoming total sums `direction='in'` amounts.
- Outgoing total sums `direction='out'` amounts.
- Net is computed server-side as `incoming - outgoing`.

### Tag Distribution Semantics
- Tags are normalized with `TRIM`; empty/NULL tags are bucketed as `untagged`.
- Aggregation runs by `(tag, direction)` and is projected into:
  - `combined` (all directions)
  - `incoming` (`direction='in'`)
  - `outgoing` (`direction='out'`)
- Each bucket list is sorted by `totalCents` descending, then tag name ascending for ties.

### Error Behavior
- Invalid date formats or invalid date ranges return `400` with a `date_range` message.
- Unexpected storage/runtime failures return `500`.

### Important Notes
- Storage implementation includes SQL dialect handling for month extraction (Postgres and SQLite).
- Outgoing values are represented as positive totals in `outgoingCents`; sign semantics are conveyed via the dedicated incoming/outgoing fields.
- Both endpoints are table-level analytics and do not currently apply an explicit account scope parameter.
- Analytics request/response DTO contracts are documented in code to keep schema intent stable.

### Code References
- Analytics handlers and date-range parsing:
  - [transactions.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/transactions.go)
- Analytics storage queries/aggregation logic:
  - [sqlx_transaction_cashflow_analytics.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_transaction_cashflow_analytics.go)
- Analytics request/response DTOs:
  - [requests.go](C:/personal/git/my-finances-tracker/apps/api/api/requests.go)
  - [responses.go](C:/personal/git/my-finances-tracker/apps/api/api/responses.go)
- Frontend analytics client/types:
  - [cashflowTransactions.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/cashflowTransactions.ts)
  - [cashflow.ts](C:/personal/git/my-finances-tracker/apps/web/src/types/cashflow.ts)

### Validation Coverage
- Integration tests verify monthly aggregation totals, date range behavior, default ignored filtering, and `include_ignored=true` behavior.
- Integration tests verify tag distribution buckets (`combined/incoming/outgoing`), `untagged` normalization, and default ignored exclusion.
  - [endpoints_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/endpoints_integration_test.go)

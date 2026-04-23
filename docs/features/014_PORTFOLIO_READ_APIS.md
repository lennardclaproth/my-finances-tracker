## [014] Portfolio Read APIs

### Summary
The portfolio read surface consists of three account-scoped endpoints for snapshots, positions, and transactions, with query validation and normalized response DTOs for chart/table consumption.

### Why This Feature Exists
Portfolio UI requires separate projections: time-series performance (`snapshots`), current holdings (`positions`), and detailed ledger history (`transactions`).

### Endpoint Contracts
1. Snapshots
- Endpoint: `GET /portfolio/snapshots`
- Query: `account_id` (required), `from`/`to` (optional `YYYY-MM-DD`)
- Response: `200 OK` with `[]api.PortfolioSnapshotPointResponse`

2. Positions
- Endpoint: `GET /portfolio/positions`
- Query: `account_id` (required), `include_closed` (optional bool, default false)
- Response: `200 OK` with `api.PortfolioPositionsResponse`

3. Transactions
- Endpoint: `GET /portfolio/transactions`
- Query: `account_id` (required), plus optional
  - `from`, `to`
  - `limit`, `offset`
  - `sort_by`, `sort_order`
  - `q`, `type`, `origin`, `source`, `listing`
- Response: `200 OK` with `api.PortfolioTransactionsResponse`

### Shared Validation and Account Gating
- All three handlers verify account existence first; unknown accounts return `404` with `account_id` message.
- `from`/`to` parsing enforces `YYYY-MM-DD` and `from <= to`.
- `to` is normalized to end-of-day UTC (`23:59:59.999999999`) so day filters are inclusive.

### Snapshots Read Semantics
- Snapshot store reads rows sorted by `occurred_at ASC`.
- Handler computes extra view fields:
  - `return_vs_cost_basis_pct` from `total_pnl / cost_basis` (0 when cost basis is zero)
  - `value_index = 100 * (1 + time_weighted_return_pct/100)`

### Positions Read Semantics
- Data is fetched with latest snapshot per position via left-join/subquery.
- `include_closed=false` filters out rows with `close_date` set.
- Response includes live valuation fields when latest snapshot exists (`market_value`, `unrealized_pnl_pct`, `last_snapshot_at`).

### Transactions Read Semantics
- Request defaults:
  - `limit=25` when omitted
  - sort defaults to `date desc`
- Validation constraints:
  - `limit` must be one of `10, 25, 50, 100` when non-zero
  - `sort_by` supports only `date`
  - `sort_order` must be `asc|desc`
  - `type` allowed: `BUY, SELL, DIVIDEND, TAX, FEE, CASH`
  - `origin` allowed: `IMPORT, MANUAL`
- Store filtering:
  - always account-scoped (`account_id` in WHERE)
  - optional contains filters for `source`, listing (`symbol/isin`), and global `q` across description/source/symbol/isin
- Pagination:
  - total count query + paged data query
  - ordered by `occurred_at`, then `created_at`, then `id`

### Response Formatting Notes
- Transaction numeric fields (`amount`, `quantity`, `unit_price`) are returned as trimmed decimal strings.
- For `CASH` transactions with negative quantity, handler negates amount for signed output consistency.

### Error Behavior
- `400` for invalid query values/date range.
- `404` for missing account.
- `500` for storage/runtime failures.

### Frontend Consumption
- Portfolio page consumes all three endpoints and maps route query state to API params.
- Route utilities enforce client-side defaults aligned with backend constraints (`limit=25`, `sort_by=date`, `sort_order=desc`).

### Important Notes
- Portfolio read request/response DTO contracts are documented in code so filter and payload semantics stay explicit.
- Snapshot derived-metric calculations used by read responses are centralized in portfolio projection helpers instead of handler-local formulas.

### Code References
- Portfolio read handlers and date parsing:
  - [portfolio.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/portfolio.go)
- Portfolio request/response DTOs:
  - [requests.go](C:/personal/git/my-finances-tracker/apps/api/api/requests.go)
  - [responses.go](C:/personal/git/my-finances-tracker/apps/api/api/responses.go)
- Snapshot store:
  - [sqlx_portfolio_snapshot_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_portfolio_snapshot_store.go)
- Position store:
  - [sqlx_position_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_position_store.go)
- Portfolio transaction store/query normalization:
  - [sqlx_transaction_portfolio_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_transaction_portfolio_store.go)
  - [transaction_query.go](C:/personal/git/my-finances-tracker/apps/api/internal/portfolio/transaction_query.go)
- Decimal formatting helper used by portfolio responses:
  - [portfolio_transactions.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/portfolio_transactions.go)
- Frontend consumers:
  - [portfolio.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/portfolio.ts)
  - [PortfolioPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/PortfolioPage.vue)
  - [routeQuery.ts](C:/personal/git/my-finances-tracker/apps/web/src/utils/routeQuery.ts)

### Validation Coverage
- Handler tests cover unknown-account behavior, date-range validation, positions include-closed behavior, transaction query mapping, and response shaping.
  - [portfolio_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/portfolio_test.go)

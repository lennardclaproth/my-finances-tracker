## [009] Cashflow Transaction Querying

### Summary
`GET /cashflow/transactions` provides paginated cashflow transaction retrieval with composable filters, explicit sort controls, and deterministic response metadata for list/table UIs.

### Why This Feature Exists
Cashflow workflows need a single read API that supports exploration (search), operational filtering (direction/date/tag/ignored), and predictable pagination for large imported datasets.

### HTTP Contract
- Endpoint: `GET /cashflow/transactions`
- Query parameters:
  - Pagination: `limit`, `offset`
  - Sorting: `sort_by`, `sort_order`
  - Search/filter: `q`, `description`, `note`, `source`, `direction`, `tags`, `untagged`, `hide_ignored`, `from`, `to`
- Success response: `200 OK` with `api.CashflowTransactionsResponse`
  - `pagination`: `limit`, `offset`, `count`, `total`
  - `data`: transaction rows (`id`, `description`, `note`, `source`, `amountCents`, `direction`, `date`, `tag`, `ignored`)

### Query Validation Behavior
- `limit` and `offset` must be `>= 0`.
- `sort_by` must be one of `date | description | note | tag | source | amount`.
- `sort_order` must be `asc` or `desc`.
- `direction` must be `in` or `out`.
- `from` and `to` must parse as `YYYY-MM-DD`.
- If both dates are present, `from` must be before or equal to `to`.

### Defaults and Normalization
- Default `limit` is `100` when omitted or `0`.
- Default sort field is `date`.
- Default sort order is `DESC`.
- `sort_by=amount` maps to database column `amount_cents`.
- `tags` is split by comma, whitespace-trimmed, and case-insensitively deduplicated before querying.

### Filter Semantics
1. Field filters (`description`, `note`, `source`)
- Case-insensitive `contains` matching (`LIKE %value%`).

2. Direction
- Exact match against normalized lowercase value (`in` / `out`).

3. Tags
- Multiple tags are combined with `OR` semantics (`food,travel` returns either tag).
- Tag matching is case-insensitive.

4. Untagged and ignored state
- `untagged=true` filters rows where tag is empty/NULL.
- `hide_ignored=true` filters out rows with `ignored=true`.

5. Date range and fuzzy query
- `from` and `to` are inclusive bounds on `date`.
- `q` applies fuzzy contains matching across `description`, `note`, and `tag`.

### Error Behavior
- Invalid query inputs return `400` with field-level payloads.
- Unexpected storage/runtime failures return `500`.

### Important Notes
- This endpoint is not account-scoped at the request level; it queries the cashflow transaction table directly using provided filters.
- Pagination returns both `count` (current page size) and `total` (matching rows before pagination), enabling stable client-side pager controls.
- Frontend query construction intentionally omits false booleans (`untagged`/`hide_ignored`) unless enabled.
- Cashflow transaction query request/response DTO contracts are documented in code.

### Code References
- Cashflow transaction read handler and filter parsing:
  - [transactions.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/transactions.go)
- Request DTO validation:
  - [requests.go](C:/personal/git/my-finances-tracker/apps/api/api/requests.go)
- Store query builder, filtering, counting, and sort normalization:
  - [sqlx_transaction_cashflow_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_transaction_cashflow_store.go)
- API response DTOs:
  - [responses.go](C:/personal/git/my-finances-tracker/apps/api/api/responses.go)
- Frontend API client/query mapping:
  - [cashflowTransactions.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/cashflowTransactions.ts)
  - [cashflow.ts](C:/personal/git/my-finances-tracker/apps/web/src/types/cashflow.ts)

### Validation Coverage
- Integration tests cover combined search/filter behavior, pagination, `hide_ignored`, amount sorting, direction filtering, and invalid direction handling.
  - [endpoints_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/endpoints_integration_test.go)
- Endpoint decoding tests cover invalid query decoding behavior (for example non-integer `limit`).
  - [endpoint_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/endpoint_test.go)

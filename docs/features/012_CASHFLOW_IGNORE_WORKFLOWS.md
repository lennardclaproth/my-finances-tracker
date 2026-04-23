## [012] Cashflow Ignore Workflows

### Summary
Cashflow ignore workflows allow marking transactions as ignored/not ignored by explicit selection or by filter query, affecting both transaction list visibility and analytics totals.

### Why This Feature Exists
Users need a reversible way to exclude noisy or non-relevant transactions from daily cashflow analysis without deleting historical data.

### HTTP Contract
1. Selection-based ignore update
- Endpoint: `POST /cashflow/transactions/ignore/selection`
- Body: `{ ids: uuid[], ignored?: boolean }`
- Success: `200 OK` with `updated_count` and status message

2. Filter-based ignore update
- Endpoint: `POST /cashflow/transactions/ignore/filter`
- Body: `{ filters, ignored?: boolean }`
- Success: `200 OK` with `updated_count` and status message

### Defaulting and Validation Rules
- If `ignored` is omitted, handlers default it to `true`.
- Selection endpoint requires non-empty `ids`.
- Filter endpoint reuses shared cashflow filter parsing semantics:
  - supports `q`, `description`, `note`, `source`, `direction`, `tags`, `untagged`, `hide_ignored`, `from`, `to`
- Invalid filter date/direction/range values return `400` with a `filters` message.

### Update Semantics
- Selection endpoint updates only IDs in the payload (`WHERE id IN (...)`).
- Filter endpoint updates all rows matching the built `CashflowTransactionQuery`.
- Both paths return total affected rows in `updated_count`.
- Status message includes final state label (`ignored` or `not ignored`).

### Operational Differences vs Tagging
- Ignore workflows have no async/offloaded path; all updates execute synchronously.
- No `account_id` parameter is required for ignore mutations.
- No dedicated completion event is emitted specifically for ignore updates.

### Downstream Effects
- Querying with `hide_ignored=true` excludes ignored rows from `GET /cashflow/transactions`.
- Analytics endpoints exclude ignored rows by default unless `include_ignored=true` is explicitly set.

### Error Behavior
- Validation/parsing failures return `400`.
- Storage/runtime failures return `500`.

### Important Notes
- Empty filter objects are accepted for filter updates; this can update broad/all rows depending on existing data.
- The ignored flag is a soft-classification attribute and can be toggled back to active (`ignored=false`).
- Ignore workflow request/response DTO contracts are documented in code.

### Code References
- Ignore handlers and defaulting logic:
  - [transactions.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/transactions.go)
- Ignore request DTOs and selection validation:
  - [requests.go](C:/personal/git/my-finances-tracker/apps/api/api/requests.go)
- Store mutation methods:
  - [sqlx_transaction_cashflow_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_transaction_cashflow_store.go)
- Frontend mutation API calls and action flow:
  - [cashflowTransactions.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/cashflowTransactions.ts)
  - [CashflowTransactionsPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/CashflowTransactionsPage.vue)

### Validation Coverage
- Integration tests verify selection-based ignore updates and filter-based ignore updates with expected row-level effects.
  - [endpoints_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/endpoints_integration_test.go)

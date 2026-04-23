## [023] Manual Cashflow Transaction Creation

### Summary
`POST /cashflow/transactions/manual` creates manual cashflow transactions in bulk for a selected account, and the cashflow page now exposes a floating action button with a multi-row modal form to submit those entries.

### Why This Feature Exists
Users need to manually add missing cashflow transactions without importing a CSV, including creating multiple entries in one action.

### HTTP Contract
- Endpoint: `POST /cashflow/transactions/manual`
- Body fields:
  - Required top-level: `account_id`, `transactions[]`
  - Required per transaction: `date`, `amount`, `type`, `description`, `note`, `tag`
  - Optional per transaction: `vendor`
- Success: `201 Created` with:
  - `created_count`
  - `data[]` containing created transaction rows

### Validation and Domain Rules
1. Account and projection checks
- `account_id` must exist.
- The service ensures cashflow/import account projections exist before persisting rows.

2. Row-level field checks
- `date` must be `YYYY-MM-DD`.
- `amount` must be a positive decimal string (up to 6 decimals).
- `type` must be `in` or `out` (also accepts `income`/`expense` synonyms).
- `description`, `note`, and `tag` are required and trimmed.
- Batch size supports up to `100` rows per request.

3. Manual source flag behavior
- Created transactions are marked as manual via source:
  - `manual` when vendor is omitted
  - `manual:<vendor>` when vendor is provided
- This enables distinguishing manual rows from imported vendor-source rows.

### Persistence Behavior
- The service creates a completed import record used as lineage for the inserted rows.
- Each created transaction is linked to that import record and account.
- The response returns created rows with direction, date, tag, and source values.

### Error Mapping
- `400`: validation or domain input errors
- `404`: unknown account
- `409`: duplicate transaction checksum conflict
- `500`: unexpected persistence/runtime failures

### Frontend Behavior
- Cashflow page shows a floating action button to open create modal.
- Floating action button placement/properties are aligned with the portfolio and assets pages for consistent create-action UX.
- Modal supports multi-row entry in one submit.
- Date fields use the shared calendar popover UI pattern used by navbar date selection (instead of native browser date inputs).
- Save waits for server success, then:
  - closes modal
  - shows success toast
  - refreshes transactions and analytics

### Code References
- API handler and route wiring:
  - [transactions.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/transactions.go)
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)
- Manual create service:
  - [manual_create_service.go](C:/personal/git/my-finances-tracker/apps/api/internal/cashflow/service/manual_create_service.go)
- API DTO contracts:
  - [requests.go](C:/personal/git/my-finances-tracker/apps/api/api/requests.go)
  - [responses.go](C:/personal/git/my-finances-tracker/apps/api/api/responses.go)
- Frontend create flow:
  - [CreateCashflowTransactionsModal.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/CreateCashflowTransactionsModal.vue)
  - [CashflowTransactionsPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/CashflowTransactionsPage.vue)
  - [cashflowTransactions.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/cashflowTransactions.ts)

### Validation Coverage
- Service tests cover successful bulk creation, unknown-account handling, and invalid amount validation.
  - [manual_create_service_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/cashflow/service/manual_create_service_test.go)
- Handler tests cover success and status mapping (`400`/`404`/`409`).
  - [transactions_manual_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/transactions_manual_test.go)
- Integration tests cover endpoint happy-path and unknown-account behavior.
  - [cashflow_manual_transactions_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/cashflow_manual_transactions_integration_test.go)

## [015] Manual Portfolio Transaction Creation

### Summary
`POST /portfolio/transactions/manual` creates manual-origin portfolio transactions (`MANUAL`) with strict domain validation, persists them without immediate rebuild execution, and returns a normalized transaction DTO.

### Why This Feature Exists
Users need to correct or add portfolio ledger entries that are not present in imported files while preserving the same transaction model used by downstream portfolio reads/builds.

### HTTP Contract
- Endpoint: `POST /portfolio/transactions/manual`
- Body fields:
  - Required: `account_id`, `vendor_id`, `occurred_at`, `type`, `amount`
  - Conditional: `listing_id` (required for non-cash, forbidden for cash), `quantity` (required for BUY/SELL)
  - Optional: `description`
- Success: `201 Created` with `api.ManualPortfolioTransactionResponse`

### Validation and Domain Rules
1. Entity checks
- Account must exist.
- Vendor must exist, be active, and be of type `brokerage`.
- For non-cash types, `listing_id` must exist and listing must expose at least one of `isin` or `symbol`.

2. Field format/value checks
- `occurred_at` must be `YYYY-MM-DD`.
- `type` must be one of: `BUY, SELL, DIVIDEND, TAX, FEE, CASH`.
- `amount` and `quantity` use decimal-string validation (up to 6 decimals).
- Non-cash amount must be positive.
- Cash amount must be non-zero (positive or negative allowed).
- `quantity` rules:
  - required and positive for `BUY`/`SELL`
  - forbidden for all other types

3. Derived values
- `BUY`/`SELL` compute `unit_price = amount / quantity`.
- Cash transactions encode direction via quantity marker (`-1` for outflow, `+1` for inflow) while stored amount remains absolute.
- Response includes `signed` amount view behavior (cash out is negative in API output).

### Persistence and Response Behavior
- Transaction is created via `NewManualTransaction` with:
  - `origin = MANUAL`
  - `import_id = nil`
- Record is persisted in portfolio transactions store with checksum-based duplicate detection.
- Handler returns decimal fields (`amount`, `quantity`, `unit_price`) as trimmed string values.

### Error Mapping
- `400`: validation and domain input errors
- `404`: account/vendor/listing not found
- `409`: duplicate transaction
- `422`: unsupported vendor type for manual portfolio transactions
- `500`: unexpected runtime/store errors

### Important Notes
- Endpoint intentionally does not publish rebuild events on create; rebuild/refresh is handled through separate workflows.
- Frontend create modal mirrors key backend rules (brokerage vendor filter, decimal validation, type-specific required fields) before submit.
- Frontend occurred-at date input uses the shared calendar popover UI pattern used by navbar date selection (instead of native browser date input).
- Manual transaction create request/response DTO contracts are documented in code.

### Code References
- Manual transaction HTTP handler:
  - [portfolio_transactions.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/portfolio_transactions.go)
- Manual transaction domain service and validation rules:
  - [manual_transaction_service.go](C:/personal/git/my-finances-tracker/apps/api/internal/portfolio/manual_transaction_service.go)
- Transaction model/origin semantics:
  - [transaction.go](C:/personal/git/my-finances-tracker/apps/api/internal/portfolio/transaction.go)
- Portfolio transaction persistence (duplicates/checksum):
  - [sqlx_transaction_portfolio_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_transaction_portfolio_store.go)
- Frontend create flow:
  - [portfolio.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/portfolio.ts)
  - [CreatePortfolioTransactionModal.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/CreatePortfolioTransactionModal.vue)

### Validation Coverage
- Handler tests cover success response mapping, duplicate handling (`409`), vendor-type rejection (`422`), and request validation (`400`).
  - [portfolio_transactions_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/portfolio_transactions_test.go)
- Service tests cover successful BUY/CASH creation, signed cash direction behavior, vendor-type rejection, and duplicate propagation.
  - [manual_transaction_service_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/portfolio/manual_transaction_service_test.go)

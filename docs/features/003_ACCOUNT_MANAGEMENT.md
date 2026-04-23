## [003] Account Management Endpoints

### Summary
The API provides account creation and listing endpoints. Created accounts are the canonical identity used across importer, cashflow, and portfolio domains.

### Why This Feature Exists
Most domain operations are account-scoped (imports, portfolio reads/rebuilds, cashflow views). A central account record is required so all module projections and workflows can reference the same account ID.

### HTTP Contract
1. Create account
- Endpoint: `POST /accounts`
- Request body:
  - `name` (required, non-empty)
  - `external_id` (optional UUID)
- Success response: `200 OK` with `api.AccountResponse`
- Error responses:
  - `400` for validation issues (for example missing/blank name)
  - `409` if account already exists

2. List accounts
- Endpoint: `GET /accounts`
- Success response: `200 OK` with `[]api.AccountResponse`
- Ordering: sorted by account name ascending

### Domain Model Behavior
Account entity fields:
- `id` (generated UUID)
- `external_id` (optional)
- `name` (trimmed, required)
- `created_at`, `updated_at`

Validation behavior:
- Name is trimmed and must not be empty (`ErrAccountNameRequired`).

### Persistence and Uniqueness
- Accounts are stored in the account schema/table via `SQLXAccountStore`.
- Duplicate create attempts map to `ErrAccountAlreadyExists` and return HTTP `409`.
- Uniqueness detection supports both PostgreSQL unique-violation code handling and SQLite-style unique error message matching.

### Event-Driven Projection Fanout
After successful account creation, the create service publishes `AccountCreated.v1` on the internal bus.

Bus subscribers create module-local account projections:
- Portfolio account projection (`portfolio_accounts`)
- Cashflow account projection (`cashflow_accounts`)
- Importer account projection (`import_accounts`)

Projection inserts use `ON CONFLICT (account_id) DO NOTHING`, making projection creation idempotent across retries.

### Downstream Feature Impact
This feature enables:
- import account validation before accepting `/import/csv`
- portfolio build lock/account existence checks
- consistent account-scoped filtering and reads in cashflow/portfolio APIs

### Important Notes
- The API currently returns `200` (not `201`) for successful account creation.
- `external_id` is optional and passed through as-is.
- Projection fanout depends on bus availability; when bus is configured (standard server path), creation triggers all projections.

### Code References
- Account HTTP handlers:
  - [accounts.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/accounts.go)
- Account domain model and validation:
  - [account.go](C:/personal/git/my-finances-tracker/apps/api/internal/account/account.go)
- Account create service + event publish:
  - [service.go](C:/personal/git/my-finances-tracker/apps/api/internal/account/service.go)
- Account primary store:
  - [sqlx_account_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_account_store.go)
- Projection stores:
  - [sqlx_portfolio_account_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_portfolio_account_store.go)
  - [sqlx_cashflow_account_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_cashflow_account_store.go)
  - [sqlx_import_account_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_import_account_store.go)
- Route and subscription wiring:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)

### Validation Coverage
- Integration coverage validates account endpoint behavior directly:
  - `POST /accounts` returns `400` for blank names.
  - `POST /accounts` returns `409` for duplicate names.
  - `GET /accounts` returns rows sorted by account name ascending.
- Event subscription wiring for `AccountCreated` is defined in server composition root.

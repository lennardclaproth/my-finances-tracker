## [019] Automatic Migrations and Bootstrap Seeding

### Summary
On startup, the API ensures database readiness, runs embedded migrations, and bootstraps baseline domain records (vendors, default account, providers) before serving traffic.

### Why This Feature Exists
Fresh environments should become runnable without manual schema setup or seed scripts, while existing environments should remain idempotent across restarts.

### Startup Flow
1. Configuration load
- Reads `config.yaml`, hydrates provider env values, applies APM defaults, validates required settings.

2. Database initialization
- Opens DB connection based on configured type (`sqlite3` or `postgres`).
- For Postgres, ensures target database exists by connecting to `postgres` DB first.
- Runs migrations via goose provider using embedded SQL files.

3. Bootstrap seed phase
- Seeds supported vendors.
- Ensures one default account exists (stable UUID/name).
- Seeds provider records from config/env plus manual provider(s).

### Migration System Details
- Migration SQL is embedded at build time (`//go:embed postgres sqlite`).
- Dialect is selected at runtime (`goose.DialectSQLite3` or `goose.DialectPostgres`).
- Current baseline migration initializes core domains:
  - account, vendor, import, cashflow, portfolio, marketdata, provider, daily uploads.

### Bootstrap Semantics
1. Vendors bootstrap
- Iterates `vendor.SupportedVendors` and creates missing vendors.
- Existing vendors are logged and skipped.

2. Accounts bootstrap
- Checks for fixed bootstrap account ID.
- Creates default account if missing.
- Uses account create service, so account-created event fanout also runs through normal path.

3. Providers bootstrap
- API providers (`marketstack`, `alphavantage`) created per configured API key.
- If no keys configured for a provider, bootstrap logs skip.
- Manual providers are always seeded (currently `BrandNewDay`).

### Error and Idempotency Behavior
- Bootstrap functions panic on unexpected failures to fail-fast startup.
- Expected already-exists paths are handled idempotently.
- Migration execution reports applied count and is safe across repeated startup runs.

### Important Notes
- Default bootstrap account is hardcoded (`64a3d50f-c71a-4015-9ee1-45572147ce56`, "Lennard Claproth").
- Provider API keys are sourced from env vars and support comma-separated multi-key input.
- APM transaction sample rate defaults to `0.2` in production and `1.0` otherwise unless explicitly set.
- Provider bootstrap contracts used during startup seeding are documented in code.

### Code References
- Startup composition and bootstrap invocation:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)
- Migrator implementation and embedded FS loading:
  - [migrator.go](C:/personal/git/my-finances-tracker/apps/api/migrations/migrator.go)
- Baseline migration files:
  - [20260202192351_init.sql](C:/personal/git/my-finances-tracker/apps/api/migrations/postgres/20260202192351_init.sql)
  - [20260202192351_init.sql](C:/personal/git/my-finances-tracker/apps/api/migrations/sqlite/20260202192351_init.sql)
- Bootstrap seed implementations:
  - [vendors.go](C:/personal/git/my-finances-tracker/apps/api/internal/bootstrap/vendors.go)
  - [accounts.go](C:/personal/git/my-finances-tracker/apps/api/internal/bootstrap/accounts.go)
  - [providers.go](C:/personal/git/my-finances-tracker/apps/api/internal/bootstrap/providers.go)
- Config loading/hydration/defaults:
  - [config.go](C:/personal/git/my-finances-tracker/apps/api/internal/config/config.go)

### Validation Coverage
- Config tests validate APM sample-rate defaulting and env override behavior.
  - [config_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/config/config_test.go)

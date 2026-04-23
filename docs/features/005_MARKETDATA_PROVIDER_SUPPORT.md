## [005] Market Data Provider Support

### Summary
The platform has a provider layer that models how each market data source is ingested (`API` vs `MANUAL`), bootstraps provider records at startup, and lets market data workflows choose behavior based on provider capabilities.

### Why This Feature Exists
Different data sources have different ingestion modes:
- API providers require base URI + API key and support automated syncing.
- Manual providers do not support automated API pulls and are intended for file-based uploads.

A dedicated provider model keeps this behavior explicit and enforceable.

### Provider Model
Each provider record includes:
- `name` (for example `marketstack`, `alphavantage`, `brandnewday`)
- `ingestion_mode` (`API` or `MANUAL`)
- connection fields (`base_uri`, `api_key`) for API mode
- token counters (`remaining`, `used`, `total`, `resets_at`) for API usage accounting

Validation rules:
- API mode requires non-empty `base_uri` and `api_key`.
- Manual mode requires connection fields to be `nil`.
- Unsupported source/name mappings are rejected.

### Configuration and Bootstrap
Provider settings are hydrated from config + environment:
- `MARKETSTACK_BASE_URI` / `MARKETSTACK_API_KEY`
- `ALPHA_VANTAGE_BASE_URI` / `ALPHA_VANTAGE_API_KEY`

Bootstrap behavior at startup:
1. Create API provider entries for configured API keys (one entry per key).
2. Create manual provider entries for manual-only providers (currently Brand New Day).
3. Skip API provider bootstrap when no keys are configured.

Provider store creation is idempotent for duplicates (manual by name+mode, API by name+mode+api_key).

### Provider Selection Strategy
When resolving a provider by name, store selection prioritizes:
1. Manual ingestion provider first (if present).
2. Otherwise API provider with available quota preference:
   - deprioritize exhausted keys (`remaining <= 0` when `total > 0`)
   - then prefer higher remaining quota, lower used count.

### API Ingestion Path
The MarketStack client uses provider metadata to build requests dynamically:
- fetch provider by name
- verify API mode + required connection fields
- call provider API endpoint with access key
- deduct token budget based on page count used

This keeps API credentials and quota usage in provider storage rather than hardcoded in client code.

### Manual Ingestion Path
Manual providers affect runtime behavior in two key places:

1. Daily read behavior (`GetDailies`)
- If listing source maps to a manual provider, automatic sync is skipped.
- Data is served from DB only, with metadata message:
  - `Manual provider configured; automatic sync disabled`

2. Daily file upload behavior (`/marketdata/dailies/upload`)
- Upload is accepted only when listing source maps to a manual provider.
- If provider is API mode or missing, request is rejected with `422`.

### Downstream Integration Points
Provider support directly controls:
- whether listing creation triggers async daily sync,
- whether dailies reads can trigger background accumulation,
- whether manual daily upload is allowed for a listing source.

### Important Notes
- There is no public provider-management HTTP endpoint in current API; provider records are bootstrapped and used internally.
- Manual mode is capability-driven; source-to-provider mapping is derived from listing `source`.
- API key rotation/update helpers exist in store, but runtime flow currently centers on bootstrap + selection + token deduction.
- Provider/bootstrap exported contracts are documented in code to keep ingestion-mode behavior easier to maintain.

### Code References
- Provider domain model and validation:
  - [provider.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/provider.go)
- Provider bootstrap logic:
  - [providers.go](C:/personal/git/my-finances-tracker/apps/api/internal/bootstrap/providers.go)
- Provider config hydration:
  - [config.go](C:/personal/git/my-finances-tracker/apps/api/internal/config/config.go)
- Provider persistence and selection rules:
  - [sqlx_provider_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_provider_store.go)
- API client usage of provider metadata:
  - [client_marketstack.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/client_marketstack.go)
- Market data service behavior by provider mode:
  - [service.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/service.go)
- Manual upload provider enforcement:
  - [daily_uploads.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/daily_uploads.go)
- Wiring at startup:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)

### Validation Coverage
- Provider constructors and mapping tests:
  - [provider_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/provider_test.go)
- Service tests verifying manual provider disables async sync:
  - [service_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/service_test.go)
- Handler tests verifying upload rejection for non-manual/missing providers:
  - [daily_uploads_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/daily_uploads_test.go)

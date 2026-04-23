## [004] Active Vendor Listing

### Summary
The API exposes active vendors for client selection and returns vendor import capability metadata (`import_disabled`) used to gate CSV imports.

### Why This Feature Exists
Import workflows need a stable list of available vendors, while still allowing operations teams to disable importing for specific vendors without removing them from the system.

### HTTP Contract
- Endpoint: `GET /vendors`
- Success response: `200 OK` with `[]api.VendorResponse`
- Response fields include:
  - `id`
  - `name`
  - `type` (`bank` or `brokerage`)
  - `active`
  - `import_disabled`
  - audit timestamps (`created_at`, `updated_at`)

### Selection Semantics
- `GET /vendors` returns vendors where `active = true`.
- Result ordering is alphabetical by vendor name.
- `import_disabled` is informational in listing output and is enforced by import ingestion logic.

### Vendor Capability Enforcement
Vendor listing and import enforcement are intentionally separated:

1. Listing behavior
- Active vendors are returned even if `import_disabled = true`.
- This allows clients/admin tools to display vendor state explicitly.

2. Import behavior
- During `/import/csv`, importer fetches vendor by ID and rejects import when `ImportDisabled` is true.
- Rejection maps to `400` with vendor-specific problem payload.

### Data Model and Constraints
Vendor entity stores:
- identity (`id`, `name`)
- capability flags (`active`, `import_disabled`)
- classification (`type`)
- timestamps

Supported vendor names and types are defined in code (`SupportedVendors`) and validated when new vendors are created.

### Bootstrap Behavior
At startup, the app bootstraps supported vendors:
- attempts to create each configured supported vendor
- skips records that already exist
- panics on unexpected bootstrap failures

This ensures a predictable baseline vendor catalog for imports and UI selection.

### Important Notes
- “Active” controls visibility in `/vendors` results.
- “Import disabled” controls whether `/import/csv` accepts that vendor.
- Duplicate vendor names are rejected at store level (`ErrVendorAlreadyExists`).
- Duplicate detection supports PostgreSQL unique-violation handling and SQLite-style unique message matching so bootstrap can safely skip already-seeded vendors across both backends.

### Code References
- Vendor listing handler:
  - [vendors.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/vendors.go)
- Vendor store (`ListActive`, fetch/create):
  - [sqlx_vendor_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_vendor_store.go)
- Vendor model and supported vendor/type definitions:
  - [vendor.go](C:/personal/git/my-finances-tracker/apps/api/internal/vendor/vendor.go)
- Vendor creation helper used for bootstrap:
  - [create.go](C:/personal/git/my-finances-tracker/apps/api/internal/vendor/create.go)
- Vendor bootstrap at startup:
  - [vendors.go](C:/personal/git/my-finances-tracker/apps/api/internal/bootstrap/vendors.go)
- Import-side vendor enforcement:
  - [from_csv.go](C:/personal/git/my-finances-tracker/apps/api/internal/importer/from_csv.go)
  - [imports.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/imports.go)

### Validation Coverage
- Integration test verifies `/vendors` returns active vendors only and excludes inactive ones.
  - [endpoints_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/endpoints_integration_test.go)

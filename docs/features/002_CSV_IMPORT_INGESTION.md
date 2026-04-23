## [002] CSV Import Ingestion

### Summary
The API supports multipart CSV uploads that create durable import jobs, persist the raw file to disk, and process imports asynchronously through the background `ImportJob` worker.

### Why This Feature Exists
Importing transaction history can be large and vendor-specific. Processing it synchronously in HTTP would be slow and fragile, so uploads are accepted quickly and reconciled reliably in background workers.

### HTTP Contract
- Endpoint: `POST /import/csv`
- Content type: `multipart/form-data`
- Required fields:
  - `file` (must be `text/csv`)
  - `vendor_id` (UUID)
  - `account_id` (UUID)
- Response on success: `200 OK` with the created import ID as a raw JSON UUID string

### Request Validation and Limits
- Max upload size: 20 MB (`MaxBytes`)
- Multipart memory budget: 40 MB (`MaxMemory`)
- Allowed file MIME type: `text/csv`
- Domain checks:
  - vendor must exist and not be import-disabled
  - account must exist in importer account projection

### Ingestion Flow
1. HTTP handler decodes multipart payload and validates fields.
2. CSV file is written to disk under configured import storage path.
3. Import record is created with status `pending`.
4. Handler returns the import UUID immediately.
5. If enqueue succeeds, processing starts soon; if enqueue fails, periodic DB reconciliation still picks up pending imports.

### Durable Import Model
Each import is persisted with:
- identifiers (`id`, `vendor_id`, optional `account_id`)
- file location (`path`)
- lifecycle state (`pending`, `in_progress`, `completed`, `failed`)
- result metrics (`duplicates`, `total_rows`, `imported`, `failed`)
- timestamps and optional status message

### Async Processing Semantics
`ImportJob` processes queued and reconciled pending imports with these behaviors:
- Deduplicates in-queue/in-flight IDs to avoid duplicate processing.
- Uses `TryMarkInProgress` to claim pending imports safely.
- Parses and persists cashflow rows for all supported vendors.
- For brokerage vendors, also parses and persists portfolio transactions.
- Marks import as completed even when partial row failures occur, with counters persisted.
- Marks import as failed when critical processing steps fail.

### Reliability and Recovery
- Queue is bounded; full queue returns enqueue error to caller path, but pending imports remain in DB.
- Periodic `syncQueueFromDB` re-enqueues pending imports (catch-up behavior).
- Failed immediate enqueue from HTTP does not lose the import because persistence already happened.
- Partial file writes are cleaned up on write/sync failure.

### Downstream Effects
On successful processing, import flow publishes internal events:
- `TransactionsCreated` for portfolio rebuild pipeline (when account is present)
- `ImportCompleted` for realtime client notifications

### Important Notes
- This feature currently returns only the raw import UUID response; it does not expose a dedicated import status read endpoint.
- Import account validation depends on importer account projection availability.
- Account ID is effectively required for the current import path and downstream processing.

### Code References
- HTTP endpoint and multipart decoding:
  - [imports.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/imports.go)
- Import domain handler:
  - [from_csv.go](C:/personal/git/my-finances-tracker/apps/api/internal/importer/from_csv.go)
- Import model and statuses:
  - [import.go](C:/personal/git/my-finances-tracker/apps/api/internal/importer/import.go)
- Disk persistence:
  - [disk.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/disk.go)
- Import store and state transitions:
  - [sqlx_import_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_import_store.go)
- Background worker implementation:
  - [import.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/import.go)
- Route and job wiring:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)

### Validation Coverage
- Endpoint integration test verifies CSV upload creates import record and file on disk.
  - [endpoints_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/endpoints_integration_test.go)
- Job tests verify enqueue deduplication, queue reconciliation, and catch-up after queue-full conditions.
  - [import_queue_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/import_queue_test.go)

## [008] Asynchronous Daily Upload Processing

### Summary
The API supports asynchronous ingestion of manually provided daily market data files through `POST /marketdata/dailies/upload`, then exposes processing progress and row-level diagnostics through `GET /marketdata/dailies/uploads/{upload_id}`.

### Why This Feature Exists
Manual market-data providers require operator-uploaded files, and processing those files synchronously would block request latency and reduce reliability during heavier ingestion runs.

### HTTP Contract
- Submit upload: `POST /marketdata/dailies/upload`
- Poll status: `GET /marketdata/dailies/uploads/{upload_id}`
- Submit request content type: `multipart/form-data`
- Required form fields:
  - `file`
  - `listing_id`
- Submit success response: `202 Accepted`
  - `upload_id`
  - `status` (initially `PENDING`)
- Status success response: `200 OK` with upload lifecycle/counters:
  - `id`, `listing_id`, `source`, `status`, `status_message`
  - `total_rows`, `inserted_rows`, `duplicate_rows`, `error_rows`
  - `row_errors`
  - `created_at`, `started_at`, `finished_at`, `updated_at`

### Upload Validation and Guardrails
1. Transport-level decode/validation
- Multipart decoding is centralized and returns structured client errors.
- Maximum request size is capped (`MaxBytes: 10MB`) with `413` when exceeded.

2. Upload request validation
- `listing_id` is required and must be a UUID.
- Filename extension must be `.csv` or `.txt`.
- Empty files are rejected.

3. Domain compatibility checks
- `listing_id` must resolve to an existing listing.
- Listing source must map to a provider.
- Provider must exist and be configured for manual ingestion.
- Listing source must have a registered daily parser.

### Asynchronous Processing Lifecycle
1. The HTTP handler stores the uploaded file to disk and creates a `daily_uploads` record in `PENDING` state.
2. The handler attempts to enqueue the upload for background processing.
3. If enqueue fails, the request still returns `202 Accepted` (record persisted), and the failure is logged for reconciliation.
4. The `DailyUploadJob` reconciles pending records from DB at startup and on a periodic interval, then processes queued uploads.
5. Processing attempts an atomic claim (`TryMarkProcessing`) to avoid duplicate concurrent execution.

### Row Processing Semantics
- Parser output row errors are preserved.
- Valid rows are converted to domain `Daily` values and persisted one-by-one.
- Duplicate day inserts are tracked separately from inserted rows.
- Persist/validation failures are converted into row-level error entries.
- Final status resolution:
  - `SUCCEEDED` when `error_rows == 0`
  - `PARTIAL` when there are errors but at least one inserted/duplicate row
  - `FAILED` when no successful/duplicate rows remain
- Stored `row_errors` are sampled/capped to 50 entries.

### Error Behavior
- Client input failures return structured `4xx` payloads (for example invalid UUID, unsupported extension, missing file, oversized payload).
- Listing-not-found returns `404`.
- Provider/parser incompatibility returns `422`.
- Unexpected storage/runtime failures return `500`.
- Status lookup returns `404` when `upload_id` does not exist.

### Important Notes
- File names on disk are randomized and normalized to `.csv`; original filename is retained in upload metadata.
- File cleanup is attempted when upload record creation fails after file write.
- Background processing is resilient to process restarts because pending uploads are reconciled from storage.
- Daily upload request/status DTO contracts are documented in code so async API contracts remain clear.
- Manual-upload eligibility rules (provider mode + parser availability) are enforced through a marketdata policy service instead of HTTP handler branching.

### Code References
- HTTP upload/status handlers:
  - [daily_uploads.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/daily_uploads.go)
- Multipart decode/error mapping:
  - [codec.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/codec.go)
  - [endpoint.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/endpoint.go)
- Upload domain model and status transitions:
  - [daily_upload.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/daily_upload.go)
- Background job queue/reconcile/processing:
  - [daily_upload.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/daily_upload.go)
- Persistent storage implementation:
  - [sqlx_daily_upload_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_daily_upload_store.go)
- File persistence implementation:
  - [disk.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/disk.go)
- Route registration and job wiring:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)

### Validation Coverage
- Handler tests cover invalid UUID, upload not found, extension validation, provider compatibility checks, and response payload behavior.
  - [daily_uploads_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/daily_uploads_test.go)
- Job tests cover queue deduplication and reconciliation-driven processing with inserted/duplicate counter behavior.
  - [daily_upload_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/daily_upload_test.go)

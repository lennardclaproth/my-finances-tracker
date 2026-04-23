## [016] Background Job Orchestration

### Summary
The API runs a managed background job set for imports, daily upload processing, agent-based tagging, and bulk tag mutations, all started alongside the HTTP server and stopped via shared context cancellation.

### Why This Feature Exists
Long-running ingestion and enrichment workflows should not block request/response paths; job orchestration keeps endpoints fast while ensuring queued work progresses and recovers after restarts.

### Composition and Lifecycle
- Jobs are constructed in `setupJobs(...)` and passed to `jobs.Manager`.
- The server and manager run concurrently under one `errgroup`.
- Shutdown is coordinated by context cancellation (SIGINT/SIGTERM via `signal.NotifyContext`).

### Managed Job Set
1. `ImportJob`
- Queue-backed import processing (`/import/csv` enqueue path).
- Reconciles pending imports from DB on startup and periodically.
- Parses and persists cashflow (and portfolio for brokerage vendors).
- Publishes `TransactionsCreated.v1` and `ImportCompleted.v1` events.

2. `DailyUploadJob`
- Queue-backed processing for manual daily uploads.
- Reconciles pending uploads from DB and claims work atomically.
- Parses rows, persists dailies, tracks inserted/duplicate/error counters.

3. `TaggerJob`
- Polls untagged cashflow transactions.
- Calls agent runner for auto-tagging.
- Applies fallback tag (`unk`) on agent client errors.
- Uses exponential backoff when no work or failures occur.
- Registration is conditional on `agent.enabled`; when disabled, this job is omitted.

4. `BulkTagJob`
- Worker-pool job for large filtered tag updates.
- Processes queued filter/tag tasks asynchronously.
- Publishes `BulkTagCompleted.v1` after successful updates.

### Queue and Recovery Semantics
- Import and daily-upload jobs use bounded queues with enqueue deduplication (`inQueue` + `inFlight`).
- If enqueue fails after persistence (queue full or transient issues), periodic DB reconciliation retries scheduling.
- Bulk-tag queue is bounded and supports configurable worker count (capped to max).

### Concurrency and Error Handling
- Manager starts each job in its own goroutine and logs job exit errors.
- Import/daily jobs process one dequeued item at a time per job instance while reconciling additional pending work.
- Job-specific failures are logged and reflected in domain state when applicable (for example import/upload marked failed).

### Observability Integration
- Job queues preserve propagation headers from producer context.
- Processing starts APM transactions/spans with operation names such as `job.import.process`, `job.daily_upload.process`, `job.bulk_tag.process`, `job.tagger.process`.
- Structured logs include safe allowlisted fields (`import_id`, `upload_id`, `worker_id`, etc.).

### Important Notes
- Manager does not currently implement restart/backoff of failed jobs; it logs failures and waits on context lifecycle.
- Queue sizes and reconcile intervals are currently hardcoded in composition root defaults (`256`, `5s`, etc.).
- Exported job-manager contracts are documented in code to keep orchestration intent explicit.
- Import and daily-upload domain processing logic is delegated to feature-layer processors; job workers focus on queueing/reconciliation concerns.

### Code References
- Job manager and lifecycle:
  - [manager.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/manager.go)
- Import job:
  - [import.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/import.go)
- Daily upload job:
  - [daily_upload.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/daily_upload.go)
- Tagger job:
  - [tagger.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/tagger.go)
- Bulk tag job:
  - [bulk_tag.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/bulk_tag.go)
- Job/server composition:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)

### Validation Coverage
- Import queue tests cover deduped enqueueing, reconciliation, queue-full catch-up, and pending processing.
  - [import_queue_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/import_queue_test.go)
- Daily upload tests cover queue deduplication and reconciliation-driven processing.
  - [daily_upload_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/daily_upload_test.go)
- Bulk tag tests cover worker bounds, queue-full behavior, and queued processing.
  - [bulk_tag_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/bulk_tag_test.go)
- Tagger tests cover backoff/default interval behavior.
  - [tagger_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/jobs/tagger_test.go)

# Traceability Runbook

## Trace a failed request end-to-end

1. Start with `request_id` or `correlation_id` from API response headers.
2. Query logs in Kibana:
   - `request_id: "<value>" OR correlation_id: "<value>"`
3. Capture `trace.id` and open the APM trace view.
4. Follow `operation` path:
   - `http.*` -> `bus.publish.*` -> `bus.consume.*` -> `job.*`
5. Validate related business IDs:
   - `account_id`, `vendor_id`, `listing_id`, `import_id`, `upload_id`.

## Diagnose stuck async flows

1. Query jobs with failures:
   - `component: "job" AND outcome: "failure"`
2. Query bus consumers:
   - `component: "bus" AND outcome: "failure"`
3. Pivot by `correlation_id` and `trace.id`.
4. Check queue growth:
   - `operation: "job.import.process" OR operation: "job.daily_upload.process"`.
5. If repeated parse failures:
   - inspect upload/import source rows and parser-specific validation errors.

## Identifier priority

1. `request_id` (best entry from API/UI)
2. `correlation_id` (cross-async continuity)
3. `trace.id` (APM deep dive)
4. Business IDs (`import_id`, `upload_id`, `account_id`, `listing_id`, `vendor_id`)

## Data policy

- Allowed fields are technical metadata and curated IDs only.
- Free text, descriptions, notes, raw payloads, and personal data must not be logged/labeled.


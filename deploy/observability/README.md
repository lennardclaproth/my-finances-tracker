# Observability Pack (Elastic APM + ELK)

This folder contains the traceability assets for Kibana:

- `kibana/dashboards.ndjson`: saved objects for request, async pipeline, and correlation lookup dashboards.
- `kibana/alerts.ndjson`: saved objects for latency, error-rate, job-failure, and backlog alerts.

## Import

1. Open Kibana: `Stack Management -> Saved Objects -> Import`.
2. Import `dashboards.ndjson`.
3. Import `alerts.ndjson`.
4. Rebind index patterns if your deployment uses custom data stream names.

## Field conventions expected

- `trace.id`
- `transaction.id`
- `span.id`
- `request_id`
- `correlation_id`
- `operation`
- `component`
- `outcome`
- `account_id`
- `vendor_id`
- `listing_id`
- `import_id`
- `upload_id`


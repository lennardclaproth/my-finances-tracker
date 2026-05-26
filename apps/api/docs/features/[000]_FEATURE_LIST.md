# Feature List

- [001] Account management
- [002] CSV Imports
- [003] Realtime Updates
- [004] Portfolio Management
- [005] Cashflow Management
- [006] Asset Management
- [007] Asset Sync
- [008] Vendor Listing
- [009] Listing Management
- [010] EOD Data
- [011] Provider Support
- [012] Health Checks
- [013] API Docs
- [014] Event Handlers
- [015] Database Bootstrap
- [016] Observability

## Users

### [001] Account Management

Users can create and list accounts. Accounts are the shared scope for imports, cashflow, portfolio, and assets.

### [002] CSV Imports

Users can upload vendor CSV files for an account. Uploads are accepted quickly, stored durably, and processed asynchronously in the background.

### [003] Realtime Updates

Clients can subscribe to account-scoped WebSocket notifications. These updates tell the UI when async imports, portfolio rebuilds, bulk tags, and asset snapshot rebuilds have completed.

### [004] Portfolio Management

Users can read portfolio snapshots, current positions, and transaction history. Users can also request asynchronous portfolio rebuilds and create manual portfolio transactions.

### [005] Cashflow Management

Users can query transactions with filters, sorting, and pagination. Users can create manual cashflow transactions, view monthly and tag analytics, tag transactions, bulk tag filtered transactions, and ignore or unignore transactions.

### [006] Asset Management

Users can manage account-scoped asset classes and asset items. Users can set or adjust item worth, inspect class details and history, and read total asset snapshots.

### [007] Asset Sync

Portfolio rebuild completion syncs portfolio worth into a read-only portfolio asset class and rebuilds total asset snapshots.

## Admin

### [008] Vendor Listing

Admin workflows can list active vendors. The response includes import capability metadata so clients can show whether CSV imports are disabled for a vendor.

### [009] Listing Management

Admins can create, update, list, and search market data listings. Listings are the canonical market instruments used by portfolio transactions and daily pricing.

### [010] EOD Data

Admins can retrieve EOD price history, upload manual EOD market data files, and check upload processing status.

### [011] Provider Support

The API bootstraps API and manual market data providers. Provider mode controls listing synchronization behavior and manual upload eligibility.

## Platform Support

### [012] Health Checks

The API exposes a lightweight endpoint for checking whether the service is running.

### [013] API Docs

Swagger and OpenAPI documentation publish the generated API contract and serve the interactive Swagger UI.

### [014] Event Handlers

Messaging handlers project account events, trigger portfolio and asset rebuild flows, and bridge completion events into realtime client notifications.

### [015] Database Bootstrap

Migrations and bootstrap seeding prepare the database schema and baseline domain records before the API serves traffic.

### [016] Observability

Observability support provides request identifiers, structured logs, trace propagation, and redaction controls across HTTP, events, and jobs.

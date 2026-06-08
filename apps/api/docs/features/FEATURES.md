# Features

A concise overview of what My Finances Tracker does. Each entry is the application
capability, not its implementation; see the code under `apps/api/internal/<feature>`
and `apps/api/transport` for details.

Each capability links to a **detailed feature doc** (with class, ERD, state-machine, and
sequence diagrams) in this same `features/` directory. Tightly-related capabilities share a
doc (e.g. all cashflow and all market-data features).

## Core (user-facing)

### Account management
Users create and list accounts. An account is the shared scope that imports, cashflow,
portfolio, and assets all hang off of.

→ Details: [Account management](%5B003%5D_ACCOUNT_MANAGEMENT.md)

### CSV imports
Users upload a vendor CSV for an account and it is accepted immediately, stored durably,
and processed asynchronously. Three explicit import types are supported — cashflow,
portfolio, and end-of-day market data — with vendor-specific parsers (DeGiro, ING, N26
for cashflow; DeGiro for portfolio; BrandNewDay for EOD).

→ Details: [CSV imports](%5B002%5D_CSV_IMPORTS.md)

### Cashflow insights
Users query bank/payment transactions with filtering, sorting, and pagination, and add
manual transactions. They can view monthly and per-tag analytics, tag transactions
(individually, by selection, or by filter), and ignore/unignore transactions to exclude
them from totals.

→ Details: [Cashflow](%5B009%5D_CASHFLOW.md)

### Portfolio management
Users read portfolio snapshots, current positions, and transaction history, add manual
portfolio transactions, and request asynchronous portfolio rebuilds that recompute
positions and snapshots from transactions and market data.

→ Details: [Portfolio](%5B013%5D_PORTFOLIO.md)

### Asset management
Users manage account-scoped asset classes and the items within them, set or adjust each
item's worth, inspect class details and mutations, and read total net-worth snapshots
across all classes.

→ Details: [Assets](%5B024%5D_ASSETS.md)

### Asset sync
When a portfolio rebuild completes, the computed portfolio worth is synced into a
read-only "portfolio" asset class and the total asset snapshots are rebuilt, so net worth
always reflects the latest portfolio value.

→ Details: [Assets](%5B024%5D_ASSETS.md)

### Realtime updates (WebSocket)
Clients subscribe to an account-scoped WebSocket (`/ws/accounts/{account_id}`) and receive
push notifications when long-running work finishes — import completed, portfolio rebuilt,
and asset snapshots rebuilt — so the UI can refresh without polling.

→ Details: [Realtime updates](%5B001%5D_REALTIME_UPDATES.md)

## Market data & admin

### Listing management
Admins create, update, list, and search market-data listings. Listings are the canonical
instruments referenced by portfolio transactions and end-of-day pricing.

→ Details: [Market data](%5B005%5D_MARKET_DATA.md)

### EOD market data
Admins retrieve end-of-day price history, upload manual EOD price files, and check the
processing status of an upload.

→ Details: [Market data](%5B005%5D_MARKET_DATA.md)

### Provider support
The API supports both API-based market-data providers (MarketStack, Alpha Vantage) and a
manual provider; the configured provider mode controls listing synchronization and whether
manual EOD uploads are allowed.

→ Details: [Market data](%5B005%5D_MARKET_DATA.md)

### Vendor listing
Lists the active import vendors along with capability metadata, so clients can show which
vendors support CSV imports and which are disabled.

→ Details: [Vendor listing](%5B004%5D_VENDORS.md)

## Platform

### Event-driven messaging
An in-memory event bus fans domain events out to many features — for example, creating an
account projects it into portfolio, cashflow, assets, and importer, and completion events
bridge into the realtime WebSocket notifications.

→ Details: [Event-driven messaging](%5B017%5D_EVENT_DRIVEN_MESSAGING.md)

### Observability
Every request carries an identifier and structured logs with trace propagation and
sensitive-data redaction, wired through Elastic APM across HTTP and SQL.

→ Details: [Platform operations](%5B018%5D_PLATFORM_OPERATIONS.md)

### Database bootstrap & migrations
The service runs goose migrations (mirrored for Postgres and SQLite) and seeds baseline
records on startup before serving traffic.

→ Details: [Platform operations](%5B018%5D_PLATFORM_OPERATIONS.md)

### Health & API docs
A lightweight `/health` endpoint reports liveness, and Swagger/OpenAPI publishes the
generated API contract with an interactive UI at `/swagger/`.

→ Details: [Platform operations](%5B018%5D_PLATFORM_OPERATIONS.md)

### Web app
A SvelteKit (Svelte 5) frontend in `web/`; it is early-stage — currently an Atomic-Design
component library built in Storybook, not yet consuming the API — and most application
logic lives in the Go backend.

→ Details: [Web app](%5B020%5D_WEB_APP.md)

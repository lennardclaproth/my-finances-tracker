# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- [001] Added dynamic client updates with web sockets.
- [002] Added CSV import ingestion so users can upload vendor files, persist import jobs, and process them asynchronously in background workers.
- [003] Added account management endpoints to create and list accounts used across cashflow imports and portfolio calculations.
- [004] Added active vendor listing support so the UI can present import-capable vendors and enforce vendor-level import rules.
- [005] Added market data provider support with API and manual ingestion modes, including provider bootstrap from configuration and environment keys.
- [006] Added listing management endpoints to create, partially update, list, and search market listings with pagination and validation.
- [007] Added daily market data retrieval with filtering by listing or symbol, date-range support, pagination, and sort ordering.
- [008] Added asynchronous daily upload processing so users can upload manual dailies files and poll per-upload processing status and row errors.
- [009] Added cashflow transaction querying with rich filters (text, source, direction, tags, ignored state, date range), sorting, and pagination.
- [010] Added cashflow analytics endpoints for monthly incoming/outgoing/net trends and tag distribution aggregates.
- [011] Added cashflow tagging workflows for single-transaction tagging, selection-based bulk tagging, and filter-based tagging with async bulk job fallback for large sets.
- [012] Added cashflow ignore workflows so users can mark selected or filtered transactions as ignored or not ignored.
- [013] Added portfolio rebuild requests over the internal event bus so rebuilds are decoupled and processed asynchronously.
- [014] Added portfolio read APIs for snapshots, positions, and transactions with account checks, filtering, sorting, and pagination.
- [015] Added manual portfolio transaction creation for BUY/SELL/DIVIDEND/TAX/FEE/CASH flows without forcing immediate rebuild triggers.
- [016] Added background job orchestration for import processing, daily uploads, auto-tagging via agent runner, and bulk tag processing.
- [017] Added event-driven messaging handlers so account creation and imported transactions propagate to portfolio, cashflow, and importer projections.
- [018] Added structured observability with request and correlation IDs, operation-level logging, and APM trace propagation across HTTP and jobs.
- [019] Added automatic database migration execution and bootstrap seeding for vendors, default accounts, and providers at startup.
- [020] Added a SvelteKit web app experience with cashflow and portfolio pages, plus admin listings and admin dailies pages guarded by admin mode.
- [021] Added generated Swagger/OpenAPI documentation and Swagger UI serving for API discoverability and client integration.
- [022] Added configurable agent auto-tag background control so `TaggerJob` can be enabled or disabled via `agent.enabled`.
- [023] Added manual cashflow transaction creation with bulk entry support from the cashflow page, including account-scoped API validation and manual-source tagging (`source=manual[:vendor]`).
- [024] Added account-scoped asset management with manual asset classes, multiple assets per class, worth history, growth views, and class/item worth mutation APIs plus web UI workflows.
- [025] Added portfolio-linked asset class synchronization so portfolio worth is projected into a read-only assets class only when portfolio rebuild completion events are processed.
- [024] Added backend-owned daily asset snapshot projection (`GET /assets/snapshots`) with persisted account/day total-worth points and async rebuild events.
- [003] Added `GET /accounts` (list) to complete the account-management endpoints; the web app resolves its active account id from it via a new `accountStore` instead of a hard-coded constant.
- [026] Added CORS support to the API — configurable `server.cors_allowed_origins` (default the SvelteKit dev origin `http://localhost:5199`) applied outermost with OPTIONS preflight handling, so the browser frontend can call the API cross-origin.
- [027] Added a complete frontend write service layer (`apiSend` JSON + `apiUpload` multipart) covering every mutation endpoint — cashflow create/tag/ignore, asset class/asset/worth mutations, listing create/update, manual portfolio transactions + rebuild, and cashflow/portfolio/EOD file imports — each with a mock fallback.

### Changed
- [020] Began the SvelteKit visual rebuild per `web/docs/DESIGN_PLAN.md` (Phases 0–2): fixed the theme split so the app and Storybook load the same `app.css`, defined the `taupe` token ramp and valid `@font-face` rules, added a documented z-index scale and a framework-agnostic `charts/theme.ts`, and added foundational molecule primitives (popover, dialog, tabs, sortable-header, breadcrumb, search-input).
- [020] Added a typed, API-shaped stub-data layer (`$lib/api` types + money helpers, `$lib/data` fixtures, `$lib/services`) with a mocks-on-by-default flag so the web app runs and is testable independently of the Go API.
- [020] Added Phase 3 web design-system molecules: a calendar core with single date-picker and dual-month date-range-picker (presets, min/max), popover-based filters (text/direction/select/visibility) via a shared filter-popover shell, animated action- and account-menus, and an async listing-search-select autocomplete.
- [020] Added Phase 4 charts (Chart.js 4 dependency): a `charts/*` layer (theme palette, registration, option builders, dashed-grid/zero-line plugin, area gradients, and a shared drag range-selection plugin emitting ISO `{from,to}`) plus `TimeSeriesChart` (line/bar/area combo, range-select, below-zero flip) and `DonutChart` (click-to-filter, "+N more" popover); all canvas code is client-guarded so SSR builds stay clean.
- [020] Added Phase 5 toast system: a single runes `toast` store (tone helpers, auto-dismiss/sticky, status→intent heuristic) and one `ToastHost` organism that renders the `alert` molecule top-right with fly/flip transitions and stacking.
- [020] Added Phase 6 organisms: `KpiRow`, a generic `DataTable` (sticky header, indeterminate multi-select, sort, per-column filter slots, loading/empty/error, footer slot), `FooterBar` (pagination + bulk actions), a right slide-in `Drawer`, `TopNavbar`, and concrete `CashflowTransactionsTable`, `AssetClassDrawer`, and `TransactionFormModal` examples composing those engines.
- [020] Added Phase 7 templates + state plumbing: `AppShellTemplate` and `PageContentTemplate`, pure URL-as-state helpers (`url/routeQuery.ts`) plus a SvelteKit `pushQuery` wrapper, a localStorage-backed `adminMode` store, and a mock/SSR-safe `connectRealtime` WebSocket service with a debounced refresh.
- [020] Added Phase 8 pages: SvelteKit routes for `/cashflow` (fully URL-state-driven with analytics charts, transactions table, create modal, toasts, realtime, chart-range→date-filter), `/portfolio`, `/assets` (+ asset-class drawer), and admin-gated `/admin/listings` and `/admin/dailies`; `/` redirects to `/cashflow` and the welcome page is removed. Runs against the mock layer with no backend.
- [020] Reworked the web UI for visual coherence: a single reusable `AnalyticsCard` (Panel-based) replaces the ad-hoc chart cards across cashflow/assets/portfolio; `TopNavbar` now floats as separate cards (left menu-button `NavMenu` for cross-page navigation with active-route highlighting, centered bare search, and a global date-range selector) and uses the `Heading` typography component for larger page titles; filter-popover triggers became ghost rounded pills with a blue active state; modal footers use semantic `success`/ghost buttons instead of primary; and the chart drag-selection band is recolored blue. The global date range now drives every page's charts and donuts — the cashflow monthly/tag-distribution mocks were made date-aware so the trend zooms and the donuts re-aggregate on selection.
- [027] Wired the frontend write UIs to the API (they previously only toasted): the cashflow "New transaction" modal and bulk-tag action, the admin listings create dialog, a new assets "New asset class" dialog, and a new portfolio "New transaction" modal plus a portfolio-rebuild navbar action now persist to the backend.
- Added Makefile targets to run the app end-to-end (`db-up` / `db-up-all` / `db-down`, `web-install` / `web-dev` / `web-build` / `web-env`) and a `web/.env.example`; rewrote the root `README.md` into a full local run guide and refreshed stale `CLAUDE.md` notes (single `cmd/my-finances-tracker` entrypoint with a green build, runtime theme loading, real app routing, working `make web-lint`).

### Fixed
- [024] Fixed `GET /assets/snapshots` date handling: it used the literal layout `"2026-01-02"` instead of Go's reference layout `"2006-01-02"` when parsing `from`/`to` and formatting the snapshot date (so date-range filtering was wrong), and added the missing early return when query decoding fails.
- Completed the committed `apps/api/go.sum` (was missing the two `github.com/gorilla/websocket` v1.5.3 hashes) so a clean checkout builds and vets under the default `-mod=readonly`.
- [020] Phase 9 cross-cutting verification fixes: raised muted text from `slate-400` to `slate-500` (WCAG AA contrast, matching the design's "muted = slate-500"), added a default document `<title>`, and enabled the Storybook a11y addon in error mode — a live axe-core scan of the cashflow/portfolio pages now reports zero violations; responsive (navbar stacking, analytics grid collapse) and the single z-index scale were verified.
- [020] Removed the redundant `sortable-header` molecule (superseded by `DataTable`'s integrated sort+filter header; its `SortDirection` type moved into `data-table.types`) and two dead helper exports (`centsToNumber`, `rampColor`) after a usage audit.
- [020] Fixed unclickable checkboxes: the `Checkbox` atom's hidden `sr-only` input is now a transparent full-size overlay, so standalone checkboxes (e.g. `DataTable` select-all and per-row selection) respond to clicks instead of relying on an external `<label>`. Also enlarged the too-small `DataTable` sort chevrons (`size-3`→`size-4`).
- [002] Clarified `/import/csv` feature documentation to describe the raw UUID response format and aligned import enqueue-degraded logging to warning level.
- [023] Aligned the cashflow create-transaction floating action button behavior/placement with the portfolio and assets pages.
- [024] Refined asset-management UX with top-card growth/distribution charts, table-level class settings actions, and table styling aligned to the existing app pattern.
- [024] Updated assets analytics/presentation by aligning chart styles with existing portfolio/cashflow charts, computing class growth percentages from inception, removing header subtitle/archive toggle, and paginating class history in the slider.
- [025] Changed portfolio-linked asset syncing to rebuild class growth history from portfolio snapshots and set current worth from the latest snapshot on rebuild.
- [024] Added assets-page date-range controls via the shared navbar date selector and applied date scoping to assets growth/history views, while aligning growth indicators with portfolio badge styling and normalizing drawer table/card styling with app table/card patterns.
- [025] Clarified portfolio-linked asset projection semantics by explicitly using portfolio snapshot market value as sync input for the read-only portfolio class.
- [023] Updated manual cashflow transaction date entry to use the shared custom calendar selector used by navbar date controls.
- [015] Updated manual portfolio transaction date entry to use the shared custom calendar selector used by navbar date controls.
- [024] Changed assets-page top growth/current-worth rendering to consume server snapshot points directly instead of client-side class-growth aggregation.
- [025] Changed portfolio rebuild flow to trigger account-level assets snapshot rebuild requests and realtime `assets.rebuilt` notifications on completion.
- [019] Consolidated the database schema into a single from-scratch initial migration for SQLite and PostgreSQL, tightening column sizes/types (`VARCHAR(n)` limits), aligning `NULL`/`NOT NULL` with the Go domain model, completing foreign keys/unique constraints/`CHECK`s, and redesigning indexes around the storage layer's real query patterns.
- [019] Corrected schema/code drift in the initial migration: included the `processing` import status in the status `CHECK`, removed the unused `import.accounts` projection table, and changed `account.accounts.external_id` from `UUID` to a nullable string.

### Fixed
- [001] Fixed websocket feature hygiene by adding realtime endpoint Swagger coverage and explicit client-side websocket error handling/logging.
- [001] Fixed websocket `/ws/accounts/{account_id}` upgrade reliability when request logging middleware is enabled by preserving response hijacking support.
- [001] Fixed websocket observability noise by excluding `/ws/accounts/{account_id}` handshake requests from APM HTTP server transactions.
- [018] Fixed duplicated trace/request/correlation identifiers in request-completed logs by preventing double context-field attachment.
- [003] Fixed account management validation coverage by adding API integration tests for create validation, duplicate conflicts, and account list ordering.
- [004] Fixed vendor bootstrap reliability across SQLite and PostgreSQL by mapping duplicate vendor-name conflicts consistently to `ErrVendorAlreadyExists`.
- [005] Fixed provider/bootstrap maintainability by documenting exported provider and provider-store contracts used by ingestion/bootstrap flows.
- [006] Fixed listing-management API maintainability by documenting exported listing request/response DTO contracts.
- [007] Fixed dailies retrieval API maintainability by documenting exported request/response DTO contracts.
- [008] Fixed daily-upload API maintainability by documenting exported upload/status request and response DTO contracts.
- [008] Fixed daily-upload layering by moving provider/parser manual-upload eligibility rules from HTTP handler code into marketdata policy service.
- [009] Fixed cashflow-transaction query API maintainability by documenting exported request/response DTO contracts.
- [010] Fixed cashflow analytics API maintainability by documenting exported analytics request/response DTO contracts.
- [011] Fixed cashflow-tagging API maintainability by documenting exported mutation request/response DTO contracts.
- [011] Fixed cashflow-tag filter orchestration layering by moving async cutoff and enqueue policy from handler logic into a cashflow feature service.
- [012] Fixed cashflow-ignore API maintainability by documenting exported mutation request/response DTO contracts.
- [013] Fixed portfolio-rebuild API maintainability by documenting exported async command/acceptance DTO contracts.
- [014] Fixed portfolio-read API maintainability by documenting exported request/response DTO contracts.
- [014] Fixed portfolio-read layering by moving derived snapshot/amount projection calculations from handlers into portfolio feature helpers.
- [015] Fixed manual-portfolio-transaction API maintainability by documenting exported create request/response DTO contracts.
- [016] Fixed job-orchestration maintainability by documenting exported manager/router HTTP plumbing contracts.
- [016] Fixed job-layer boundaries by delegating import and daily-upload domain processing from workers into feature-layer processors.
- [017] Fixed event-contract maintainability by documenting exported bus message payloads and topics for cashflow/portfolio/import workflows.
- [018] Fixed observability consistency by routing `/health` and `/swagger/` through request-logging middleware and documenting exported HTTP middleware contracts.
- [019] Fixed migration/bootstrap maintainability by documenting exported startup provider-bootstrap contracts.
- [020] Fixed web-app integration maintainability by documenting shared API DTO contracts used by frontend service consumers.
- [021] Fixed Makefile lint/vet/fmt execution from repository root by running Go tooling within `apps/api`.
- [021] Fixed Swagger route observability consistency by applying request-logging middleware to `/swagger/`.
- [024] Fixed asset worth adjustment rejecting every request by correcting the direction validation (`||` → `&&`) so only directions that are neither `INCREASE` nor `DECREASE` are rejected.
- [024] Fixed `/assets/classes/{class_id}` growth data to read from the newest bounded history window, preventing stale oldest-window graph values.
- [024] Fixed assets-page “correct then drop” top worth behavior by removing client recomputation drift after graph load.

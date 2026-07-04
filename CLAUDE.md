# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository shape

Monorepo with two apps:
- `apps/api` — Go backend (Go module `github.com/lennardclaproth/my-finances-tracker`, single-module `go.work`). This is where almost all logic lives.
- `web` — SvelteKit 2 / Svelte 5 frontend. The directory is `web/` (there is no `apps/web`). The Makefile's `WEB_DIR := ./web` and the `web-*` targets (`make web-dev`, `make web-build`, `make web-install`, `make web-env`, `make web-lint`) target it correctly.

Authoritative agent guidance already exists in [AGENTS.md](AGENTS.md) (root, repo-wide) and [apps/api/AGENTS.md](apps/api/AGENTS.md) (backend-specific). Read both — the rules below summarize and reconcile them, but the AGENTS.md files are the source of truth for conventions.

## Build & entrypoint (the refactor has landed)

The backend builds, boots, and serves the full HTTP surface. There is a **single entrypoint**,
`cmd/my-finances-tracker/main.go` (the old `cmd/server` was removed); the Makefile and the VS Code
"Launch Server" config target it. `make build` / `make run` / `make test` work. `go build ./...` and
`go vet ./...` are green from `apps/api`.

Follow the current structure: `internal/eventbus` (not the removed `internal/bus`), `transport/http`
(not the removed `internal/http`), feature packages under `internal/`.

Worktree note: a bare `go build ./...` from inside a git worktree needs `GOWORK=off` because the repo
`go.work` resolves to the main checkout; `make build`/`make run`/VS Code launch are unaffected.

## Commands

The root `Makefile` is the source of truth (run `make help`). See [README.md](README.md) for the
end-to-end run guide. Key targets run from the repo root:
- `make build` / `make run` — build/run the API (`cmd/my-finances-tracker`, port from `config.yaml`).
- `make db-up` / `make db-down` — start/stop local Postgres via `deploy/docker/compose.dev.yaml`
  (`make db-up-all` also starts Elasticsearch/Kibana/APM). The API auto-creates the DB + migrates on start.
- `make test` — `go test ./apps/api/...` (unit). `make test-integration` runs the `integration`-tagged suite.
- `make lint` — `go fmt` + `go vet` + `golangci-lint` (cd's into `apps/api`).
- `make swagger` — regenerate OpenAPI docs into `apps/api/docs` via `swag init -g cmd/my-finances-tracker/main.go`. Never hand-edit generated files under `apps/api/docs`.
- `make migrate-up` / `make migrate-down` / `make migrate-create name=<n>` — goose migrations (needs `DATABASE_URL`).
- `make web-install` / `make web-dev` / `make web-build` / `make web-env` — frontend install / dev (`:5199`) / build / seed `web/.env`.
- `make install-tools` — installs `swag`, `goose`, `air`, `golangci-lint`. `make dev` runs the API with `air` hot reload.

Run a single Go test: `go test -run TestName ./internal/<pkg>/...` (from `apps/api`).

Frontend (inside `web/`): `npm run dev`, `npm run build`, `npm run check` (svelte-check — the type gate), `npm run lint` (prettier + eslint), `npm run format`, `npm run storybook`. There is no top-level `test` script; component tests run through Storybook's Vitest addon.

Local config: the server reads `apps/api/config.yaml` (relative to cwd, so run from `apps/api`) plus a `.env` (VS Code "Launch Server" uses `apps/api/.env`; `make env` seeds it from `config.example.env`). The frontend reads `web/.env` (`VITE_API_URL`, `VITE_USE_MOCKS`); `make web-env` seeds it from `web/.env.example`. With no `VITE_API_URL` the frontend runs on mock fixtures.

## Backend architecture (`apps/api`)

**Package-by-feature.** Domain/feature packages under `internal/`: `account`, `cashflow`, `portfolio`, `assets`, `marketdata`, `vendor`, `importer`, `files`. Supporting packages: `storage`, `eventbus`, `notify` (WebSocket fan-out), `bootstrap`, `logging`, `observability`, `config`, plus shared utilities `money`, `date`, `sorting`. (`internal/messaging` is an empty leftover; the event-bus *handlers* live under `transport/messaging/handlers`, not `internal/`.)

**Application boundary = `Commands` + `Queries`.** Each feature exposes a `Commands` type (writes; mutate state and publish events) and a `Queries` type (reads). These depend on small *feature-owned* interfaces (e.g. `creator`, `queryStore`) that `internal/storage` implements directly — do not introduce adapter interfaces or duplicate domain types to avoid touching the real boundary. Other collaborators (builders, syncers, processors, services) are used where that's the established nearby pattern.

**Event bus** is `internal/eventbus` (the canonical one — *not* the removed `internal/bus`). In-memory implementation under `internal/eventbus/memory`. Publish via `Commands` (e.g. `account.Commands.Create` publishes `AccountCreated`); subscribe with typed handlers. Event handlers live in `transport/messaging/handlers/<feature>` (currently `assets`, `importer`, `portfolio`) and stay thin — they react and delegate to feature collaborators. The startup wiring fans one event out to many features (e.g. `AccountCreated` → portfolio, cashflow, assets, and importer account projections).

**Storage** (`internal/storage`) uses `sqlx`, one `sqlx_<aggregate>_store.go` per store. Patterns to follow:
- `DB` wraps `*sqlx.DB` and supports both Postgres and SQLite from the same code.
- Transactions: `DB.WithTx(ctx, fn)` stashes the `*sqlx.Tx` in context; stores call `db.GetExecutor(ctx)` to transparently use the tx or the pool.
- `qualifyTable` prefixes a schema (e.g. `cashflow.transactions`) **only on Postgres**; SQLite is schemaless and gets the bare table name. Schema/table names are constants in `db.go`.
- Follow existing rebind / duplicate-mapping / `sql.ErrNoRows` handling.

**HTTP transport** lives in `transport/http` (module path `.../transport/http` — **not** under `internal/`); handlers are organized per feature under `transport/http/handlers/<feature>`. New/refactored handlers:
- Return a plain `http.HandlerFunc`, define **feature-local request/response DTOs**, and use the codec helpers in `transport/http` (imported as `httpx`): `JSONDecode`, `JSONEncode`, `DecodeQuery`, `DecodeMultipartFile`, `WriteDecodeError`.
- Carry Swagger annotations on the handler constructor.
- Keep handlers thin: decode, validate transport input, map to feature inputs, map known feature errors to HTTP status, encode. No business rules or orchestration in handlers.
- Do **not** use the legacy `internal/http.Endpoint` wrapper or import the (removed) `internal/http`.
- `Router` is a thin slice over `http.ServeMux`; routes use Go 1.22 method+pattern syntax (`"POST /accounts"`) and are registered with `HandleWithMiddleware`. `Server` wraps the mux with Elastic APM, request-identifier, and CORS middleware (`WithCORS`, allowing `server.cors_allowed_origins` from config; applied outermost so OPTIONS preflight is answered before routing) and does graceful shutdown.

**Importer subsystem** (`internal/importer`): explicit import types `cashflow` / `portfolio` / `eod`, each with a `parsers/` package (vendor-specific: `degiro`, `ing`, `n26`, `brandnewday`) and a `processor.go` implementing the `Processor` interface (`Process(ctx, *Import) (ProcessResult, error)`). Parser factories resolve a parser by vendor/source. `Import` is the durable record with a status lifecycle (`pending → processing → completed/failed/...`).

**Market data** (`internal/marketdata`): providers (MarketStack, Alpha Vantage), listings, end-of-day "eods", a client, and a syncer.

**Config & observability:** `internal/config` reads `config.yaml` and overlays env vars; provider API keys come from env only (`MARKETSTACK_API_KEY`, `ALPHA_VANTAGE_API_KEY`) and it exports Elastic APM env vars. Elastic APM is wired through HTTP (`apmhttp`) and SQL (`apmsql`); use the `internal/logging` slog wrapper (`logging.Logger`) for all logging with appropriate levels, and never log sensitive data.

**Reuse shared packages** before writing helpers: `internal/money` (currency/price/decimal), `internal/date` (parsing/ranges/formatting), `internal/sorting` (sort fields/directions). Put feature-level errors in that package's `errors.go`.

## Frontend architecture (`web`)

SvelteKit 2 + Svelte 5, Vite, TailwindCSS 4, TypeScript, Storybook 10, Vitest + Playwright, ESLint + Prettier.

Follow **Atomic Design strictly**: components live under `src/lib/components/` (`atoms`, `molecules`, `organisms`, and `templates` — e.g. `app-shell`, `page-content`; a `pages` tier can be added as the app grows). Each component folder co-locates `Component.svelte`, `Component.stories.svelte`, `component.types.ts`, and `component.variants.ts` (Tailwind variant maps). Preserve and extend this structure; don't bypass it. Handle errors explicitly and surface user-relevant failures in the UI (Elastic APM error reporting may be extended for server-side observability).

Data access goes through `src/lib/services/*` (one module per feature) over the fetch client in `src/lib/api` (`apiGet` / `apiSend` JSON, `apiUpload` multipart). Every service has a mock branch (fixtures under `src/lib/data/fixtures`) and a live branch; `VITE_API_URL` + `VITE_USE_MOCKS` select which. The active account id comes from the `accountStore` (`src/lib/stores/account.svelte.ts`), which resolves it from `GET /accounts` — don't hard-code `DEMO_ACCOUNT_ID` in pages.

## Database & migrations

Two databases are supported (`sqlite3` for local dev, `postgres` for deployed). Migrations are **mirrored**: every change needs matching files under both `apps/api/migrations/postgres/` and `apps/api/migrations/sqlite/` (goose, timestamped). In dev the app auto-creates the DB and runs migrations on startup. Do **not** modify existing migrations or add/change deps, Docker, or infra files without explicit approval — add new timestamped migrations instead.

## Change & documentation discipline

From the AGENTS.md files — these are enforced expectations, not suggestions:
- Prefer minimal, focused diffs. Don't refactor, rename, or "modernize" unrelated code. Edit existing files over adding new abstractions unless there's clear benefit.
- Feature docs are consolidated in `apps/api/docs/features/FEATURES.md` — a single short overview of every feature. The legacy per-feature `NNN_NAME.md` files under `docs/features/` have been removed. Read `FEATURES.md` before changing a documented feature, and update its entry in the same task when behavior changes. (`apps/api/docs/` also holds the *generated* Swagger output — that part is generated, `features/FEATURES.md` is hand-authored.)
- When a change maps to a feature, also update the root `CHANGELOG.md` under `Unreleased` using Keep a Changelog categories (`Added`/`Changed`/`Fixed`) and reference the feature ID. New feature IDs come from the changelog sequence.
- For API-contract changes, keep Swagger annotations in sync and regenerate with `make swagger` (skippable during the refactor — say so if you skip it).
- Document exported Go types/functions; document non-obvious unexported ones. Avoid comments that merely restate code, and never write numbered step comments.
- For complex multi-step work, the repo convention is to keep a concise `tmp/plan.md` and delete it when done.

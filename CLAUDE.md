# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository shape

Monorepo with two apps:
- `apps/api` — Go backend (Go module `github.com/lennardclaproth/my-finances-tracker`, single-module `go.work`). This is where almost all logic lives.
- `web` — SvelteKit 2 / Svelte 5 frontend. **Note:** the Makefile and `AGENTS.md` refer to `apps/web`, but the actual directory is `web/`. There is no `apps/web`. `make web-lint` will fail; run frontend scripts directly inside `web/`.

Authoritative agent guidance already exists in [AGENTS.md](AGENTS.md) (root, repo-wide) and [apps/api/AGENTS.md](apps/api/AGENTS.md) (backend-specific). Read both — the rules below summarize and reconcile them, but the AGENTS.md files are the source of truth for conventions.

## ⚠️ The backend is mid-refactor and does not currently compile as a whole

The API is being restructured (package renames, moved files). Whole-tree builds currently **fail**, so do not trust `make build` / `make test` / `go build ./...` to pass out of the box:
- There are two `cmd/` entrypoints. `cmd/server/main.go` is the *old* composition root and still imports removed packages (`internal/bus`, `internal/jobs`, `internal/http`, `internal/cashflow/service`). `cmd/my-finances-tracker/main.go` is the *new* entrypoint but is currently an empty `package main` stub.
- Some `internal/` files also still reference removed packages, and a few are empty placeholders (e.g. `internal/transport/http/handlers/import/cashflow.go`).

What works: individual leaf packages compile and their tests pass. To run tests, target compiling packages directly, e.g. `go test ./internal/money/... ./internal/importer/cashflow/parsers/...`. Per `apps/api/AGENTS.md`, full validation is *not* mandatory during this refactor — but state clearly when you skip it, and don't claim a command passed if it didn't.

When old and new code conflict, follow the new structure (`internal/eventbus`, `internal/transport/http`, feature packages under `internal/`), not the stale `cmd/server/main.go`.

## Commands

The root `Makefile` is the source of truth (run `make help`). Key targets run from the repo root:
- `make build` / `make run` — build/run `cmd/server` (currently broken; see above).
- `make test` — `go test ./apps/api/...` (currently fails to compile whole-tree; prefer per-package `go test`).
- `make lint` — `go fmt` + `go vet` + `golangci-lint` (cd's into `apps/api`).
- `make swagger` — regenerate OpenAPI docs into `apps/api/docs` via `swag init -g cmd/server/main.go`. Never hand-edit generated files under `apps/api/docs`.
- `make migrate-up` / `make migrate-down` / `make migrate-create name=<n>` — goose migrations (needs `DATABASE_URL`).
- `make install-tools` — installs `swag`, `goose`, `air`, `golangci-lint`. `make dev` runs with `air` hot reload.

Run a single Go test: `go test -run TestName ./internal/<pkg>/...` (from `apps/api`).

Frontend (inside `web/`): `npm run dev`, `npm run build`, `npm run check` (svelte-check), `npm run lint` (prettier + eslint), `npm run format`, `npm run storybook`. There is no top-level `test` script; component tests run through Storybook's Vitest addon.

Local config: the server reads `apps/api/config.yaml` plus a `.env` (VS Code "Launch Server" uses `apps/api/.env`; `make env` seeds it from `config.example.env`).

## Backend architecture (`apps/api`)

**Package-by-feature.** Domain/feature packages under `internal/`: `account`, `cashflow`, `portfolio`, `assets`, `marketdata`, `vendor`, `importer`, `files`. Supporting packages: `storage`, `eventbus`, `messaging`, `bootstrap`, `logging`, `observability`, `config`, plus shared utilities `money`, `date`, `sorting`.

**Application boundary = `Commands` + `Queries`.** Each feature exposes a `Commands` type (writes; mutate state and publish events) and a `Queries` type (reads). These depend on small *feature-owned* interfaces (e.g. `creator`, `queryStore`) that `internal/storage` implements directly — do not introduce adapter interfaces or duplicate domain types to avoid touching the real boundary. Other collaborators (builders, syncers, processors, services) are used where that's the established nearby pattern.

**Event bus** is `internal/eventbus` (the canonical one — *not* the removed `internal/bus`). In-memory implementation under `internal/eventbus/memory`. Publish via `Commands` (e.g. `account.Commands.Create` publishes `AccountCreated`); subscribe with typed handlers. Event handlers live in `internal/messaging/handlers/<feature>` and stay thin — they react and delegate to feature collaborators. The startup wiring fans one event out to many features (e.g. `AccountCreated` → portfolio, cashflow, assets, and importer account projections).

**Storage** (`internal/storage`) uses `sqlx`, one `sqlx_<aggregate>_store.go` per store. Patterns to follow:
- `DB` wraps `*sqlx.DB` and supports both Postgres and SQLite from the same code.
- Transactions: `DB.WithTx(ctx, fn)` stashes the `*sqlx.Tx` in context; stores call `db.GetExecutor(ctx)` to transparently use the tx or the pool.
- `qualifyTable` prefixes a schema (e.g. `cashflow.transactions`) **only on Postgres**; SQLite is schemaless and gets the bare table name. Schema/table names are constants in `db.go`.
- Follow existing rebind / duplicate-mapping / `sql.ErrNoRows` handling.

**HTTP transport** lives in `internal/transport/http`. New/refactored handlers:
- Return a plain `http.HandlerFunc`, define **feature-local request/response DTOs**, and use the codec helpers in `internal/transport/http` (imported as `httpx`): `JSONDecode`, `JSONEncode`, `DecodeQuery`, `DecodeMultipartFile`, `WriteDecodeError`.
- Carry Swagger annotations on the handler constructor.
- Keep handlers thin: decode, validate transport input, map to feature inputs, map known feature errors to HTTP status, encode. No business rules or orchestration in handlers.
- Do **not** use the legacy `internal/http.Endpoint` wrapper or import the (removed) `internal/http`.
- `Router` is a thin slice over `http.ServeMux`; routes use Go 1.22 method+pattern syntax (`"POST /accounts"`) and are registered with `HandleWithMiddleware`. `Server` wraps the mux with Elastic APM and request-identifier middleware and does graceful shutdown.

**Importer subsystem** (`internal/importer`): explicit import types `cashflow` / `portfolio` / `eod`, each with a `parsers/` package (vendor-specific: `degiro`, `ing`, `n26`, `brandnewday`) and a `processor.go` implementing the `Processor` interface (`Process(ctx, *Import) (ProcessResult, error)`). Parser factories resolve a parser by vendor/source. `Import` is the durable record with a status lifecycle (`pending → processing → completed/failed/...`).

**Market data** (`internal/marketdata`): providers (MarketStack, Alpha Vantage), listings, end-of-day "dailies", a client, and a syncer.

**Config & observability:** `internal/config` reads `config.yaml` and overlays env vars; provider API keys come from env only (`MARKETSTACK_API_KEY`, `ALPHA_VANTAGE_API_KEY`) and it exports Elastic APM env vars. Elastic APM is wired through HTTP (`apmhttp`) and SQL (`apmsql`); use the `internal/logging` slog wrapper (`logging.Logger`) for all logging with appropriate levels, and never log sensitive data.

**Reuse shared packages** before writing helpers: `internal/money` (currency/price/decimal), `internal/date` (parsing/ranges/formatting), `internal/sorting` (sort fields/directions). Put feature-level errors in that package's `errors.go`.

## Frontend architecture (`web`)

SvelteKit 2 + Svelte 5, Vite, TailwindCSS 4, TypeScript, Storybook 10, Vitest + Playwright, ESLint + Prettier.

Follow **Atomic Design strictly**: components live under `src/lib/components/{atoms,molecules,organisms,templates,pages}`. Each component folder co-locates `Component.svelte`, `Component.stories.svelte`, `component.types.ts`, and `component.variants.ts` (Tailwind variant maps). Preserve and extend this structure; don't bypass it. Handle errors explicitly and surface user-relevant failures in the UI (Elastic APM error reporting may be extended for server-side observability).

## Database & migrations

Two databases are supported (`sqlite3` for local dev, `postgres` for deployed). Migrations are **mirrored**: every change needs matching files under both `apps/api/migrations/postgres/` and `apps/api/migrations/sqlite/` (goose, timestamped). In dev the app auto-creates the DB and runs migrations on startup. Do **not** modify existing migrations or add/change deps, Docker, or infra files without explicit approval — add new timestamped migrations instead.

## Change & documentation discipline

From the AGENTS.md files — these are enforced expectations, not suggestions:
- Prefer minimal, focused diffs. Don't refactor, rename, or "modernize" unrelated code. Edit existing files over adding new abstractions unless there's clear benefit.
- Feature docs live in `docs/features/` (note: newer refactor-era docs are also appearing under `apps/api/docs/features/` with bracketed `[NNN]_NAME.md` names). Read the relevant feature doc before changing a documented feature, and update it in the same task when behavior changes.
- When a change maps to a feature, also update the root `CHANGELOG.md` under `Unreleased` using Keep a Changelog categories (`Added`/`Changed`/`Fixed`) and reference the feature ID. New feature IDs come from the changelog sequence.
- For API-contract changes, keep Swagger annotations in sync and regenerate with `make swagger` (skippable during the refactor — say so if you skip it).
- Document exported Go types/functions; document non-obvious unexported ones. Avoid comments that merely restate code, and never write numbered step comments.
- For complex multi-step work, the repo convention is to keep a concise `tmp/plan.md` and delete it when done.

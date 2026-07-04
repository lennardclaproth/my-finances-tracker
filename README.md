# my-finances-tracker

A personal finances tracker: a Go backend (`apps/api`) and a SvelteKit frontend (`web`). It tracks
cashflow, an investment portfolio, and other assets, with CSV/statement imports and market data.

- **`apps/api`** — Go 1.25 API (module `github.com/lennardclaproth/my-finances-tracker`). All domain
  logic lives here. HTTP transport under `transport/http`, features under `internal/`.
- **`web`** — SvelteKit 2 / Svelte 5 frontend. Runs standalone against mock fixtures, or against the
  live API by setting one env var.

The `Makefile` at the repo root is the single source of truth for common tasks — run `make help`.

## Prerequisites

- **Go 1.25+**
- **Node 20+** and npm
- **Docker** (for local Postgres via `deploy/docker/compose.dev.yaml`)
- **make**
- Dev tools (goose, swag, air, golangci-lint): `make install-tools`

## Quick start — run the whole app locally (Postgres)

From the repo root:

```bash
# 1. Start Postgres in Docker (the API auto-creates the DB and runs migrations on first start).
make db-up

# 2. Build and run the API on http://localhost:6060 (reads apps/api/config.yaml).
make run

# 3. In a second terminal: point the frontend at the API and start it on http://localhost:5199.
make web-env          # creates web/.env from web/.env.example (VITE_API_URL=http://localhost:6060)
make web-install      # first time only
make web-dev

# 4. Open http://localhost:5199 — it redirects to /cashflow. Reads load live from the API.
```

No API keys are required to run: the market-data providers (MarketStack / Alpha Vantage) are skipped
when their keys are unset; only live market-data sync needs them (set `MARKETSTACK_API_KEY` /
`ALPHA_VANTAGE_API_KEY` in the environment). On first start the backend seeds a default account, so
the account-scoped screens (portfolio, assets) have an account to load; the frontend discovers it via
`GET /accounts`.

To stop Postgres (and any other containers): `make db-down`.

### Run the frontend standalone (mocks, no backend)

The frontend runs entirely on in-memory fixtures when `VITE_API_URL` is unset — useful for UI work and
Storybook. Just `cd web && npm run dev` without creating `web/.env` (or set `VITE_USE_MOCKS=true`).

### CORS

The API allows the frontend dev origin (`http://localhost:5199`) out of the box via
`server.cors_allowed_origins` in `apps/api/config.yaml`. Change or extend that list for other origins;
`["*"]` allows any origin (dev only).

## Backend (`apps/api`)

### Configuration

`apps/api/config.yaml` is read relative to the working directory, so the server must run from
`apps/api` (`make run` and the VS Code "Launch Server" config already do this).

- **Database**: `database.type` is `postgres` (default; local via `make db-up`, or deployed) or
  `sqlite3` (zero-infra local — set a file connection string like `file:mft.db`). Either way the app
  creates the database if missing and runs migrations on startup.
- **Server**: `server.port` (default `6060`) and `server.cors_allowed_origins`.
- **APM** (Elastic) is configured but optional to run locally — the app boots even if no APM server is
  listening. Use `make db-up-all` to also start Elasticsearch / Kibana / APM.

### Database migrations

Migrations are goose files mirrored under `apps/api/migrations/postgres` and `.../sqlite`. They run
automatically on startup in development. Manual goose commands (need `DATABASE_URL`):

```bash
make migrate-up            # apply
make migrate-down          # roll back the last migration
make migrate-status
make migrate-create name=<migration_name>
```

### API docs (Swagger / OpenAPI)

Generated into `apps/api/docs` from handler annotations. Regenerate with `make swagger`. When the API
is running, the UI is served at `http://localhost:6060/swagger/index.html`.

### Health

`GET /health` returns `{"status":"ok"}`.

## Frontend (`web`)

SvelteKit 2 + Svelte 5 (runes), TypeScript, Tailwind CSS 4, Storybook. Scripts (from `web/`, or via
the Makefile):

```bash
make web-dev      # dev server on :5199   (cd web && npm run dev)
make web-build    # production build       (cd web && npm run build)
cd web && npm run check     # svelte-check (the type gate)
cd web && npm run lint      # prettier + eslint
cd web && npm run storybook # component workbench
```

Data access goes through a small service layer (`web/src/lib/services`) over a fetch client
(`web/src/lib/api`). Each service has a mock branch and a live branch; `VITE_API_URL` + `VITE_USE_MOCKS`
select which is used.

## Makefile reference

Run `make help` for the full list. Highlights:

| Target | What it does |
| --- | --- |
| `make build` / `make run` | Build / build+run the API |
| `make dev` | API with hot reload (air) |
| `make db-up` / `make db-down` | Start / stop local Postgres (Docker) |
| `make db-up-all` | Start Postgres + Elasticsearch/Kibana/APM |
| `make web-install` / `make web-dev` / `make web-build` | Frontend install / dev / build |
| `make web-env` | Create `web/.env` from the example |
| `make test` / `make test-integration` / `make test-all` | Backend tests |
| `make lint` | go fmt + vet + golangci-lint |
| `make swagger` | Regenerate OpenAPI docs |
| `make migrate-*` | goose migrations |
| `make install-tools` | Install swag, goose, air, golangci-lint |

## Testing

- Unit tests: `make test` (fast, no infrastructure).
- Integration tests: `make test-integration` (real DB; in-process SQLite, no extra infra). See
  `apps/api/docs/TESTING.md`.
- Frontend component tests run through Storybook's Vitest addon; there is no top-level `test` script.

## Agent / contributor guidance

Conventions live in [AGENTS.md](AGENTS.md) (repo-wide), [apps/api/AGENTS.md](apps/api/AGENTS.md)
(backend), and [web/AGENTS.md](web/AGENTS.md) (frontend). The `CLAUDE.md` files summarize them for
Claude Code.

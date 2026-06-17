# Testing

This backend uses the standard library `testing` package only — no assertion or
mocking frameworks. Tests fall into three tiers.

## Tiers

| Tier | Tests | Location | Infrastructure |
| --- | --- | --- | --- |
| **Unit** | One package in isolation, collaborators replaced with hand-written fakes and the in-memory event bus. | Co-located `*_test.go` next to the code. | None. |
| **Integration** | A feature/store against a **real, migrated database**. | `apps/api/test/` (`package integration`), behind the `integration` build tag. | SQLite, in-process (no Docker). Postgres is a planned addition. |
| **End-to-end** | The **real server binary**, driven over HTTP. | `apps/api/test/`, same package and tag. | The built binary + a throwaway SQLite DB. No Docker. |

**Unit tests live next to the package they exercise.** Integration and
end-to-end tests live together under `apps/api/test/`, separate from the code
under test.

## Separation: build tags

Everything under `apps/api/test/` starts with the build constraint:

```go
//go:build integration
```

- `go test ./...` (and `make test`) compiles **without** the tag, so the `test/`
  package contains no buildable files and is skipped — unit tests stay fast and
  infrastructure-free.
- `go test -tags=integration ./...` (and `make test-integration`) includes it.

## Running

```sh
make test              # unit only (fast, no infra) — the default
make test-integration  # integration + e2e (real SQLite DB, real server binary)
make test-all          # both
```

`make test-integration` runs locally with no Docker, including on Windows
(modernc.org/sqlite is pure Go). CI additionally runs with `-race`; the race
detector needs a C toolchain, so it is left out of the local Make target to keep
it dependency-free on Windows.

## Harnesses

Both live in `apps/api/test/harness_test.go`.

### In-process database (storage-level integration)

`newSQLiteDB(t)` opens a throwaway SQLite database in a temp dir using the
**same** `storage.NewDB` constructor and the **same embedded migrations**
(`migrations.GetFS`) the server uses at startup, run through a quiet goose
provider. `eachDialect(t, fn)` runs the body as a subtest per dialect:

```go
func TestThing(t *testing.T) {
    eachDialect(t, func(t *testing.T, db *storage.DB) {
        store := storage.NewSQLXAccountStore(db)
        // ... exercise the store against a real DB ...
    })
}
```

`eachDialect` runs **SQLite** today; **Postgres** is the planned second dialect
(the same store code branches on dialect via `qualifyTable`, so both should be
covered). When Postgres is added it slots into `eachDialect` with **no change**
to the tests that call it.

Note: the harness uses a temp **file** DB, not `:memory:`. A `modernc/sqlite`
`:memory:` database is per-connection, so migrations applied on one pooled
connection would be invisible to the next.

### Spawned server (end-to-end, black-box)

`startServer(t)` builds the application binary once per run, launches it against
a throwaway SQLite DB on a free port with a generated `config.yaml`, waits for
`GET /health`, and kills the process on cleanup. Tests then drive it over real
HTTP:

```go
func TestHTTPEndpoints(t *testing.T) {
    ts := startServer(t)
    status, body := ts.get(t, "/vendors")
    // ... assert on the HTTP response ...
}
```

This is true black-box testing of the real `cmd/my-finances-tracker` binary, so
it requires **no production code change** — the trade-off is that it is slower
than an in-process handler (process build + startup + migrations + bootstrap)
and asserts through HTTP rather than reaching into internal state. The APM agent
is disabled in the child via `ELASTIC_APM_ACTIVE=false`.

## Fixtures

Co-locate fixtures in a `testdata/` directory (the Go toolchain ignores
`testdata/`, and the test binary runs with its working directory set to the
package directory, so `os.ReadFile("testdata/...")` works). Captured
external-API responses (e.g. the existing market-data JSON) belong here when the
tests that consume them are written.

## External services and async flows

- **External HTTP** (market-data providers) must never be hit live. Serve a
  captured fixture from `httptest.NewServer` and point the provider's base URI at
  it.
- **Event bus** dispatch is asynchronous. Synchronize on outcomes (poll an
  endpoint, or subscribe to the terminal event and wait) rather than sleeping.

## Postgres (planned)

Postgres coverage was deferred to keep this setup dependency- and Docker-free.
Adding it means: a `newPostgresDB(t)` helper that provisions Postgres (a service
container in CI, or testcontainers locally) and migrates it, registered as the
second branch in `eachDialect` and skipping cleanly when Postgres is
unavailable. For end-to-end coverage, the spawned server's generated config
would point at that Postgres instead of SQLite. testcontainers would be a new
dependency and requires approval.

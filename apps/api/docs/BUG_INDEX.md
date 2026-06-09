# Bug-Fix Index

A catalogue of code-level defects surfaced while authoring the feature docs
(`docs/features/`). This is **not** the place for the expected mid-refactor wiring gaps —
those are listed separately at the bottom and cross-referenced from each feature doc's
"Refactor state" section.

- **High** items were verified against the current code (2026-06-06) — line references are accurate as of then.
- **Medium / Low** items were identified during an automated audit pass; confirm the exact line before fixing, since the tree is mid-refactor.
- No `CHANGELOG.md` entry is implied by this index; add one per fix as the AGENTS.md rules require.

## Summary

| ID | Area | Sev | Defect |
| --- | --- | --- | --- |
| BUG-001 | Assets | High | `ADJUST`+`INCREASE` records the wrong worth (delta stored as new worth) |
| BUG-002 | Assets | High | Adjust-worth direction validation is a tautology → endpoint always 400 |
| BUG-003 | Assets | High | Date layout `"2026-01-02"` instead of `"2006-01-02"` (parse + format) |
| BUG-004 | Assets | High | `CreateAsset` dereferences a possibly-nil `note` → panic |
| BUG-005 | Assets | High | `GetClassDetails` missing `return` on error → nil deref panic |
| BUG-006 | Assets | High | `ClassDetails` indexes `Mutations[0]`/`[len-1]` → panic on empty mutation list |
| BUG-007 | Market data | High | `SyncEOD` writes a nil result pointer → panic; wrong per-row error check |
| BUG-008 | Realtime | High | WebSocket has no auth and accept-all origin |
| BUG-009 | Account | Med | Create 409 branch missing `return` → double write (409 then 500) |
| BUG-010 | Account | Med | 409 message says "external_id" but uniqueness is on `name` |
| BUG-011 | Account | Med | `Commands.Create` drops the `NewAccount` error (no `%w`) |
| BUG-012 | Assets | Med | `DeleteClass`/`CreateClass` handlers don't `return` after error |
| BUG-013 | Assets | Med | Write handlers collapse domain errors to 500 (no 400/404/409) |
| BUG-014 | Assets | Med | Sync lock is a no-op → concurrent portfolio syncs unguarded |
| BUG-015 | Market data | Med | EOD responses lack JSON tags → inconsistent field casing |
| BUG-016 | Event bus | Med | No per-handler timeout → goroutine exhaustion on a hung handler |
| BUG-017 | Cashflow | Med | `GetTransactions` Swagger `@Success` names the wrong response type |
| BUG-018 | Assets | Med | `Class.Update` assigns the untrimmed name though it validates trimmed |
| BUG-019 | Event bus | Low | `import.failed` is published but has no subscriber |
| BUG-020 | Cashflow | Low | `auto_tagger.go` is orphaned dead code (no callers) |
| BUG-021 | Cashflow | Low | `NewTransactionData` accepts but ignores its `rowNumber` param |
| BUG-022 | Multiple | Low | Unused declared types/errors |
| BUG-023 | Portfolio | Low | `ErrBuildInProgress` message says "listing" (it's the account lock) |
| BUG-024 | Assets/Market data | Low | `ErrReleaseSyncLock` mis-worded + duplicated declaration |
| BUG-025 | Market data | Low | `listings.provider` FK column never set on insert |
| BUG-026 | Market data | Low | `provider_test.go` is fully commented out |
| BUG-027 | Market data | Low | Naming drift (EOD type vs dailies/TableHistories) — RESOLVED by the eods/TableEOD rename |
| BUG-028 | Market data | Low | `should_accumulate` DB default differs from `NewListing` |
| BUG-029 | Platform | Low | `migrator.go` passes printf format strings as the slog message |
| BUG-030 | Platform | Low | `HealthHandler` ignores its `r`/`log` params |
| BUG-031 | Frontend | Low | `Makefile` `WEB_DIR=./apps/web` (real dir is `web/`) → `make web-lint` fails |
| BUG-032 | Frontend | Low | Theme split: app loads `routes/layout.css`; real theme only in Storybook |
| BUG-033 | Frontend | Low | `taupe` color used (`bg-taupe-100`) but undefined in any `@theme` |
| BUG-034 | Frontend | Low | Prettier set to tabs but files use 2-space → `lint` not green |

---

## High

### BUG-001 — Assets: `ADJUST`+`INCREASE` records the wrong worth
`internal/assets/mutation.go:61-67` (`NewMutation`). For `ADJUST`+`DECREASE`,
`newWorth = previousWorth - amount` (correct). For `INCREASE`, the `else` branch sets
`amount = previousWorth + amount` but leaves `newWorth = amount` (the original delta). The
result is swapped: `NewWorth` holds the delta and `Amount` holds the new total.
**Fix:** for increase, set `newWorth = previousWorth + amount` and keep `Amount` as the delta.

### BUG-002 — Assets: adjust-worth direction validation is a tautology
`transport/http/handlers/assets/asset_worth.go:120`:
`if r.Direction != INCREASE || r.Direction != DECREASE` is always true (any value differs from
at least one), so `problems["direction"]` is always set and `PUT /assets/{asset_id}/adjust`
always returns 400. **Fix:** use `&&` (reject only when it is neither).

### BUG-003 — Assets: wrong Go date layout
`transport/http/handlers/assets/asset_worth.go:175` and
`transport/http/handlers/assets/snapshots_get.go:49,60,78` use `time.Parse`/`Format` with the
literal `"2026-01-02"` instead of Go's reference layout `"2006-01-02"`. Parsing real dates
fails (snapshots `from`/`to` → 400) and the adjust handler swallows the parse error
(`effectiveDate, _ := ...`) so it silently stores a zero date. **Fix:** use `"2006-01-02"`
everywhere (note `isValid` at `asset_worth.go:129` already uses the correct layout).

### BUG-004 — Assets: `CreateAsset` nil-note dereference
`internal/assets/commands.go:96`: `NewAsset(..., *note)` dereferences `note *string`, which is
optional and nil when the client omits it → panic. **Fix:** guard nil (pass `""` when nil) or
make `NewAsset` take `*string`.

### BUG-005 — Assets: `GetClassDetails` missing return → nil deref
`transport/http/handlers/assets/class_get.go:143-149`: on a `ClassDetails` error it writes 500
but does not `return`, then calls `toClassDetailsResponse(cd)` with `cd == nil`, which
dereferences `details.Assets` → panic (and a double JSON write). **Fix:** `return` after the
500.

### BUG-006 — Assets: `ClassDetails` panics on a class with no mutations
`internal/assets/queries.go:134-135,158`: indexes `class.Mutations[0]` and
`class.Mutations[len-1]` (and `class.Mutations[0].ClassTotalWorth`) with no length check →
index-out-of-range panic for a class that has zero mutations. **Fix:** guard empty mutations
and return zeroed bounds/worth.

### BUG-007 — Market data: `SyncEOD` nil-result panic
`internal/marketdata/syncer.go:44-111`: the named return `result *SyncEODResult` is never
initialized, yet the fetch loop writes `result.FetchErrors`/`result.FetchedCount++`/… → nil
deref panic as soon as `GetEOD` yields a row. Separately, line 96 checks the outer `err`
instead of the loop's `fetchErr`. **Fix:** initialize `result := &SyncEODResult{...}` and test
`fetchErr`. (This is the lazy auto-sync path behind `GET /marketdata/eods` for API listings.)

### BUG-008 — Realtime: WebSocket has no authn/authz
`internal/notify/hub.go` (`ServeWS`): the `/ws/accounts/{account_id}` route has no auth and no
ownership check — any caller who knows an account UUID can subscribe to that account's updates;
the gorilla `CheckOrigin` defaults to accept-all. **Fix:** add auth + account-ownership
verification and restrict origins before the upgrade.

## Medium

### BUG-009 — Account: create 409 missing return
`transport/http/handlers/account/create.go:67-74`: on `ErrAccountAlreadyExists` it writes 409
but does not `return`, then also writes 500 and logs. **Fix:** `return` after the 409.

### BUG-010 — Account: misleading 409 message
`create.go:69` says "account with the same external_id already exists", but the unique
constraint is on `name` (not `external_id`). **Fix:** correct the message (and decide whether
`external_id` should be unique).

### BUG-011 — Account: `Commands.Create` drops the domain error
`internal/account/commands.go:34` replaces the `NewAccount` error with a generic
`fmt.Errorf(...)` (no `%w`), so callers can't detect `ErrAccountNameRequired`. **Fix:** wrap
with `%w`.

### BUG-012 — Assets: class handlers fall through after error
`transport/http/handlers/assets/class_delete.go` and `class_create.go` write an error response
without `return` in some branches, falling through to a 204/continue. **Fix:** `return` after
each error write.

### BUG-013 — Assets: write handlers collapse domain errors to 500
`CreateAsset`, `CreateClass`, `UpdateClass` map `ErrAccountNotFound`/`ErrClassNotFound`/
`ErrClassAlreadyExists` to a blanket 500 despite documenting 400/404/409. **Fix:** map known
errors to their documented statuses (mirror the `SetAssetWorth`/`AdjustAssetWorth` pattern).

### BUG-014 — Assets: sync lock is a no-op
`internal/storage/sqlx_assets_store.go` `TryAcquireSyncLock`/`ReleaseSyncLock` always return
success (there is no per-account lock column), so concurrent portfolio syncs are unguarded and
`ErrSyncInProgress` is unreachable. **Fix:** add a lock column (migration) or another guard.

### BUG-015 — Market data: EOD responses lack JSON tags
`internal/marketdata` `EODResult`/`Metadata` and the handler `GetEODResponse`/
`GetEODMetadataResponse` have no `json:` tags, so they serialize with Go field names
(`Data`, `Message`, `ResultCount`, …) — inconsistent with the snake_case listing DTOs. **Fix:**
add `json:` tags.

### BUG-016 — Event bus: no per-handler timeout
`internal/eventbus/memory/memory_bus.go:240` has `// TODO: implement a timeout to prevent
goroutine exhaustion`. A hung handler ties up a worker indefinitely. **Fix:** add a per-handler
timeout/context deadline.

### BUG-017 — Cashflow: Swagger response-type mismatch
`GetTransactions` annotates `@Success ... TransactionsResponse` but returns
`GetTransactionsResponse`, so the generated spec is wrong. **Fix:** correct the annotation.

### BUG-018 — Assets: `Class.Update` stores untrimmed name
`internal/assets/class.go:81` validates the trimmed name but assigns the untrimmed `*name`.
**Fix:** assign the trimmed value.

## Low / cosmetic / dead code

- **BUG-019** — Event bus: `import.failed` (`importer.TopicFailed`) is published but nothing
  subscribes; failures are silent beyond logs/APM.
- **BUG-020** — Cashflow: `internal/cashflow/auto_tagger.go` (`AutoTagProcessor`/`AutoTagRunner`/…)
  has no callers (orphaned with the removed jobs subsystem). Remove or wire.
- **BUG-021** — Cashflow: `NewTransactionData(..., rowNumber *int, ...)` ignores `rowNumber`;
  `RowNumber` stays 0.
- **BUG-022** — Unused declarations: cashflow `TagByFilterCommand`, `GetTransactionResponse`;
  exported errors never returned (`ErrManualCashflowInvalidAmount`, `ErrManualCashflowInvalidType`,
  `ErrDuplicateTransaction`, `ErrUnsupportedDirection`, `ErrNoTransactionFound`); assets
  `ErrAssetNotInClass`.
- **BUG-023** — Portfolio: `ErrBuildInProgress` reads "sync is already in progress for this
  listing" but guards the account rebuild lock.
- **BUG-024** — `ErrReleaseSyncLock` = "Failed to release lock for the listing" is capitalized
  and listing-specific, and is declared in both `internal/marketdata/errors.go` and
  `internal/assets/errors.go`.
- **BUG-025** — Market data: `listings.provider` FK column is never set on insert; provider
  resolution happens by name at runtime, so the FK link is always null.
- **BUG-026** — Market data: `internal/marketdata/provider_test.go` is entirely commented out.
- **BUG-027** — Market data: naming drift, domain `EOD` vs table/const `dailies`/`TableHistories`.
  **Resolved** by the EOD sweep — table → `eods`, const → `TableEOD`, errors → `ErrEOD*`, parser types
  → `EODParser`/`EODRow`, and the `NewDaily`/`yieldDailies` comments fixed.
- **BUG-028** — Market data: `listings.should_accumulate` DB default is `FALSE` but
  `NewListing` sets it `true`, so app-created rows differ from the DB default.
- **BUG-029** — Platform: `migrations/migrator.go` passes printf format strings
  (`"Database %q already exists\n"`, `"Applied %d migration(s)\n"`) as the slog message with the
  arg as a field.
- **BUG-030** — Platform: `transport/http/handlers/health.go` `HealthHandler` ignores its `r`
  and `log` parameters.
- **BUG-031** — Frontend: root `Makefile` sets `WEB_DIR := ./apps/web` but the directory is
  `web/`, so `make web-lint` fails.
- **BUG-032** — Frontend: the running app loads `web/src/routes/layout.css` (no `@theme`) while
  the real theme lives in `web/src/app.css`, imported only by Storybook → app and Storybook
  render differently.
- **BUG-033** — Frontend: `app.css` uses `bg-taupe-100`/taupe utilities, but no
  `--color-taupe-*` is declared in any `@theme`.
- **BUG-034** — Frontend: Prettier is configured for tabs while most components use 2-space
  indent, so `npm run lint` is not green out of the box.

---

## Refactor-state gaps (expected — not regressions)

These are consequences of the in-progress backend refactor, already noted in the relevant
feature docs' "Refactor state" sections. They are tracked here only so they aren't
re-discovered as bugs.

- **No compiling composition root.** The new `cmd/my-finances-tracker/main.go` is an empty
  stub; `cmd/server/main.go` is stale (imports removed `internal/bus`, `internal/jobs`,
  `internal/http`, `internal/cashflow/service`, and reads a non-existent `cfg.Agent.*`).
  Nothing wires the new HTTP routes or bus subscriptions yet.
- **`eventbus.Subscribe[T]` has zero call sites** — handlers exist but are not subscribed ([017]).
- **Market-data runtime wiring is partial** — `Commands` can now be composed for provider
  bootstrap, but `Queries`/`Syncer`, the full EOD syncer/fetcher path, and market-data routes
  remain unwired ([005]).
- **`GET /portfolio/transactions` has no store implementation** — the `FetchForAccount`
  interface is unsatisfied, so the endpoint can't be served ([013]).
- **Cashflow async bulk tagging is unimplemented** (`TagByFilter` TODO; dead 202 branch); the
  `[022]` auto-tag agent was removed with the jobs subsystem ([009]).
- **`cashflow.accounts` projection is never populated** in the new tree (no `account.created`
  handler / store), so once writes are wired, cashflow inserts could violate the
  `account_id` FK ([003]/[009]). ⚠ Worth wiring before the cashflow write path is enabled.
- **`[008]` EOD-upload records + status API are unbuilt** — the `eod_uploads` table is
  orphaned schema; manual EOD currently flows through the importer (`/imports/eod`) ([005]).
- **The web app is not yet API/WebSocket-connected** — it is a Storybook component library ([020]).

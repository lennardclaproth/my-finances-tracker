# AGENTS.md

## Scope
- These instructions apply only to work inside `apps/api`.
- The repository root `AGENTS.md` still applies unless this file gives a more specific API rule.
- Do not apply frontend-specific guidance while working in `apps/api`.

## Current refactor state
- The API is currently being refactored and restructured.
- Do not assume the refactor is complete.
- When existing code conflicts, follow the explicit rules in this file first, then nearby code in the same refactored area.
- If intent is unclear or multiple architectural directions are reasonable, surface the uncertainty and trade-offs before making significant structural changes.
- Do not invent architecture, conventions, abstractions, or requirements that are not supported by existing code, documentation, or explicit instruction.

## Architecture and package boundaries
- Prefer feature-oriented packages.
- Existing feature/domain packages include `account`, `cashflow`, `portfolio`, `assets`, `marketdata`, `vendor`, `importer`, and `eodimport`.
- Supporting packages include `storage`, `jobs`, `bootstrap`, `logging`, `observability`, and `eventbus`.
- Use `Commands` and `Queries` as the default application/domain boundary when they fit the implementation.
- Other feature-level collaborators such as builders, syncers, processors, or services may be appropriate when that is the existing nearby pattern.
- Before adding a helper, utility, abstraction, validator, mapper, error type, logger wrapper, package, or reusable type, search `apps/api` for an established equivalent or nearby pattern.
- Do not add adapter interfaces, reflection bridges, or duplicate domain types just to avoid updating the real boundary. Prefer feature-owned query/command types, have storage implement those feature interfaces directly, and reuse shared packages such as `internal/sorting` for sort fields and directions.
- Put feature-level errors in the package's `errors.go` file when one exists.

## HTTP transport
- New or refactored HTTP code should follow the standards in `internal/transport/http`.
- Use the newer handler organization shown by packages such as:
  - `internal/transport/http/handlers/account`
  - `internal/transport/http/handlers/assets`
- Refactored handlers in `internal/transport/http/handlers/*` should use local request/response DTOs, return a direct `http.HandlerFunc`, and use codec helpers from `internal/transport/http`.
- Do not use the legacy `internal/http.Endpoint` wrapper or import `internal/http` from refactored `internal/transport/http` handlers.
- During the current refactor, moving related legacy handler code into `internal/transport/http` is allowed when it directly supports the requested change.
- Keep handlers thin.
- Handlers may decode requests, validate inputs, map transport models to feature inputs, map known feature errors to HTTP responses, and encode responses.
- Handlers must not contain business rules, orchestration policy, or domain decision-making that belongs in the feature package.

## DTOs and validation
- Prefer feature-local HTTP DTOs for new refactored handlers, following the style used under `internal/transport/http/handlers/assets`.
- Use shared `api` package DTOs only outside the refactored transport tree, or when a shared API contract explicitly must remain shared.
- Keep request validation close to the transport layer when it is transport-specific.
- Keep domain invariants and business validation inside the relevant feature package.

## Event bus and messaging
- The canonical event bus package is `internal/eventbus`.
- Do not introduce or depend on a separate `internal/bus` package unless explicitly instructed.
- Keep event payload contracts explicit and documented when they are exported.
- Event handlers should remain focused on reacting to events and delegating work to feature-level collaborators.

## Storage and migrations
- SQL persistence belongs in `internal/storage`.
- Follow existing SQLX store patterns for context propagation, query rebinding, transaction handling, duplicate mapping, and `sql.ErrNoRows` handling.
- Do not modify migrations unless explicitly instructed.
- When migrations are explicitly requested, create matching PostgreSQL and SQLite migrations.
- Prefer adding new timestamped migrations instead of editing existing migrations unless the user specifies otherwise.

## Change discipline
- Write the minimum code necessary for the requested task.
- Prefer small, focused diffs.
- Do not refactor unrelated code.
- Do not rename, reorganize, clean up, or modernize nearby code unless it is directly necessary for the task.
- Clean up only issues introduced by the current change unless broader cleanup is explicitly requested.
- During the current refactor, agents are not required to fix unrelated compile-breaking or consistency issues discovered elsewhere in the refactor.
- If unrelated refactor breakage blocks the requested change, explain the blocker and ask how to proceed.

## Logging and observability
- Do not make broad logging or observability changes unless explicitly requested.
- When touching code that already logs, preserve the existing logging style and avoid logging sensitive data.
- New logging or APM instrumentation is not required during the current refactor unless the task specifically asks for it.

## Tests and validation
- Use the root `Makefile` as the source of truth for commands.
- Prefer Make targets when running project commands.
- During the current refactor, full `make test`, `make lint`, `make build`, and `make swagger` are not mandatory unless explicitly requested or directly relevant to the task.
- Testing may be skipped for now when making API refactor changes.
- If validation is skipped, say so clearly in the final response.
- Do not claim a command was run if it was not run.

## Swagger and generated files
- Swagger/OpenAPI annotations should stay aligned when endpoint behavior or contracts are intentionally updated.
- During the current refactor, running `make swagger` and committing generated Swagger output can be ignored unless explicitly requested.
- Do not manually edit generated Swagger files in `apps/api/docs`; regenerate them when Swagger output is intentionally updated.

## Feature documentation and changelog
- Feature documentation lives in `/docs/features`.
- Read relevant feature documentation when the user identifies an overarching feature or when the task clearly changes a documented feature.
- Do not create new feature documentation unless explicitly asked.
- Update feature documentation only when the overarching feature significantly changes or the user requests it.
- Do not update `CHANGELOG.md` for small refactor-only changes unless explicitly requested or tied to a significant feature change.

## Comments and documentation
- Document exported Go types and functions when adding or changing them.
- Document important unexported types or functions when their purpose, invariants, or behavior are not obvious.
- Avoid comments that merely narrate syntax or restate the code.
- Keep comments aligned with actual behavior.

## Completion
- Keep final responses concise.
- Mention changed files.
- Mention validations run, or explicitly state when validation was skipped.

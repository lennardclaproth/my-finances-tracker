<!-- # AGENTS.md

## Scope
- This repository is a monorepo.
- Main applications live in `apps/api` (Go backend) and `web` (SvelteKit frontend — the directory is `web/`, not `apps/web`).
- Apply these instructions for all work unless a more specific `AGENTS.md` exists deeper in the tree.

## Commands
- Use the `Makefile` as the source of truth for project commands.
- Prefer `make build`, `make lint`, `make test`, and `make swagger`.
- Do not invent alternative commands when a Make target exists.
- Frontend scripts may exist, but default to the Makefile unless a task explicitly requires otherwise.

## Execution and validation
- Always run the full test suite with `make test` after code changes.
- When finalizing relevant changes, also run `make lint` and `make build`.
- Always run `make swagger` when finalizing API-related changes.
- Do not claim work is complete if relevant validation has not been run.

## Planning
- For complex or non-trivial tasks, create `tmp/plan.md` before making substantial changes.
- Use `tmp/plan.md` for cross-cutting work, multi-file feature work, refactors, or tasks requiring several ordered steps.
- Keep the plan concise and task-specific.
- Update the plan as work progresses.
- Delete `tmp/plan.md` when the task is complete.

## Change discipline
- Prefer minimal diffs.
- Do not refactor unrelated code.
- Prefer editing existing files over introducing new files or abstractions unless there is a clear benefit.
- Create new files when necessary to avoid bloated files or to preserve maintainability.

## Dependencies and restricted changes
- Do not add, remove, upgrade, or replace dependencies without explicit approval.
- Do not modify database migrations, CI/CD files, Docker files, infrastructure files, or similar operational files unless explicitly instructed.
- Do not delete documentation files unless explicitly told which files may be deleted.

## Backend architecture (`apps/api`)
- Prefer package-by-feature organization.
- Existing feature-oriented packages such as account, portfolio, cashflow, and vendor should remain coherent.
- Keep controllers thin.
- Controllers may parse requests, validate inputs, and map transport models to application/domain inputs.
- Controllers must not contain business rules, orchestration logic, or domain decision-making.
- Place business behavior in the appropriate feature/domain/application code, not in controllers.

## Backend coding conventions
- Use idiomatic Go.
- Add clear comments describing intent and behavior.
- Do not write numbered step comments.
- Document exported functions and types.
- Also document important unexported functions and types when their purpose is not obvious.
- Keep comments aligned with actual behavior; update comments when behavior changes.
- Use proper error handling everywhere.
- Use structured logging through the project's `slog` wrapper.
- Use appropriate log levels: debug, info, warning, error.
- Avoid logging sensitive data.

## Frontend architecture (`web`)
- Follow the Atomic Design pattern strictly.
- Preserve and extend the established atom/molecule/organism/template/page structure.
- Do not bypass the Atomic Design structure without explicit instruction.
- Keep frontend comments clear and useful, and update them when behavior changes.
- Handle errors explicitly; do not silently swallow them.
- Surface user-relevant failures appropriately in the UI.
- Where appropriate, add or maintain error reporting that can be observed server-side.
- Elastic APM integration may be extended when implementing new error reporting paths.

## Testing expectations
- Add or update tests when behavior changes.
- Do not change tests only to force a pass without understanding intended behavior.
- Prefer unit tests first, and add integration tests when they are the right fit.

## OpenAPI and Swagger
- Keep OpenAPI/Swagger documentation in sync with API changes.
- Run `make swagger` for API-related work before finalizing.

## Feature documentation
- Feature documentation is consolidated in `apps/api/docs/features/FEATURES.md` — a single short overview of all features.
- The legacy per-feature numbered files (`001_FEATURE_NAME.md`, …) under `/docs/features` have been removed; do not recreate them without explicit instruction.
- Read `FEATURES.md` before editing an existing feature.
- When a feature's behavior changes, update its entry in `FEATURES.md` in the same task.
- Also update the root `CHANGELOG.md` in the same task.
- Changelog entries go under `Unreleased` and use the Keep a Changelog categories `Added`, `Changed`, and `Fixed`.

## Feature identification rules
- The `CHANGELOG.md` still tracks features with `[NNN]` IDs (at least three digits).
- Existing feature IDs will usually be provided; for new features, infer the next ID from `CHANGELOG.md`.
- If a changed feature cannot be mapped confidently to an existing changelog feature, stop and ask.
- A feature is generally large enough to warrant a changelog entry when it introduces a new user-visible capability, meaningful backend domain behavior, or a substantial UI workflow change.
- Changelog entries for feature work must explicitly mention the feature ID.

## Maintaining this file
- If the user corrects recurring behavior or adds durable repo rules, update `AGENTS.md` so the guidance persists. -->
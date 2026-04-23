## [021] Swagger/OpenAPI Documentation and UI

### Summary
The API publishes generated Swagger 2.0 artifacts (`docs.go`, `swagger.json`, `swagger.yaml`) and serves interactive Swagger UI at runtime, with a parity test guarding route coverage.

### Why This Feature Exists
Generated OpenAPI docs provide an integration contract for frontend and external clients while reducing drift risk between registered routes and published API docs.

### Documentation Pipeline
- Handler-level swag annotations (`@Summary`, `@Param`, `@Success`, `@Router`, etc.) are the source input.
- `swag init` generates artifacts into `apps/api/docs`.
- Repository Make target `make swagger` runs generation command:
  - `swag init -g ./apps/api/cmd/server/main.go -o ./apps/api/docs`

### Runtime Serving Behavior
- Server imports generated docs package for side-effect registration.
- Route `GET /swagger/` serves Swagger UI through `http-swagger` wrapper.
- Server startup logs include a local Swagger URL hint (`/swagger/index.html`).

### Coverage and Drift Protection
- `swagger_parity_test.go` parses generated `swagger.json` and asserts expected path coverage for registered HTTP routes.
- Parity list includes major endpoint groups:
  - import/accounts/vendors
  - listings + dailies + daily uploads
  - portfolio rebuild/read/manual transaction
  - cashflow query/analytics/tag/ignore
  - health

### Contract Shape Notes
- Generated spec currently targets Swagger/OpenAPI 2.0.
- Definitions include API DTOs and selected domain response types (for example `marketdata.DailyResponse`).
- Current top-level Swagger info fields (title/version/description/host/basePath) are minimally populated in generated artifact.

### Important Notes
- Swagger artifacts are generated code (`docs.go` marked DO NOT EDIT) and should be regenerated from annotations, not edited manually.
- Parity test checks path existence, not full request/response schema correctness.
- Root-level `make fmt`, `make vet`, and `make lint` execute Go tooling within `apps/api` for monorepo compatibility.
- `/swagger/` requests also pass through request-logging middleware for consistent HTTP observability.

### Code References
- Runtime docs serving and registration:
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)
- Generated artifacts:
  - [docs.go](C:/personal/git/my-finances-tracker/apps/api/docs/docs.go)
  - [swagger.json](C:/personal/git/my-finances-tracker/apps/api/docs/swagger.json)
  - [swagger.yaml](C:/personal/git/my-finances-tracker/apps/api/docs/swagger.yaml)
- Parity guard test:
  - [swagger_parity_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/swagger_parity_test.go)
- Example annotated handlers:
  - [accounts.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/accounts.go)
  - [transactions.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/transactions.go)
  - [portfolio.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/portfolio.go)
  - [daily_uploads.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/daily_uploads.go)
  - [health.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/health.go)
- Developer commands:
  - [Makefile](C:/personal/git/my-finances-tracker/Makefile)
  - [README.md](C:/personal/git/my-finances-tracker/README.md)

### Validation Coverage
- Route-level Swagger parity is enforced by automated test against generated `swagger.json`.
  - [swagger_parity_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/swagger_parity_test.go)

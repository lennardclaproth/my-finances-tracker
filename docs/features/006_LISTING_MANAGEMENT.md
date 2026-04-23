## [006] Listing Management Endpoints

### Summary
The API provides listing management for market data entities, including create, partial update, full list, and search endpoints used by admin and selection workflows.

### Why This Feature Exists
Listings are the canonical market instruments used by portfolio transactions and daily price history. The system needs a managed registry with searchable metadata and controlled updates.

### HTTP Endpoints
1. Create listing
- `POST /marketdata/listing`
- Creates a listing record from required fields (`name`, `symbol`, `source`) plus optional metadata.
- Returns `200` with `api.ListingResponse`.
- Returns `409` when listing already exists for the same `symbol` + `source`.

2. Update listing fields
- `PATCH /marketdata/listing`
- Requires listing `id` and applies only provided optional fields.
- Returns `200` with updated `api.ListingResponse`.
- Returns `400` for invalid patch payload (for example no fields, invalid currency).
- Returns `404` when listing is not found.

3. List listings
- `GET /marketdata/listings`
- Returns all listings as `[]api.ListingResponse`.
- Sorted by `symbol ASC, id ASC`.

4. Search listings
- `GET /marketdata/listings/search?q=...&limit=...&offset=...`
- Case-insensitive partial search over `symbol`, `name`, and `isin`.
- Returns paginated `api.ListingsSearchResponse`.
- Default limit is 25, max limit is 100.

### Create Flow Behavior
Create pipeline:
1. Decode + validate request.
2. Build listing with optional typed fields (currency/type/ticker/isin/etc).
3. Persist listing with uniqueness protection (`symbol` + `source`).
4. Determine provider ingestion mode for listing source.
5. If source is API-ingestion capable, start async daily sync; if manual provider, skip async sync.

This means listing creation can trigger downstream market data accumulation automatically.

### Partial Update Semantics
Patch is field-selective:
- Only provided pointers are applied.
- Omitted fields remain unchanged.
- At least one mutable field must be provided.
- Currency is validated before persistence.

### Persistence and Query Semantics
Listing store behavior:
- `Create`: inserts record and maps uniqueness violations to `ErrListingAlreadyExists`.
- `List`: deterministic ordering for stable UI output.
- `Search`: wildcard matching with server-side paging.
- `FetchByID` and `FetchBySymbol`: used by service for updates and existence checks.

### Response Mapping
Handlers map internal listing model to API response fields:
- preserve optional fields as nullable pointers,
- convert internal currency type to API string pointer,
- return source and audit timestamps.

### Downstream Integration Impact
Listing records are consumed by:
- dailies retrieval and sync orchestration,
- daily upload validation (source/provider compatibility),
- portfolio transaction views requiring symbol/ISIN context.

### Important Notes
- `FetchBySymbol` checks symbol globally; duplicate guard enforces duplicate only when source also matches.
- Listing create currently returns `200` (not `201`).
- Search request validation requires non-empty `q`; negative offset/limit are rejected.
- Listing request/response DTO contracts used by this feature are documented in code to keep endpoint behavior explicit.

### Code References
- Listing HTTP handlers:
  - [listings.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/listings.go)
- Request validation structs:
  - [requests.go](C:/personal/git/my-finances-tracker/apps/api/api/requests.go)
- Listing service logic (create/update/list/search + async sync trigger):
  - [service.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/service.go)
- Listing persistence/search implementation:
  - [sqlx_listing_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_listing_store.go)
- Listing domain model/options:
  - [listing.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/listing.go)

### Validation Coverage
- Integration tests cover create, patch, and list endpoint behavior and async sync side-effects after create.
  - [endpoints_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/endpoints_integration_test.go)
- Service tests cover duplicate handling, patch behavior, and async sync decisions by provider mode.
  - [service_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/marketdata/service_test.go)

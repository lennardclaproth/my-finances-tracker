## [024] Asset Management Classes and Manual Worth Workflows

### Summary
Introduces an account-scoped asset management domain with manual asset classes, multiple assets inside each class, worth mutation history (`SET` and `ADJUST`), growth tracking, and a dedicated assets page with:
- top-card analytics (total assets growth line chart + class distribution donut chart)
- an asset classes table aligned with existing app table styling
- a right-side class slider for class/item worth operations
- persisted daily account snapshots (`/assets/snapshots`) for graph-ready total-worth points owned by the backend projection

### Why This Feature Exists
Users need to track non-portfolio wealth categories such as property, art, and savings, including explicit manual worth updates and adjustment deltas over time.

### Backend HTTP Contract
1. Class APIs
- `GET /assets/classes`
- `POST /assets/classes`
- `PATCH /assets/classes`
- `DELETE /assets/classes/{class_id}`
- `GET /assets/classes/{class_id}`
- `GET /assets/snapshots`

2. Item + worth mutation APIs
- `POST /assets/items`
- `POST /assets/items/worth/set`
- `POST /assets/items/worth/adjust`

### Domain Rules
1. Scope and ownership
- All operations are account-scoped (`account_id` required).
- Assets projections are created only for existing accounts.

2. Class behavior
- Manual classes can be created, renamed, archived/unarchived, and deleted.
- Class names must be non-empty; reserved portfolio class name is blocked for manual classes.

3. Item behavior
- Each class supports multiple assets/items.
- Item names are required and unique within a class.

4. Worth mutations
- `SET`: replace item worth with an absolute value.
- `ADJUST`: apply explicit `increase` or `decrease` with positive amount.
- Negative resulting worth is allowed.
- `effective_date` is required and cannot be in the future.

5. History and growth
- Every mutation writes a history row with previous/new worth and class total worth.
- Class growth points are derived from history and exposed in class-details responses.
- Class growth percentage summaries are calculated from inception (earliest history worth) to latest worth.
- Class-details growth now reads from a newest bounded history window (instead of oldest-window reads), so the latest growth point matches recent class worth even for large histories.

6. Snapshot projection and top graph behavior
- Asset mutations trigger async account snapshot rebuild requests.
- Rebuild projects one daily point per account day with carry-forward and always includes today.
- If an account has no history rows, rebuild falls back to one snapshot for today based on current account item totals (including archived items/classes).
- The web top growth chart and current-worth badge now consume `/assets/snapshots` directly, avoiding frontend class-growth aggregation drift.

### Frontend Behavior
1. Assets page
- New `/assets` page with class table columns:
  - name
  - source
  - current worth
  - last change
  - updated at
  - growth %
  - status
- Floating action button opens class creation modal and is positioned consistently with other page-level create FABs.
- Top card includes:
  - total asset growth line chart
  - class distribution donut chart
- Assets page now supports the shared navbar date-range selector (`from`/`to` route query), and applies it to:
  - top growth chart series
  - right-side class slider growth chart
  - right-side class slider history list
- Asset management subtitle text and archived-toggle control were removed for a cleaner header.
- Growth values in the class table and slider inception KPI use the same badge styling as portfolio performance badges.
- Drawer cards and table sections use the shared card/table border styling used across the app.

2. Right-side class slider
- Clicking a class row opens a right-side slider.
- Slider shows:
  - class summary KPIs
  - growth chart
  - class assets list (table styling aligned with the main asset classes table style)
  - class history list with pagination controls

3. Manual class settings workflow
- Class settings are managed from a table row edit action (cog button), not from the slider.
- Edit modal supports:
  - rename class
  - archive/unarchive class
  - delete class

4. Manual class workflows in slider
- Create new asset item in class.
- Set worth or adjust worth per asset item.

### Error Semantics
- `400` for validation/domain input errors.
- `404` for unknown account/class/item.
- `409` for duplicate class/item names.
- `500` for unexpected runtime/store errors.

### Code References
- Asset domain models/service:
  - [models.go](C:/personal/git/my-finances-tracker/apps/api/internal/assets/models.go)
  - [service.go](C:/personal/git/my-finances-tracker/apps/api/internal/assets/service.go)
- Asset SQL persistence:
  - [sqlx_asset_store.go](C:/personal/git/my-finances-tracker/apps/api/internal/storage/sqlx_asset_store.go)
- Snapshot migrations:
  - [20260311120000_assets_snapshots.sql](C:/personal/git/my-finances-tracker/apps/api/migrations/postgres/20260311120000_assets_snapshots.sql)
  - [20260311120000_assets_snapshots.sql](C:/personal/git/my-finances-tracker/apps/api/migrations/sqlite/20260311120000_assets_snapshots.sql)
- Asset HTTP handlers/routes:
  - [assets.go](C:/personal/git/my-finances-tracker/apps/api/internal/http/handlers/assets.go)
  - [main.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/main.go)
- Asset frontend flow:
  - [AssetsPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/AssetsPage.vue)
  - [AssetClassesTable.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/organisms/AssetClassesTable.vue)
  - [AssetClassDrawer.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/organisms/AssetClassDrawer.vue)
  - [EditAssetClassModal.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/EditAssetClassModal.vue)
  - [AssetGrowthLineChart.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/charts/AssetGrowthLineChart.vue)
  - [AssetDistributionDonutChart.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/charts/AssetDistributionDonutChart.vue)
  - [assets.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/assets.ts)

### Validation Coverage
- Integration tests validate manual class create/item create/worth mutation and class-details history behavior.
  - [assets_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/assets_integration_test.go)
- Integration tests also validate snapshot rebuild outcomes (`/assets/snapshots`) and latest-window class-growth behavior for >1000 history rows.
  - [assets_integration_test.go](C:/personal/git/my-finances-tracker/apps/api/cmd/server/assets_integration_test.go)
- Service tests cover core validation helpers (direction/date/worth parsing and growth-point derivation).
  - [service_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/assets/service_test.go)
- Service tests now cover dense daily snapshot projection carry-forward and no-history fallback snapshot behavior.
  - [service_test.go](C:/personal/git/my-finances-tracker/apps/api/internal/assets/service_test.go)

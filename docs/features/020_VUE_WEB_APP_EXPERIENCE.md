## [020] Vue Web App Experience

### Summary
The Vue frontend delivers account-scoped cashflow and portfolio workflows plus admin tools for listings and dailies, with URL-driven state, centralized API services, and websocket-triggered refresh behavior.

### Why This Feature Exists
The UI provides an operator-friendly workspace for importing data, tagging cashflow, managing manual portfolio entries, and administering listing/market-data operations.

### Application Shell and Routing
- App is routed via `vue-router` with `createWebHistory`.
- Primary pages:
  - `/cashflow`
  - `/portfolio`
  - `/admin/listings`
  - `/admin/dailies`
- Legacy aliases redirect to current pages (`/cashflow/transactions`, `/tagging`, `/analyze`).

### Session and Admin Mode
- Session composable keeps:
  - `activeAccountId` (bootstrapped default account ID)
  - `adminMode` persisted in local storage
- Route guard blocks `/admin/*` unless admin mode is enabled.
- Avatar popover toggles admin mode; turning it off while on admin routes redirects back to `/cashflow`.

### Core User Workflows
1. Cashflow workspace (`/cashflow`)
- Transaction table with pagination, sorting, and multi-filter query model.
- Bulk selection and tag/ignore actions (selection-based or filter-based).
- Monthly trend and tag-distribution charts.
- URL query is canonical state (deep-linkable filters/sort/page).

2. Portfolio workspace (`/portfolio`)
- Portfolio growth chart + KPI cards from snapshots.
- Positions view with include-closed toggle.
- Transaction ledger with filtering, pagination, and sorting.
- Manual transaction create modal and async rebuild trigger.

3. Admin listings (`/admin/listings`)
- Listing table with create-listing modal.

4. Admin dailies (`/admin/dailies`)
- Listing picker + date filtering + paged dailies table.
- Upload daily data modal for async ingestion acceptance.

### Shared Frontend Architecture
- Service layer wraps fetch via `requestJson`/`requestFormData`.
- API base is configurable with `VITE_API_BASE_URL` (fallback `/api`).
- API errors are normalized with `ApiError` and surfaced as user messages.
- Route query utility normalizes parse/serialize logic for cashflow and portfolio pages.

### Realtime UX Behavior
- Singleton realtime client subscribes to `/ws/accounts/{account_id}`.
- Accepted events: `portfolio.rebuilt`, `import.completed`, `bulk_tag.completed`.
- Pages perform debounced refetch on matching account events.
- Reconnect strategy:
  - exponential backoff for normal disconnects
  - for close code `4001`, reconnect waits for user interaction (`click`/`pointerdown`).

### UI System Notes
- Tailwind-based composition with shared shell/components.
- Typography uses custom primary/secondary font variables (`Noto Sans`, `EB Garamond`).
- Action and avatar popovers provide context-aware navigation and controls.

### Important Notes
- Frontend currently assumes a single bootstrap account ID for scope.
- Analyze page component exists but route currently redirects to portfolio.
- No dedicated frontend unit/integration test files are present in `apps/web` at this time.
- Frontend service contracts consumed from API DTOs are documented to keep UI/backend integration clearer.

### Code References
- App bootstrap and routing:
  - [main.ts](C:/personal/git/my-finances-tracker/apps/web/src/main.ts)
  - [index.ts](C:/personal/git/my-finances-tracker/apps/web/src/router/index.ts)
- Session/admin mode:
  - [useAppSession.ts](C:/personal/git/my-finances-tracker/apps/web/src/composables/useAppSession.ts)
  - [app.ts](C:/personal/git/my-finances-tracker/apps/web/src/config/app.ts)
- Core pages:
  - [CashflowTransactionsPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/CashflowTransactionsPage.vue)
  - [PortfolioPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/PortfolioPage.vue)
  - [AdminListingsPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/AdminListingsPage.vue)
  - [AdminDailiesPage.vue](C:/personal/git/my-finances-tracker/apps/web/src/pages/AdminDailiesPage.vue)
- Shared API/realtime services:
  - [http.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/http.ts)
  - [realtime.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/realtime.ts)
  - [cashflowTransactions.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/cashflowTransactions.ts)
  - [portfolio.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/portfolio.ts)
  - [listings.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/listings.ts)
  - [marketdata.ts](C:/personal/git/my-finances-tracker/apps/web/src/services/marketdata.ts)
- Layout/navigation and modals:
  - [AppShellTemplate.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/templates/AppShellTemplate.vue)
  - [TopNavbar.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/organisms/TopNavbar.vue)
  - [ActionMenuPopover.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/ActionMenuPopover.vue)
  - [AvatarAdminPopover.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/AvatarAdminPopover.vue)
  - [ImportDataModal.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/ImportDataModal.vue)
  - [CreatePortfolioTransactionModal.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/CreatePortfolioTransactionModal.vue)
  - [UploadDailyDataModal.vue](C:/personal/git/my-finances-tracker/apps/web/src/components/molecules/UploadDailyDataModal.vue)
- Route query normalization:
  - [routeQuery.ts](C:/personal/git/my-finances-tracker/apps/web/src/utils/routeQuery.ts)
- Styling foundation:
  - [style.css](C:/personal/git/my-finances-tracker/apps/web/src/style.css)

### Validation Coverage
- Frontend behavior is currently validated indirectly via backend endpoint/integration coverage and manual UI flows; dedicated `apps/web` automated test files are not present.

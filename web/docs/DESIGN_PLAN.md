# Finances Portal — Visual Rebuild & Design-System Plan (SvelteKit target)

> **Two codebases.** This plan ports the **visual language + interaction design** of the existing
> **Vue portal** (`apps/web`, the *visual reference* we are reproducing) into the **SvelteKit
> frontend** (`web/`, on branch **`1-refactor-backend`** — the *rebuild target*).
>
> - **Source / reference (read-only):** `apps/web` — Vue 3.5, Tailwind 4, Chart.js 4, Heroicons.
>   Defines the look, charts, drag interaction, toasts, tables, and pages to reproduce.
> - **Target (where we build):** `web/` — **SvelteKit 2 + Svelte 5 (runes forced)**, **TypeScript 6**,
>   **Vite 8**, **Tailwind 4 CSS-first** (`@theme` in `src/app.css`), **Storybook 10**
>   (`@storybook/sveltekit`, Svelte CSF), **Vitest 4 browser mode** (tests run *through stories*),
>   **`@iconify/svelte`** via the `Icon` atom, ESLint 10 + Prettier 3, `adapter-auto`, Chromatic.
>
> **The target's `web/AGENTS.md` is the binding source of truth for conventions.** Where this plan and
> `AGENTS.md` disagree, `AGENTS.md` wins. The target already has a populated design system (atoms +
> several molecules + one organism) and an established **brand that differs from the Vue original**
> (see §2 reconciliation). We adopt the Svelte system's tokens, not the Vue ones.

---

## 0. What already exists vs. what we build (the headline)

The target is **not** a greenfield. `web/src/lib/components/` already ships, fully storied:

- **Atoms:** avatar, badge, button, checkbox, currency-input, divider, file-input, icon, input, label,
  link, money, panel, progress-bar, radio, select, skeleton, slider, **sparkline**, spinner, switch,
  textarea, typography (Heading/Text).
- **Molecules:** alert, chip, dropzone, form-field, icon-button, icon-input, trend-indicator.
- **Organisms:** stat-card.
- **No** templates or pages yet; `routes/+page.svelte` is still the SvelteKit welcome page; `lib/index.ts` is empty.

So the work is **mostly gap-filling at the molecule/organism/template/page tiers** plus a small number
of foundational primitives the Vue UI leans on heavily (a popover, the charts, date pickers, a dialog).
The component-by-component gap analysis is §3.

---

## 1. Design Analysis (what we are reproducing)

The Vue portal is a **light, calm, soft-pastel data app**: heavily rounded white cards on a light
neutral background, thin gray borders, soft shadows, **EB Garamond serif headings over a 14px Noto Sans
body**, color reserved for meaning, and **charts as first-class citizens** with a bespoke
click-or-drag range-selection interaction. The full visual analysis (tokens, charts, toasts, tables,
pages, micro-interactions) is unchanged from the reference and is captured in §2 and §5.

**Two facts reshape the rebuild relative to the reference:**

1. **The Svelte design system already encodes a brand**, and it is *different* from the Vue one. The
   Vue app used **indigo/blue** for brand/selection/focus, **rose** for negative, **cyan** for info.
   The Svelte system uses **slate-600 + amber-200** as `primary`, **emerald-400** as `secondary`,
   **red** for `error`, **sky** for `info`, on a **taupe** body background. We keep the **Svelte
   tokens** and re-map the reference's accents onto them (§2.1). This is a deliberate design decision,
   flagged in §9 — do not reintroduce indigo/rose/cyan.

2. **There is no charting library in the target.** The only chart today is the hand-rolled SVG
   `sparkline` atom. The five real charts (3 interactive time-series + 2 donuts) and their drag
   interaction must be built, and the charting approach is an open decision (§9, item 1).

**Reusable structural wins already present in the target:** strict Atomic Design with a 4-file
co-location pattern, `as const` option arrays that are the single source of truth for both types and
Storybook controls, complete-literal Tailwind class maps, runes, required a11y, and Vitest browser
tests that run *through* the stories (so every story is also a test).

---

## 2. Design System Requirements (expressed in the target's vocabulary)

The target declares theme tokens in `src/app.css` via Tailwind 4 `@theme`, and components consume a
shared prop vocabulary: **`intent`** (`primary | secondary | warning | error | success | info`; badge
adds `neutral`), **`variant`** (`solid | outline | ghost`; badge `soft | solid | outline`), **`size`**
(`sm | md | lg`), **`shape`** (`default | rounded | pill`). Build everything on these — never raw hex
or one-off utilities.

### 2.1 Color — reconciliation (reference → target token)

| Role | Vue reference used | **Target token (authoritative)** |
|------|--------------------|----------------------------------|
| Brand / primary action | indigo / blue (`blue-100/300/800`) | **`intent="primary"`** → `bg-slate-600 text-amber-200 border-amber-200`, hover `slate-500`, ring `slate-300` |
| Secondary action | — | **`intent="secondary"`** → `bg-emerald-400 text-slate-800 border-slate-800` |
| Positive / income / success | emerald | **emerald** (`success` = `emerald-700`; soft `emerald-100/800`) |
| Negative / expense / danger | **rose** | **red** (`error` = `red-700`; soft `red-100/700`) |
| Warning | amber | **amber** (`warning` = `amber-500`) |
| Info | **cyan** | **sky** (`info` = `sky-700`; soft `sky-100/700`) |
| Neutral | slate | **slate** (badge `neutral`; dots `slate-500`) |
| App background | `slate-50` | **`taupe-100`** body bg *(taupe is currently undefined — see §2.9)* |
| Card surface | white | `panel` `default` = `bg-white`; `muted` = `bg-taupe-50` |
| Borders | `slate-200/300` | `slate-300` (panel border), `slate-200` (subtle) |
| Text | `slate-900/700/500/400` | body `slate-800`, headings `slate-900`, muted `slate-500` |
| Selection (rows/calendar) / focus | **indigo** | **re-map to brand** (slate/amber) or `sky` — decide in §9 |
| Modal scrim | `slate-900/50` | reuse `slate-900/50` (no token yet) |

**Chart palette — derive from `@theme`, do not hardcode the Vue hexes.** Replace the reference's
indigo/rose/cyan with the target tokens: income/positive → **emerald** (`#059669`/`emerald-600`),
expense/negative → **red** (`#b91c1c`/`#dc2626`), net/secondary line → **slate-700** (`#334155`) or
`sky`, info → **sky**, hover marker border → brand **amber/slate** (not indigo `#6366f1`). The donut
incoming palette becomes emerald+sky ramps; outgoing becomes red+amber ramps. Keep the *structural*
chart styling (dashed horizontal grid, solid darker zero line at width 1.75, dark `#0f172a` tooltip
with `#f8fafc` text, area-gradient fades top≈0.2→bottom≈0.02, donut cutout 58–64%, white slice borders,
`hoverOffset 3`). Centralize these in one `charts/theme.ts` mirroring the reference's `chartHelpers.ts`.

### 2.2 Typography
- **Body** `--font-sans` = Noto Sans (base ~14px); **headings** `--font-heading` = EB Garamond on
  `h1–h6` (`text-slate-900`). `--font-display` also Noto Sans. Type scale to reproduce: `text-[11px]`
  (labels/legend), `text-xs` (table headers, badges), `text-sm` (body, controls), `text-base`/`lg`
  (titles), KPI value via `Money size="xl"`. Use the `typography` atom (`Heading`/`Text`) for titles.
- **Webfont bug to fix (§2.9):** the woff2 files in `static/fonts/` are **not actually loaded** — see fix.

### 2.3 Spacing
Tailwind 4px scale. Reproduce: page frame `px-4 pb-4 gap-3`; `panel` padding `p-4` (`md`) / `p-3` (`sm`);
controls `px-3` with size-based heights (button `h-8/10/12`); badges `px-2/2.5/3`; FAB inset `bottom-6 right-6`.

### 2.4 Border radius
`panel.shape` = `sm`→`rounded-md`, `md`→`rounded-xl`, `xl`→`rounded-2xl`. **The reference's main cards
are `rounded-3xl`; the `panel` atom maxes at `rounded-2xl`.** Either add an `xl2`/`3xl` shape to `panel`
or accept `rounded-2xl` as the rebuild standard (recommend the latter for consistency — flag in §9).
Button/badge `shape`: `default`/`rounded`/`pill`.

### 2.5 Shadows
`panel.shadow` = `none | sm | md`. The reference uses up to `shadow-2xl` for modals/drawers; add those
levels on the dialog/drawer organisms directly (not every surface needs a token). Frosted sticky
headers/popovers: `bg-white/95 backdrop-blur`.

### 2.6 Z-index layering (define one scale — the reference's was ad-hoc)
Reproduce the reference stack, but as named steps: chart drag overlay (10) < sticky table header (20) <
popover/table-footer (30) < FAB/expanded scrim (40) < modal scrim + toast host (50) < portaled filter
popovers (60) < async search dropdown (70). The reference has inversions (modal `z-50` over expanded
scrim `z-40`; toast shares `z-50`); fix deliberately here.

### 2.7 Icons
`@iconify/svelte` **only** through the `Icon` atom, ids like `heroicons:check`,
`heroicons:chevron-down`, `heroicons:funnel`, `heroicons:arrow-up`/`-down`, `heroicons:eye`/`eye-slash`,
`heroicons:calendar-days`, `heroicons:magnifying-glass`, `heroicons:cloud-arrow-up`,
`heroicons:check-circle`, `heroicons:exclamation-triangle`, `heroicons:information-circle`,
`heroicons:x-mark`, `heroicons:x-circle`. **Boundary rule:** atoms receive icons via `children`;
icon-only/icon-decorated controls are molecules (`icon-button`, `icon-input`).

### 2.8 Animation & transition rules
Reproduce: control transitions `transition-all duration-150 ease-out` (+ `active:scale-[0.98]` per
button base); `animate-pulse` skeletons (use the `skeleton` atom); toggle thumb slide; drawer
`translate-x-full→0 duration-200` + scrim opacity; **toast** slide-in from top-right (`duration-200
ease-out`, `translate-y-2→0` / leave `duration-150 ease-in`) — use Svelte `transition:fly`/`fade`;
Actions-menu spring (`menu-panel-open 220ms cubic-bezier(0.16,1,0.3,1)` + staggered items + 360° icon
spin) — port as Svelte transitions/CSS keyframes; hover-intent popovers (160ms close delay); chart
dashed hover guide line.

### 2.9 Phase-0 fixes required in the target before building (from `web/AGENTS.md`)
1. **Theme split** — `+layout.svelte` imports only `layout.css` (`@import 'tailwindcss'` + typography);
   the real `@theme` (fonts) + `taupe` body bg live in `app.css`, imported **only by Storybook**. → App
   and Storybook render differently. **Fix:** import `app.css` from `+layout.svelte` (or move `@theme`
   into `layout.css`) so runtime == Storybook. *Without this, visual parity work in Storybook won't match the app.*
2. **`taupe` is undefined** — `bg-taupe-100/50` are used with no `--color-taupe-*` in any `@theme`.
   **Fix:** define the `taupe` ramp in `@theme` (it's the app background — first thing you see).
3. **Webfonts not loaded** — `app.css` declares `.noto-sans`/`.garamond` *classes* with a `src:`
   property (invalid outside `@font-face`) and references `/fonts/eb-garamond.woff2` while the static
   file is `eb-garamond-v32-latin-regular.woff2`. → Custom fonts likely fall back to system fonts.
   **Fix:** add proper `@font-face` rules with correct filenames. *Typography is a signature trait — verify it actually renders.*
4. **Formatting drift** (2-space vs Prettier tabs) — match each file's style; don't mass-reformat.
5. Seed `lib/index.ts` barrel and replace the welcome `+page.svelte` when routing starts.

### 2.10 Component states matrix (reproduce on every applicable component)
Hover · focus-visible ring (`focus-visible:ring-2 ring-offset-2`) · active (`active:scale-[0.98]`,
selected rows, active tab) · disabled (`disabled:opacity-50 disabled:pointer-events-none`) · loading
(`skeleton` atom / `spinner` / button `loading`) · empty (centered muted text) · error (use `alert`
`error`, or rose→**red** inline box) · filtered-active (filter trigger shows `intent="primary"`).

---

## 3. Component Gap Analysis & Atomic Breakdown (Vue source → Svelte target)

Strict tiers, 4-file co-location (`Component.svelte` + `component.types.ts` + `component.variants.ts` +
`Component.stories.svelte`), only import same/lower tier, boundary rule enforced.

### 3.1 Atoms
| Vue reference | Target atom | Status |
|---------------|-------------|--------|
| BaseButton | `button` (intent/variant/size/shape, loading/pressed) | ✅ reuse |
| BaseInput | `input` (+ `currency-input`, `textarea`) | ✅ reuse |
| BaseSelect | `select` | ✅ reuse |
| BaseCheckbox | `checkbox` | ✅ reuse |
| BaseToggle | `switch` | ✅ reuse |
| StatusBadge | `badge` (intent×variant, dot, shape) | ✅ reuse |
| VisibilityIndicator | `icon` (eye / eye-slash) | ✅ trivial via Icon |
| Money formatting | `money` | ✅ reuse |
| Card container (`rounded-3xl…`) | `panel` (variant/shape/shadow/border/padding) | ✅ reuse (radius caveat §2.4) |
| (mini chart) | `sparkline` | ✅ reuse |
| InputClearButton | — (provide via `icon-input` clear affordance) | ➕ small |
| — | spinner, skeleton, divider, avatar, link, label, typography, progress-bar, slider, radio, file-input | ✅ available |

### 3.2 Molecules
| Vue reference | Target molecule | Status |
|---------------|-----------------|--------|
| IconButton | `icon-button` | ✅ reuse |
| SearchQueryInput | `icon-input` (+ debounce wrapper) | ✅ reuse/extend |
| UnrealizedPnLBadge | `trend-indicator` (value, format percent/currency/number, up/down) | ✅ reuse |
| ToastMessage | `alert` (info/success/warning/error, dismissible — already 4 tones) | ✅ reuse |
| Chips / legend pills | `chip` | ✅ reuse |
| ImportDataModal file drop | `dropzone` (+ `file-input`) | ✅ reuse |
| Labelled field | `form-field` | ✅ reuse |
| **BasePopover** | **`popover`** (floating, portal, outside-click, viewport-aware) | 🔴 **build — foundational, blocks many** |
| SortableHeader | `sortable-header` | 🔴 build |
| Header/Direction/Select/Visibility FilterPopover | `*-filter` molecules (popover + badge/icon-button) | 🔴 build (depend on `popover`) |
| SingleDatePopover / DateRangePopover | `date-picker` / `date-range-picker` (calendar) | 🔴 build (depend on `popover`) |
| PortfolioTabs | `tabs` (segmented) | 🔴 build |
| PageBreadcrumb | `breadcrumb` | 🔴 build |
| ActionMenuPopover / AvatarAdminPopover | `action-menu` / `account-menu` (animated) | 🔴 build (depend on `popover`) |
| ListingSearchSelect | `listing-search-select` (async autocomplete) | 🔴 build |

### 3.3 Organisms
| Vue reference | Target organism | Status |
|---------------|-----------------|--------|
| PortfolioKpiCard / PortfolioGrowthKpis | `stat-card` (+ a KPI row wrapper) | ✅ reuse/extend |
| ToastMessage host | **`toast-host`** (renders `alert` queue from a toast store) | 🔴 build |
| **5 charts** (combo, trend line, asset-growth line, 2 donuts) | `charts/*` (canvas or SVG; **drag range-selection**) | 🔴 **build — highest risk** |
| TransactionsTable / PortfolioTransactionsTable / PortfolioPositionsTable / ListingsTable / ListingDailiesTable / AssetClassesTable | `*-table` (sticky header, multi-select, sort, filters, skeleton/empty/error, footer slot) | 🔴 build |
| TransactionsFooterBar / PortfolioTransactionsFooterBar | `*-footer-bar` (pagination + bulk actions) | 🔴 build |
| TopNavbar | `top-navbar` (action menu + breadcrumb + search + date range + account) | 🔴 build |
| AssetClassDrawer | `asset-class-drawer` (slide-in) | 🔴 build |
| TagModal + 6 form modals | **`dialog`** primitive (native `<dialog>`) + `*-modal` organisms | 🔴 build |

### 3.4 Templates (none exist — create the tier)
- `AppShellTemplate` — `flex h-screen min-h-0 flex-col bg-taupe-100 text-slate-800` with a `top`
  snippet slot and a `min-h-0 flex-1 overflow-hidden` main.
- `PageContentTemplate` — the repeated `flex h-full min-h-0 flex-col gap-3 px-4 pb-4` + content
  `panel` + FAB + toast-host pattern.

### 3.5 Pages (none exist — SvelteKit `routes/`)
`/cashflow` (also `/`), `/portfolio`, `/assets`, `/admin/listings`, `/admin/dailies`. Thin route
components that assemble organisms; **URL is the state store** (see §5.8). Admin routes guarded
(port the reference's guard via a `+layout`/`hooks` check on an admin-mode store).

---

## 4. Page Inventory

Identical UX to the reference; only the framework changes. Each page = `AppShellTemplate` → `top` slot
`TopNavbar` → analytics section(s) (`shrink-0`) above a `flex-1` content `panel` with a `primary` FAB and the toast host.

| Page | Route(s) | Purpose | Analytics (top) | Primary content (flex-1) | Notable |
|------|----------|---------|------------------|---------------------------|--------|
| **Cashflow** | `/cashflow` (`/`, legacy redirects) | Ledger: filter, tag, analyze | `grid lg:grid-cols-4`: trend line (`col-span-2`) + incoming donut + outgoing donut, each `h-[22vh] min-h-44 max-h-64`, expandable | Transactions table + footer bar; tag & create modals | expand → fullscreen chart dialog |
| **Portfolio** | `/portfolio` | Holdings, transactions, performance | KPI row (`stat-card`×3) + combo chart `h-[30vh] min-h-52` + rebuild/retry | tabs → positions (include-closed) / transactions + footer | expand → fullscreen card (`h-[92vh]`); search only on transactions tab |
| **Assets** | `/assets` | Asset-class worth/growth/distribution | `grid lg:grid-cols-3`: asset-growth line (`col-span-2`) + distribution donut, each `h-52` | asset-classes table → slide-in drawer; create/edit modals | drawer from right |
| **Admin · Listings** | `/admin/listings` | CRUD listing master data | — | listings table; create modal | minimal navbar, admin-gated |
| **Admin · Dailies** | `/admin/dailies` | Inspect daily OHLCV | listing search + retry | dailies table (pagination); upload modal | empty copy switches on selection, admin-gated |

Templatize the repeats: serif page header (`Heading` + actions row), the analytics chart `panel` with
a circular expand `icon-button`, the fullscreen "expanded" dialog, the FAB, the toast host.

---

## 5. Interaction Inventory (custom work)

### 5.1 ⭐ Chart range selection (draggable) — highest-risk
Used by the combo, trend-line, and asset-growth charts. The Vue code duplicates ~250 lines across
three components; **build it once** as a **Svelte action** `use:rangeSelect={{ data, getY, toRange }}`
(or a small rune module) so the behavior is defined once and testable in isolation. Reproduce exactly:
left-button pointer capture; `pointerdown/move/up/cancel/leave`; pixel→index via the scale's
`getValueForPixel` rounded+clamped; **<6px travel ⇒ single point** (point charts emit `from===to`;
month charts expand to first/last day), **otherwise range**; live overlay = brand-tinted band + two
dashed vertical guide lines + two circular point markers; suppress the hover guide while dragging;
emit `rangeselect {from,to}` (YYYY-MM-DD UTC). **SvelteKit/SSR caveat (§9):** all canvas + pointer work
must be client-only (`onMount` / `import { browser }` guard) — components SSR by default.

### 5.2 Chart rendering theme/plugins (match structurally; recolor to target tokens)
Dashed horizontal grid; solid darker zero line (width 1.75); dark tooltip (`#0f172a`/`#f8fafc`); area
gradient fade; positive/negative split datasets with the line flipping to **red** below zero; donut
click→filter + a "+N more" popover. Centralize in `charts/theme.ts`. **Charting library is an open
decision (§9).**

### 5.3 Toast system
Build a `toast` store + `toast-host` organism that renders the existing **`alert`** molecule (already
supports info/success/warning/error + dismiss) in the fixed top-right container with Svelte
`transition:fly`. Auto-dismiss ~4500ms; stackable. Move the reference's "scheduled/queued/background →
info" heuristic into the store helper. (The reference re-implemented toasts on every page — here it's one system.)

### 5.4 Tables
Sticky `bg-white/95 backdrop-blur` header (z-20); row hover `hover:bg-slate-50`; selected row brand
tint; multi-select header checkbox with **indeterminate**; sort via `sortable-header`; per-column
filter molecules (active → `intent="primary"`); absolutely-positioned footer bar (z-30) with pagination
(`[10,25,50,100]`) + bulk actions; `skeleton`/empty/`alert`-error states.

### 5.5 Popovers, menus, drawer, dialogs, calendars
**`popover` molecule is the keystone** (outside-click, controlled/uncontrolled, optional portal,
viewport-aware positioning) — filter popovers, menus, and date pickers all depend on it. Consider
Floating UI vs hand-rolling (§9). Modals → a `dialog` primitive (prefer native `<dialog>` for focus
trap + `Esc`). Drawer → right slide-in. Calendars → single + dual-month range with min/max disabled
days; selected/range states use the **brand** tokens (not indigo).

### 5.6 Debounce & async
`icon-input`/filter text/`listing-search-select` debounce 300–320ms; the async select tracks the active
request and repositions on scroll/resize.

### 5.7 Realtime
Subscribe to backend WS events (`assets.rebuilt`, `import.completed`, `bulk_tag.completed`,
`portfolio.rebuilt`) and debounce a refresh (250ms). Reuse/port the reference's `services/realtime.ts`.

### 5.8 URL-as-state (SvelteKit idiom)
Port the reference's `routeQuery.ts` parse/serialize/`areRouteQueriesEqual`/`hasActiveFilters` helpers,
but drive them from `$page.url.searchParams` + `goto(url, { keepFocus, noScroll, replaceState })`.
Tables/footers stay **controlled/presentational**; route load functions / page components own state.

---

## 6. Storybook Plan (target conventions)

Storybook 10 + `@storybook/sveltekit` + **Svelte CSF** (`<script module>` + `defineMeta`), `tags:
['autodocs']`, `title` following the tier (`Atoms/…`, `Molecules/…`, `Organisms/…`, `Templates/…`,
`Pages/…`), **`argTypes.options` fed from the `*.types.ts` `as const` arrays** (controls never drift
from types). **Vitest browser mode runs the stories as tests — adding a story adds coverage.** The
**a11y addon** is wired (`test: 'todo'`); **Chromatic** is available for visual regression. Decorators
needed: a **chart-frame** decorator (sized `relative` wrapper) and a **router/`$page`** mock for
URL-driven components.

**Coverage rules (minimum):** every component ships a `Playground` (args + a `template` snippet) plus
focused stories for variants, sizes, and the **states** that apply (default/hover-via-pseudo/focus/
disabled/loading/empty/error). Data-driven components must have **Loading / Empty / Error** stories.
Stateful interactions (drag, toast trigger, multi-select, sort, debounce) ship a `play`-function story.

| Component(s) | Required stories |
|--------------|------------------|
| New atoms/molecules (popover, sortable-header, tabs, breadcrumb, date pickers, filters, listing-search-select, dialog) | Playground + variants + States + (filters: Inactive/**Active**) + (date: selected/in-range/min-max disabled) + (popover: portal/outside-click `play`) + (search: Debounce/Empty/Error/Loading `play`) |
| toast-host | Trigger buttons for **info/success/warning/error** (`play`), Stacked, Auto-dismiss |
| charts (each) | Data, **Loading**, **Empty**, **DragSelection** (`play` asserts `rangeselect {from,to}`), SinglePointClick (`play`); donuts: ClickToFilter (`play`), "+N more" |
| tables | Default, **Loading**, **Empty**, **Error**, Selected, Sorted, Filtered |
| footer bars | NoSelection, WithSelection, AllMatching, PaginationEdges |
| drawer / dialogs / navbar | Closed/Open(loading/data/error); navbar User vs Admin, with/without search/date/filters |
| templates/pages | composed snapshot stories for Chromatic parity |
| existing atoms/molecules touched (recolor/extend) | keep/extend current stories; re-baseline Chromatic |

---

## 7. Step-by-Step Rebuild Plan (gap-driven order)

**Phase 0 — Foundation fixes & tokens (in `web/`).** Fix the theme split (load `app.css` at runtime),
define the **`taupe`** ramp, fix **`@font-face`** loading (correct filenames), set the z-index scale,
add `charts/theme.ts`, decide the **charting library** and **popover approach** (§9), seed `lib/index.ts`.
*Gate: app and Storybook render identically with correct fonts/background.*

**Phase 1 — Recolor/verify the existing system.** Re-baseline the existing atoms/molecules against the
reconciled palette (§2.1); confirm `button`/`badge`/`panel`/`alert`/`stat-card`/`trend-indicator` map
to the reference's roles. No new components — just parity + Chromatic baselines.

**Phase 2 — Foundational primitives (unblock everything).** `popover`, `dialog`, `tabs`,
`sortable-header`, `breadcrumb`, `icon-input` debounce wrapper. Story + a11y verify each.

**Phase 3 — Filters, menus, date pickers** (depend on `popover`): the four filter molecules,
`date-picker`, `date-range-picker`, `action-menu`, `account-menu`, `listing-search-select`.

**Phase 4 — Charts** (depend on tokens + `use:rangeSelect`): build the action and unit/story-test it
standalone; then trend-line, combo, asset-growth (shared action), then the two donuts. Verify
loading/empty/drag/click in Storybook.

**Phase 5 — Toast system.** `toast` store + `toast-host`; verify all four tones + stacking from Storybook.

**Phase 6 — Organisms.** KPI row, the six tables + footer bars, `top-navbar`, `asset-class-drawer`, the
seven `*-modal`s on `dialog`. Verify loading/empty/error/selected per story.

**Phase 7 — Templates & state plumbing.** `AppShellTemplate`, `PageContentTemplate`, `services/*`
(ported), URL-state helpers, admin-mode store + route guard, realtime.

**Phase 8 — Pages** (assemble only; no page-specific styling): Cashflow → Portfolio → Assets → Admin
Listings → Admin Dailies. Wire URL state, toasts, realtime, chart-range→filter, FAB, expand dialogs.

**Phase 9 — Cross-cutting verification.** Responsive (desktop/tablet/mobile), z-index audit, a11y
(addon to `error` in CI), Chromatic visual diff of pages/drawer/modals/charts vs the Vue reference.

**Dependency summary:** tokens/fixes → recolor existing → `popover`/`dialog`/primitives → filters &
date pickers & menus → charts(+action) → toast → organisms(tables/navbar/drawer/modals) → templates →
pages. `popover` and the charting decision are the two critical-path unlocks. No page starts before its
organisms are storied and Chromatic-clean.

---

## 8. Verifiable Acceptance Criteria

**Foundation (Phase 0)**
- [ ] `+layout.svelte` loads the same theme as Storybook (import `app.css` or merge `@theme`); a token
      smoke story renders identically in app and Storybook.
- [ ] `--color-taupe-*` is defined; `bg-taupe-100` body background renders (not white fallback).
- [ ] EB Garamond + Noto Sans load via valid `@font-face` (correct filenames); headings are serif in the running app.
- [ ] A single documented z-index scale exists and is used by overlays/popovers/modals/toasts.

**Design system / recolor**
- [ ] No raw hex in components except `charts/theme.ts`; all color via `intent`/tokens.
- [ ] Reference roles map to target tokens per §2.1 (rose→**red**, cyan→**sky**, indigo brand→**slate/amber**); a grep finds no `indigo`/`rose`/`cyan` utilities.

**Stories & tests (target idiom)**
- [ ] Every reusable component has a `*.stories.svelte` with `defineMeta`, tier-correct `title`,
      `tags:['autodocs']`, and `argTypes` driven from its `*.types.ts` arrays.
- [ ] Every story passes under Vitest browser mode (stories are the test corpus); a11y addon reports no violations on documented stories.
- [ ] `button`/`icon-button` show all `intent×variant×size` + disabled/loading/pressed; `checkbox` shows **indeterminate**; `switch` toggles via `play`.

**Charts**
- [ ] Each chart matches the reference structurally (dashed grid, solid zero line, dark tooltip, gradients, donut cutout) with **target-token colors**.
- [ ] Drag range-selection behaves identically across the three time-series charts because they share `use:rangeSelect`.
- [ ] Drag is testable in isolation: a `play` story simulates pointerdown→move→up and asserts `rangeselect {from,to}`; a <6px press asserts single-point/single-month.
- [ ] Donut click emits the filter; "+N more" lists remaining slices; every chart has Loading + Empty stories.
- [ ] All canvas/pointer code is client-guarded (no SSR errors in `npm run build`).

**Toasts**
- [ ] Toasts trigger from Storybook in **info/success/warning/error**; slide in top-right / out; auto-dismiss ~4500ms; stack; dismissible. Exactly one implementation (`toast` store + `toast-host`).

**Tables & data states**
- [ ] Every table/drawer/chart implements **Loading / Empty / Error** (Error via `alert`), each storied.
- [ ] Sticky header, `hover:bg-slate-50`, brand-tint selected row, indeterminate header checkbox, footer pagination + bulk actions match the reference.

**Pages & layout**
- [ ] Pages assemble only reusable components (no page-specific visual styling beyond layout grids / `PageContentTemplate`).
- [ ] Each page matches the reference at desktop/tablet/mobile (analytics grids collapse `lg:grid-cols-{3,4}`→1; navbar `md:flex-row`→stacked; titles `text-xl`→`md:text-2xl`).
- [ ] FAB at `bottom-6 right-6`; expand opens fullscreen dialog and **Esc** closes; filter/sort/paging/date/tab state round-trips through the URL; admin routes guarded.
- [ ] Chart range-selection updates the same date filter as the date picker.

**Visual parity gate**
- [ ] Chromatic (or side-by-side screenshots) of all 5 pages + drawer + each modal + each chart shows no unintended deltas vs `apps/web`, **accounting for the intentional brand recolor**.

---

## 9. Risks, Unknowns & Areas Needing Manual Review

1. **Charting library decision (blocking, needs approval — `AGENTS.md`: no new deps without approval).**
   Options: **(a) Chart.js 4** = exact visual parity with the reference + reuse of all chart logic, but
   a new dep and canvas needs SSR guards; **(b) LayerChart / LayerCake** = Svelte-native, composable,
   new dep; **(c) hand-rolled SVG** like the existing `sparkline` = zero deps and matches house style,
   but the drag math, gradients, tooltips, and donuts are significant manual effort. *Recommend (a) for
   fidelity unless the team wants zero-dep SVG.* Decide before Phase 4.
2. **Brand reconciliation (design call).** The target's `primary` is **slate-600 + amber-200**, not the
   reference's indigo/blue; negative is **red** not rose; info is **sky** not cyan. Confirm we re-map
   all reference accents onto the target brand (this plan assumes yes). Selection/calendar "active"
   color (was indigo-600) needs a chosen brand equivalent.
3. **`popover` approach (blocking many).** Hand-roll (like the Vue `BasePopover`) vs. Floating UI
   dependency. It gates filters, menus, and date pickers — decide early.
4. **Modal primitive.** Native `<dialog>` (free focus trap + `Esc` + top-layer) vs. a custom overlay.
   Recommend native `<dialog>`; verify Tailwind styling + scrim work across target browsers.
5. **SvelteKit SSR.** Charts (canvas), pointer drag, popover positioning, and any `window`/`document`
   access must be client-only (`onMount`, `browser` guard, `{#if browser}`); otherwise `npm run build`
   (prerender/SSR) breaks. Add to the chart/popover acceptance checks.
6. **Theme split + undefined `taupe` + unloaded fonts** (Phase 0). These already exist in the target
   and will silently degrade visual parity (white bg instead of taupe, system fonts instead of serif)
   until fixed.
7. **`panel` maxes at `rounded-2xl`; reference cards are `rounded-3xl`.** Add a shape or standardize on
   `rounded-2xl`. Minor but visible.
8. **Tests-through-stories implications.** There is no `test` script; coverage == stories under Vitest
   browser mode. Interaction logic (drag, debounce, selection) must be exercised by `play` functions in
   stories, not a separate unit suite. Budget for that.
9. **Backend contract.** This plan covers visual/UX parity only; the rebuild must consume the same API
   shapes as the reference (`services/*`, `types/*`; money is integer "cents × 1e6", `MONEY_SCALE =
   1_000_000`). Confirm the `1-refactor-backend` API is frozen for the frontend to target.
10. **Single-toast vs stacked queue.** The reference shows one toast at a time; this plan introduces a
    stack — confirm that UX change is acceptable.
11. **Doc/branch location.** This plan currently lives at `apps/web/docs/` on `claude/funny-easley-217651`;
    the build target is `web/` on `1-refactor-backend`. Decide where the plan should be committed (likely
    `web/docs/` on the target branch) so it travels with the code it governs.

---

### Appendix A — Target file conventions (quick reference)
- Tier folders: `web/src/lib/components/{atoms,molecules,organisms,templates,pages}/<name>/`.
- Four files per component: `Name.svelte`, `name.types.ts` (`as const` arrays + unions), `name.variants.ts`
  (complete-literal Tailwind maps `satisfies Record<…>`), `Name.stories.svelte` (`defineMeta`, autodocs).
- Runes only (`$props`/`$derived`/`$bindable`); `class` passthrough appended last; classes via
  `[...].filter(Boolean).join(' ')`. Icons only through the `Icon` atom; atoms never import components.
- Commands (from `web/`): `npm run dev` · `npm run check` (type gate) · `npm run storybook` ·
  `npm run lint` · `npm run build`.

### Appendix B — Reference (source) file map (`apps/web`, Vue)
Atoms `components/atoms/*` · molecules `components/molecules/*` (+ `charts/*`, `chartHelpers.ts`) ·
organisms `components/organisms/*` · template `components/templates/AppShellTemplate.vue` · pages
`pages/*` · routing `router/index.ts` · URL-state `utils/routeQuery.ts` · session
`composables/useAppSession.ts` · formatting `utils/formatters.ts` · data `services/*`, `types/*`.
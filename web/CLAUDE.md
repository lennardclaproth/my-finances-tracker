# CLAUDE.md — `web/`

Guidance for Claude Code when working in the SvelteKit frontend. The detailed conventions live in
[web/AGENTS.md](AGENTS.md) — **that file is the source of truth**; read it before changing components.
The repo-wide rules in the root [CLAUDE.md](../CLAUDE.md) also apply (minimal focused diffs,
changelog + feature-doc discipline, no dependency/migration/infra changes without approval).

> The frontend is `web/`, not `apps/web`. Run frontend scripts from inside `web/`, or use the Makefile
> `web-*` targets (`make web-dev`, `make web-build`, `make web-install`, `make web-env`, `make web-lint`),
> which point at `web/` correctly.

## Stack

- **SvelteKit 2 + Svelte 5** with **runes mode forced** (`svelte.config.js`) — use `$props`,
  `$derived`, `$bindable`, not legacy reactive syntax.
- **TypeScript 6** (strict), **Vite 8**.
- **Tailwind CSS 4, CSS-first** — no `tailwind.config.js`; theme tokens are declared in CSS via
  `@theme` (`src/app.css`). `@tailwindcss/typography` enabled.
- **Storybook 10** (`@storybook/sveltekit`) is the component workbench; **Vitest 4 in browser mode
  (Playwright)** runs tests *through the stories* — there is no `test` npm script.
- **ESLint 10** (flat) + **Prettier 3** (`prettier-plugin-svelte`, `prettier-plugin-tailwindcss`).
- Icons via **`@iconify/svelte`**, wrapped by the `Icon` atom (ids like `heroicons:check`). Atoms never
  import `Icon` — see the boundary rule below.

## Architecture — Atomic Design (strict)

Components live in `src/lib/components/<tier>/` with tiers **atoms → molecules → organisms →
templates → pages**. This taxonomy is mandatory; never bypass it, and only import from the same or a
lower tier.

**The boundary rule:** an atom may use native elements + external primitives (the `Icon` atom wraps
`@iconify/svelte`) but must **never** import another `$lib/components` component. Anything composing an
internal atom is a molecule. Atoms take icons through `children` (`<Button><Icon … /> Save</Button>`);
icon-only / icon-decorated controls are molecules. Reusing an atom's `*.types.ts` / `*.variants.ts`
tokens from a molecule is fine (that's data, not composition).

Populated today: **`atoms/`** (all pure) — avatar, badge, button, checkbox, currency-input, divider,
file-input, icon, input, label, link, money, panel, progress-bar, radio, select, skeleton, slider,
sparkline, spinner, switch, textarea, typography; **`molecules/`** — alert, dropzone, form-field,
icon-button, icon-input, trend-indicator; **`organisms/`** — stat-card. Create the remaining tier
folders (templates, pages) as you add their first component.

A non-trivial component co-locates four files in its folder:

- `Component.svelte` — runes, a local `type Props`, a `class` passthrough appended last, classes built
  as `[...].filter(Boolean).join(' ')`, accessibility attributes.
- `component.types.ts` — each option axis as an `as const` array with a derived union type; this array
  is the single source of truth shared with the story controls.
- `component.variants.ts` — Tailwind class strings stored as data, keyed by variant with
  `satisfies Record<…, string>`. Every utility must be a complete literal (no `` `bg-${x}-500` ``).
- `Component.stories.svelte` — Storybook CSF via `defineMeta`, `title: 'Atoms/…'` (or `'Molecules/…'`),
  `tags: ['autodocs']`, `argTypes` fed from the `as const` arrays.

Simple atoms may inline their types/variants (see `icon`, `checkbox`). Atoms share a consistent prop
vocabulary: `intent` (semantic color), `variant` (`solid/outline/ghost`), `size`, `shape`. See
[AGENTS.md](AGENTS.md) for the full per-file conventions and an add-a-component checklist.

## Commands (from `web/`)

`npm run dev` · `npm run check` (svelte-check — main type gate) · `npm run lint`
(prettier check + eslint) · `npm run format` · `npm run storybook` · `npm run build`.

## Data access & backend connection

- Services live in `src/lib/services/*` (one per feature) over the `src/lib/api` fetch client
  (`apiGet` / `apiSend` for JSON, `apiUpload` for multipart imports). Each service has a mock branch
  (fixtures in `src/lib/data/fixtures`) and a live branch.
- `src/lib/api/config.ts`: `useMocks` is true whenever `VITE_API_URL` is empty (so the app runs
  standalone on fixtures). Set `web/.env` (`make web-env`) with `VITE_USE_MOCKS=false` and
  `VITE_API_URL=http://localhost:6060` to hit the live Go API.
- The active account id comes from `accountStore` (`src/lib/stores/account.svelte.ts`), which resolves
  it from `GET /accounts` (falling back to `DEMO_ACCOUNT_ID` / the mock account). Read
  `accountStore.activeId` in pages instead of hard-coding an id, and `await accountStore.ensureLoaded()`
  before the first account-scoped request.

## Gotchas to know before you trust the build

- **Formatting drift.** Most component files use 2-space indent while Prettier is configured for tabs,
  so `npm run lint` is not green out of the box. Match a file's existing style when editing; don't
  mass-reformat the tree as a drive-by. `npm run check` (svelte-check) is the type gate and is green.
- The theme loads at runtime: `src/routes/+layout.svelte` → `src/routes/layout.css` → `src/app.css`
  (`@theme` fonts + the defined `taupe` palette + `bg-taupe-100` body). App and Storybook match.
- Real app routing exists: `routes/+page.ts` redirects `/` → `/cashflow`, and there are working
  `cashflow`, `portfolio`, `assets`, and `admin/*` feature routes with cross-screen navigation.

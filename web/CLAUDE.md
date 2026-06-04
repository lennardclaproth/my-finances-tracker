# CLAUDE.md — `web/`

Guidance for Claude Code when working in the SvelteKit frontend. The detailed conventions live in
[web/AGENTS.md](AGENTS.md) — **that file is the source of truth**; read it before changing components.
The repo-wide rules in the root [CLAUDE.md](../CLAUDE.md) also apply (minimal focused diffs,
changelog + feature-doc discipline, no dependency/migration/infra changes without approval).

> The frontend is `web/`, not `apps/web`. Run frontend scripts from inside `web/`; `make web-lint` is
> broken because it points at the nonexistent `apps/web`.

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
sparkline, spinner, switch, textarea, typography; **`molecules/`** — alert, chip, dropzone, form-field,
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

## Gotchas to know before you trust the build

- **Theme split.** `src/routes/+layout.svelte` (the running app) loads `src/routes/layout.css`, which
  only pulls in Tailwind. The real theme — `@theme` fonts + the `taupe` body background in
  `src/app.css` — is imported **only by Storybook**. App and Storybook can look different; wire
  `app.css` into the layout if you need the tokens at runtime.
- **`taupe` color is referenced but not defined** anywhere in `@theme` (`bg-taupe-100/50`). Define it
  before relying on it.
- **Formatting drift.** Most component files use 2-space indent while Prettier is configured for tabs,
  so `npm run lint` is not green out of the box. Match a file's existing style when editing; don't
  mass-reformat the tree as a drive-by.
- `routes/+page.svelte` is still the default SvelteKit welcome page — there's no real app routing yet.

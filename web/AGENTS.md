# AGENTS.md — `web/` (frontend)

Source of truth for working in the SvelteKit frontend. The repo-wide rules in the
root [CLAUDE.md](../CLAUDE.md) (and the now-commented root `AGENTS.md`) still apply —
minimal focused diffs, the changelog/feature-doc discipline, "don't add/upgrade deps,
touch migrations, Docker, or infra without approval." This file adds the frontend-specific
conventions.

> **Path note:** the frontend lives in `web/`, **not** `apps/web`. The root `AGENTS.md` and
> the `Makefile` (`make web-lint`) refer to `apps/web`, which does not exist. Run frontend
> scripts directly from inside `web/`.

## Stack and why

| Concern | Choice | Notes |
| --- | --- | --- |
| Framework | **SvelteKit 2** + **Svelte 5** | Runes mode is **forced** for all non-`node_modules` files (`svelte.config.js`). Use runes, not legacy reactive syntax. |
| Build | **Vite 8** | `@tailwindcss/vite` + `sveltekit()` plugins. |
| Language | **TypeScript 6** | `strict`, `moduleResolution: bundler`, `checkJs`/`allowJs` on. `<script lang="ts">` everywhere. |
| Styling | **Tailwind CSS 4** (CSS-first) | No `tailwind.config.js`. Theme is declared in CSS via `@theme`. `@tailwindcss/typography` plugin enabled. |
| Icons | **`@iconify/svelte`** | Referenced by string id, e.g. `heroicons:check`. Always go through the `Icon` atom. |
| Components catalog | **Storybook 10** (`@storybook/sveltekit`) | Addons: `addon-svelte-csf`, `addon-docs`, `addon-a11y`, `addon-vitest`, Chromatic. |
| Tests | **Vitest 4** in **browser mode** (Playwright/Chromium) | Tests are driven *by the stories* via `@storybook/addon-vitest`. There is no standalone unit-test runner or `test` npm script. |
| Lint/format | **ESLint 10** (flat config) + **Prettier 3** | Prettier plugins: `prettier-plugin-svelte`, `prettier-plugin-tailwindcss` (sorts Tailwind classes). |
| Adapter | `@sveltejs/adapter-auto` | Deploy target not yet pinned. |

Scaffolded by `sv create` (minimal TS template). `.npmrc` sets `engine-strict=true`.

## Commands (run from `web/`)

- `npm run dev` — Vite dev server.
- `npm run build` / `npm run preview` — production build / preview.
- `npm run check` — `svelte-kit sync` + `svelte-check` (type/Svelte diagnostics). Use this as the
  primary "did I break types" gate.
- `npm run lint` — `prettier --check .` then `eslint .`.
- `npm run format` — `prettier --write .`.
- `npm run storybook` — Storybook dev server on port 6006 (the main way to develop/preview components).
- `npm run build-storybook` — static Storybook build.

There is **no `test` script**. Component tests execute through Storybook stories under Vitest's
browser project (`vite.config.ts` → `storybookTest`). Adding a story adds test coverage.

## Directory layout

```
web/
  src/
    app.html            # document shell (SvelteKit placeholders)
    app.css             # @theme tokens + base layer (fonts, body bg) — see theming caveat
    app.d.ts            # App namespace type stubs
    routes/             # SvelteKit file-based routing
      +layout.svelte    # root layout; imports layout.css
      +page.svelte      # still the default SvelteKit welcome page
      layout.css        # the stylesheet the running app actually loads
    lib/
      index.ts          # $lib barrel (currently empty)
      assets/           # favicon, etc.
      components/
        atoms/          # pure primitives (no internal-component imports)
          avatar/  badge/  button/  checkbox/  currency-input/  divider/  file-input/  icon/
          input/  label/  link/  money/  panel/  progress-bar/  radio/  select/  skeleton/
          slider/  sparkline/  spinner/  switch/  textarea/  typography/
        molecules/      # compositions of atoms
          alert/  dropzone/  form-field/  icon-button/  icon-input/  trend-indicator/
        organisms/      # sections composed of molecules + atoms
          stat-card/
        templates/      # (planned — not yet created)
        pages/          # (planned — not yet created)
  static/               # fonts/, robots.txt (prettier-ignored)
  .storybook/           # main.ts, preview.ts, preview-head.html
```

## Atomic Design (strict)

Components live under `src/lib/components/<tier>/` where tier is one of **atoms → molecules →
organisms → templates → pages**. This taxonomy is mandatory; **do not bypass it** and do not place
components elsewhere.

> **The boundary rule.** An atom may render native HTML elements, inline SVG, and *external*-library
> primitives (e.g. the `Icon` atom wraps `@iconify/svelte`). An atom must **never** import another
> `$lib/components` component. Any component that composes one or more internal atoms is a **molecule**
> (or higher) and lives in `molecules/`. Icons inside an atom-level control are supplied by the
> consumer via `children` (a snippet), not by the atom importing `Icon`. Reusing an atom's styling
> *tokens* (its `*.types.ts` / `*.variants.ts` maps) from a molecule is fine — that's data, not
> component composition.

- **atoms** — indivisible primitives. Form controls: input, currency-input, select, textarea, checkbox,
  radio, switch, slider, file-input, label. Display & feedback: typography, money, badge, progress-bar,
  sparkline, spinner, skeleton, divider, avatar, link, icon, panel, button.
- **molecules** — small groups of atoms wired together (e.g. a labelled field = label + input + error).
- **organisms** — self-contained sections composed of molecules/atoms.
- **templates** — page-level layout/structure without real data.
- **pages** — concrete instances (typically thin; SvelteKit `routes/` render these).

A component may only import from its own tier or a **lower** tier (atoms never import molecules).
Compose upward.

**Current state:** `atoms/`, `molecules/`, and `organisms/` are populated. `molecules/` holds
`alert`, `dropzone`, `form-field`, `icon-button`, `icon-input`, and `trend-indicator`;
`organisms/` holds `stat-card` (it composes the `trend-indicator` molecule, which is why it is an
organism rather than a molecule). Templates/pages don't exist yet; create the folder when you add the
first component of that tier and keep the same anatomy below.

## Component anatomy (co-location)

A non-trivial component is a folder co-locating four files:

```
button/
  Button.svelte          # markup + logic
  button.types.ts        # option vocabularies + derived union types
  button.variants.ts     # Tailwind class maps keyed by variant
  Button.stories.svelte  # Storybook CSF (catalog + tests)
```

Naming: folders are lowercase (kebab-case when multi-word, e.g. `icon-button/`); **PascalCase** for
`.svelte` components and `*.stories.svelte`; **lowercase** for `*.types.ts` / `*.variants.ts`. One
folder may hold several related components that share the `types`/`variants` (e.g. `Text.svelte` +
`Heading.svelte` in `typography/`).

**Exceptions are allowed for simple atoms.** `icon/` ships only `Icon.svelte` (its tiny size map is
inline). `checkbox/` has types + component + stories but inlines its classes instead of a `variants`
file. Rule of thumb: extract `types`/`variants` once a component has a real variant matrix; inline
them when it has one trivial scale. Stories are expected for anything reusable.

### `*.types.ts` — the option vocabulary

Each axis is an `as const` array of string literals; the union type is derived from it. This array is
the **single source of truth** reused by both the component and the story's `argTypes`.

```ts
export const buttonIntents = ['primary', 'secondary', 'warning', 'error', 'success', 'info'] as const;
export type ButtonIntent = (typeof buttonIntents)[number];
```

### `*.variants.ts` — Tailwind class maps

Class strings are stored as data: base classes as a joined array, variant axes as objects keyed by the
union type and guarded with `satisfies Record<…, string>`. Nested records model an `intent × variant`
matrix.

```ts
export const buttonSizeClasses = {
  sm: 'h-8 px-2 text-sm',
  md: 'h-10 px-3 text-sm',
  lg: 'h-12 px-3 text-base'
} satisfies Record<ButtonSize, string>;
```

> **Tailwind 4 rule:** every utility must appear as a complete literal in the source so the content
> scanner (and `prettier-plugin-tailwindcss`) can see it. Key whole class strings by variant — never
> assemble a single utility from fragments (no `` `bg-${color}-500` ``). Composing *several whole
> utilities* via array `.join(' ')` is fine.

### `Component.svelte` — runtime conventions

- **Svelte 5 runes only.** Props via `$props()`, derived state via `$derived()`, two-way props via
  `$bindable()`.
- **Typed props.** Declare a local `type Props = { … }` (or `interface Props extends Omit<HTML…Attributes, …>`
  when wrapping a native element, as `Checkbox` does). Default values in the destructure.
- **`class` passthrough.** Accept `class?: string`, rename it `class: className = ''`, and append it
  **last** so callers can extend a component's styling.
- **Class assembly.** Build the final string from an array → `.filter(Boolean).join(' ')`, wrapped in
  `$derived` when it depends on props. Pull the pieces from the `variants` maps.
- **Children.** `children?: Snippet` rendered with `{@render children?.()}`.
- **Events.** Expose callback props (`onclick?: (e: MouseEvent) => void`) and guard them
  (a disabled/loading control must short-circuit before calling back).
- **Accessibility is not optional.** Mirror the existing atoms: `aria-busy`/`aria-pressed` on Button,
  `aria-invalid`/`aria-describedby` on Input, the visually-hidden `peer sr-only` real `<input>` behind
  a styled box on Checkbox, and a **required** `ariaLabel` on icon-only controls (`IconButton`). The
  a11y addon is wired into Storybook — keep stories clean.
- **Polymorphic tags.** Use `<svelte:element this={…}>` for components whose element varies
  (`Heading` `level`, `Text` `as`).
- **Icons.** The `Icon` atom (`$lib/components/atoms/icon/Icon.svelte`) wraps `@iconify/svelte`; never
  import `@iconify/svelte` anywhere else. **Atoms do not render `Icon`** — an atom-level control takes
  icons through its `children` (e.g. `<Button><Icon icon="heroicons:check" /> Save</Button>`).
  Icon-only or icon-decorated controls are molecules (`molecules/icon-button`, `molecules/icon-input`)
  that compose the atom + `Icon`.

## Design-token vocabulary

Atoms share a consistent prop language — reuse these names rather than inventing new ones:

- **`intent`** — semantic color role: `primary | secondary | warning | error | success | info`
  (Badge adds `neutral`).
- **`variant`** — fill style: `solid | outline | ghost` (Badge: `soft | solid | outline`).
- **`size`** — `sm | md | lg` (typography uses its own scales, e.g. heading `sm…2xl`).
- **`shape`** — corner rounding: `default | rounded | pill`.

Palette in use: Tailwind `slate / emerald / amber / red / sky` plus a custom **`taupe`** family for
backgrounds.

## Theming (Tailwind 4 CSS-first)

- Theme tokens are declared in `src/app.css` inside `@theme` — font tokens `--font-sans` (Noto Sans),
  `--font-display`, `--font-heading` (EB Garamond) — with a base layer setting the body background
  (`bg-taupe-100`) and serif headings. Webfonts live in `static/fonts/`.
- The prettier `tailwindStylesheet` (for class sorting) points at `src/routes/layout.css`.

⚠️ **Two-stylesheet split — keep these in sync.** The running app's root layout
(`src/routes/+layout.svelte`) imports `src/routes/layout.css`, which currently only does
`@import 'tailwindcss'` + the typography plugin. The richer theme in `src/app.css` (the `@theme`
fonts and `taupe` base styles) is imported **only by Storybook** (`.storybook/preview.ts`). As a
result the app and Storybook can render differently. If you rely on theme tokens, make sure the app
actually loads them (import `app.css` from the layout, or move the `@theme` into `layout.css`).

⚠️ **`taupe` is referenced but not defined.** `bg-taupe-100` / `bg-taupe-50` are used without a
corresponding `--color-taupe-*` declaration in any `@theme` block. Verify/define it before depending
on those classes.

## Stories and tests

Stories use the Svelte CSF format:

```svelte
<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import Button from './Button.svelte';
  import { buttonIntents, buttonVariants } from './button.types';

  const { Story } = defineMeta({
    title: 'Atoms/Button',          // tier/name — keep titles in this hierarchy
    component: Button,
    tags: ['autodocs'],             // generates the docs page
    argTypes: { intent: { control: 'select', options: buttonIntents } }  // feed options from the as-const arrays
  });
</script>
```

- `title` follows the atomic tier (`Atoms/…`, then `Molecules/…`, etc.).
- Drive `argTypes` `options` from the `*.types.ts` arrays so controls and types never drift.
- Provide a `Playground` story plus focused stories (variants, sizes, states) — these double as the
  test corpus under Vitest browser mode. Run them via Storybook's test UI / the vitest project.

## Linting and formatting

- ESLint flat config (`eslint.config.js`): JS + TS-recommended + Svelte-recommended + Storybook,
  with `eslint-config-prettier` last. `.gitignore` is honored via `includeIgnoreFile`.
- Prettier (`.prettierrc`): **tabs**, single quotes, no trailing commas, print width 100, Svelte +
  Tailwind plugins. `.prettierignore` excludes lockfiles and `/static/`.

⚠️ **Known formatting drift.** Most files under `src/lib/components/atoms/` are currently authored
with **2-space** indentation, which conflicts with the `useTabs: true` Prettier setting (config files
and `routes/` use tabs; `badge/` is mixed). `prettier --check` will therefore flag these files, so
`npm run lint` is not green out of the box. Do **not** mass-reformat the component tree as a
drive-by; either match a file's existing style when editing it, or fix formatting as its own
intentional change.

## Adding a new component — checklist

1. Pick the correct tier folder under `src/lib/components/` (create it if it's the tier's first
   component). Only import from the same or a lower tier.
2. Create `Component.svelte` following the runtime conventions above (runes, typed `Props`, `class`
   passthrough, a11y). If it needs to render an icon or otherwise compose another component, it belongs
   in `molecules/` (compose the atom + `Icon`), not in `atoms/`.
3. For a real variant matrix, add `component.types.ts` (`as const` arrays + derived unions) and
   `component.variants.ts` (`satisfies Record` class maps). Reuse the shared `intent/variant/size/shape`
   vocabulary. For a trivial atom, inlining is acceptable (see `icon`/`checkbox`).
4. Add `Component.stories.svelte` with `defineMeta`, the correct `Atoms/…`-style `title`,
   `tags: ['autodocs']`, and `argTypes` fed from the option arrays. Cover the meaningful states.
5. Run `npm run check`; preview in `npm run storybook`. Match the file's existing indentation.
6. If the component implements or changes a documented feature, update the feature doc under
   `docs/features/` and the root `CHANGELOG.md` (`Unreleased`) per the root conventions.

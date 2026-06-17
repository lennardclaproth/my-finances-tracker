// $lib barrel. Re-exports the data/infra layer and design-system primitives so consumers can import
// from `$lib` rather than deep paths. Component-level deep imports (e.g. `$lib/components/atoms/...`)
// remain valid; this barrel is additive.

// Data + API layer (types, money helpers, config flag, fetch client, fixtures, services).
export * from './api';
export * from './services';
export * as fixtures from './data';

// Styling tokens.
export * from './styles/z-index';
export * as chartTheme from './charts/theme';

// Foundational primitives (Phase 2).
export { default as Popover } from './components/molecules/popover/Popover.svelte';
export type { PopoverApi, PopoverPlacement } from './components/molecules/popover/popover.types';
export { default as Dialog } from './components/molecules/dialog/Dialog.svelte';
export type { DialogSize } from './components/molecules/dialog/dialog.types';
export { default as Tabs } from './components/molecules/tabs/Tabs.svelte';
export type { TabItem, TabsSize } from './components/molecules/tabs/tabs.types';
export { default as SortableHeader } from './components/molecules/sortable-header/SortableHeader.svelte';
export type {
	SortDirection,
	SortableHeaderAlign
} from './components/molecules/sortable-header/sortable-header.types';
export { default as Breadcrumb } from './components/molecules/breadcrumb/Breadcrumb.svelte';
export type { BreadcrumbItem } from './components/molecules/breadcrumb/breadcrumb.types';
export { default as SearchInput } from './components/molecules/search-input/SearchInput.svelte';

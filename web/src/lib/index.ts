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

// Filters, menus, and date pickers (Phase 3).
export { default as Calendar } from './components/molecules/calendar/Calendar.svelte';
export type { CalendarMode } from './components/molecules/calendar/calendar.types';
export * as calendarUtils from './components/molecules/calendar/calendar.utils';
export { default as DatePicker } from './components/molecules/date-picker/DatePicker.svelte';
export { default as DateRangePicker } from './components/molecules/date-range-picker/DateRangePicker.svelte';
export { default as FilterPopover } from './components/molecules/filter-popover/FilterPopover.svelte';
export { default as TextFilter } from './components/molecules/text-filter/TextFilter.svelte';
export { default as DirectionFilter } from './components/molecules/direction-filter/DirectionFilter.svelte';
export { default as SelectFilter } from './components/molecules/select-filter/SelectFilter.svelte';
export { default as VisibilityFilter } from './components/molecules/visibility-filter/VisibilityFilter.svelte';
export { default as ActionMenu } from './components/molecules/action-menu/ActionMenu.svelte';
export type { MenuItem } from './components/molecules/action-menu/menu.types';
export { default as AccountMenu } from './components/molecules/account-menu/AccountMenu.svelte';
export { default as ListingSearchSelect } from './components/molecules/listing-search-select/ListingSearchSelect.svelte';

// Charts (Phase 4).
export { default as TimeSeriesChart } from './components/organisms/charts/TimeSeriesChart.svelte';
export { default as DonutChart } from './components/organisms/charts/DonutChart.svelte';
export type { SeriesDataset, DonutDatum } from './charts/types';

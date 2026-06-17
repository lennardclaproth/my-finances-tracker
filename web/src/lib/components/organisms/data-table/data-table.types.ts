import type { Snippet } from 'svelte';

export type SortDirection = 'asc' | 'desc';
export type ColumnAlign = 'left' | 'center' | 'right';

/** Column definition for the DataTable. `T` is the row type. */
export interface Column<T> {
	key: string;
	header: string;
	/** When set, the column header becomes sortable and emits this key via onSort. */
	sortKey?: string;
	align?: ColumnAlign;
	/** Tailwind width utility for the column (e.g. 'w-32'). */
	width?: string;
	/** Default text accessor (used when no `cell` snippet is supplied). */
	value?: (row: T) => string | number;
	/** Custom cell renderer. */
	cell?: Snippet<[T]>;
	/** Optional header filter control (e.g. a TextFilter / SelectFilter). */
	filter?: Snippet;
}

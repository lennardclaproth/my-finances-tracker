<script lang="ts">
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import type { SortableHeaderAlign, SortDirection } from './sortable-header.types';

	type Props = {
		label: string;
		/** This column's sort key. */
		sortKey: string;
		/** The currently-sorted column key (column is active when it equals `sortKey`). */
		activeKey?: string;
		/** Current sort direction (only meaningful while active). */
		direction?: SortDirection;
		align?: SortableHeaderAlign;
		onSort?: (key: string, direction: SortDirection) => void;
		class?: string;
	};

	let {
		label,
		sortKey,
		activeKey,
		direction = 'asc',
		align = 'left',
		onSort,
		class: className = ''
	}: Props = $props();

	const active = $derived(activeKey === sortKey);

	const ariaSort = $derived(
		(active ? (direction === 'asc' ? 'ascending' : 'descending') : 'none') as
			| 'ascending'
			| 'descending'
			| 'none'
	);

	const iconName = $derived(
		active
			? direction === 'asc'
				? 'heroicons:arrow-up'
				: 'heroicons:arrow-down'
			: 'heroicons:chevron-up-down'
	);

	const alignClasses = {
		left: 'justify-start text-left',
		center: 'justify-center text-center',
		right: 'justify-end text-right'
	} satisfies Record<SortableHeaderAlign, string>;

	function handleClick() {
		// Toggle direction when already active; otherwise start ascending.
		const next: SortDirection = active && direction === 'asc' ? 'desc' : 'asc';
		onSort?.(sortKey, next);
	}

	const thClasses = $derived(
		['px-3 py-2 text-xs font-medium text-slate-500', alignClasses[align], className]
			.filter(Boolean)
			.join(' ')
	);
</script>

<th scope="col" aria-sort={ariaSort} class={thClasses}>
	<button
		type="button"
		class={[
			'group inline-flex items-center gap-1 rounded-md whitespace-nowrap',
			'transition-colors hover:text-slate-800',
			'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
			active ? 'text-slate-800' : ''
		]
			.filter(Boolean)
			.join(' ')}
		aria-label={`Sort by ${label}`}
		onclick={handleClick}
	>
		<span>{label}</span>
		<Icon
			icon={iconName}
			size="sm"
			class={active ? 'text-slate-700' : 'text-slate-300 group-hover:text-slate-400'}
		/>
	</button>
</th>

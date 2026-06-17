<script lang="ts" generics="T">
	import type { Snippet } from 'svelte';
	import Checkbox from '$lib/components/atoms/checkbox/Checkbox.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Skeleton from '$lib/components/atoms/skeleton/Skeleton.svelte';
	import { zClasses } from '$lib/styles/z-index';
	import type { Column, ColumnAlign, SortDirection } from './data-table.types';

	type Props = {
		columns: Column<T>[];
		rows: T[];
		getRowId?: (row: T) => string;
		loading?: boolean;
		error?: string | null;
		emptyText?: string;
		selectable?: boolean;
		/** Bindable selected row ids. */
		selectedIds?: string[];
		sortKey?: string;
		sortDirection?: SortDirection;
		onSort?: (key: string, direction: SortDirection) => void;
		onRowClick?: (row: T) => void;
		/** Footer content (e.g. a FooterBar) rendered below the scroll area. */
		footer?: Snippet;
		class?: string;
	};

	let {
		columns,
		rows,
		getRowId = (row: T) => String((row as { id?: unknown }).id ?? ''),
		loading = false,
		error = null,
		emptyText = 'No results',
		selectable = false,
		selectedIds = $bindable([]),
		sortKey,
		sortDirection = 'asc',
		onSort,
		onRowClick,
		footer,
		class: className = ''
	}: Props = $props();

	const alignClasses = {
		left: 'text-left',
		center: 'text-center',
		right: 'text-right'
	} satisfies Record<ColumnAlign, string>;

	const allIds = $derived(rows.map(getRowId));
	const allSelected = $derived(
		selectable && rows.length > 0 && allIds.every((id) => selectedIds.includes(id))
	);
	const someSelected = $derived(
		selectable && allIds.some((id) => selectedIds.includes(id)) && !allSelected
	);
	const colSpan = $derived(columns.length + (selectable ? 1 : 0));
	const isEmpty = $derived(!loading && !error && rows.length === 0);

	function toggleAll() {
		if (allSelected) {
			selectedIds = selectedIds.filter((id) => !allIds.includes(id));
		} else {
			const merged = [...selectedIds];
			for (const id of allIds) if (!merged.includes(id)) merged.push(id);
			selectedIds = merged;
		}
	}

	function toggleRow(id: string) {
		selectedIds = selectedIds.includes(id)
			? selectedIds.filter((x) => x !== id)
			: [...selectedIds, id];
	}

	function headerSort(column: Column<T>) {
		if (!column.sortKey) return;
		const next: SortDirection =
			sortKey === column.sortKey && sortDirection === 'asc' ? 'desc' : 'asc';
		onSort?.(column.sortKey, next);
	}

	function ariaSort(column: Column<T>): 'ascending' | 'descending' | 'none' | undefined {
		if (!column.sortKey) return undefined;
		if (sortKey !== column.sortKey) return 'none';
		return sortDirection === 'asc' ? 'ascending' : 'descending';
	}

	function sortIcon(column: Column<T>): string {
		if (sortKey !== column.sortKey) return 'heroicons:chevron-up-down';
		return sortDirection === 'asc' ? 'heroicons:arrow-up' : 'heroicons:arrow-down';
	}
</script>

<div class={['flex min-h-0 flex-col', className].filter(Boolean).join(' ')}>
	<div class="min-h-0 flex-1 overflow-auto">
		<table class="w-full border-collapse text-sm">
			<thead class={['sticky top-0 bg-white/95 backdrop-blur', zClasses.stickyHeader].join(' ')}>
				<tr class="border-b border-slate-200">
					{#if selectable}
						<th class="w-10 px-3 py-2">
							<Checkbox
								checked={allSelected}
								indeterminate={someSelected}
								onchange={toggleAll}
								aria-label="Select all"
							/>
						</th>
					{/if}
					{#each columns as column (column.key)}
						<th
							scope="col"
							aria-sort={ariaSort(column)}
							class={[
								'px-3 py-2 text-xs font-medium text-slate-500',
								alignClasses[column.align ?? 'left'],
								column.width ?? ''
							]
								.filter(Boolean)
								.join(' ')}
						>
							<div
								class={[
									'flex items-center gap-1',
									column.align === 'right' ? 'justify-end' : '',
									column.align === 'center' ? 'justify-center' : ''
								]
									.filter(Boolean)
									.join(' ')}
							>
								{#if column.sortKey}
									<button
										type="button"
										class="group inline-flex items-center gap-1 rounded-md whitespace-nowrap transition-colors hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
										onclick={() => headerSort(column)}
									>
										<span>{column.header}</span>
										<Icon
											icon={sortIcon(column)}
											size="sm"
											class={sortKey === column.sortKey
												? 'text-slate-700'
												: 'text-slate-300 group-hover:text-slate-500'}
										/>
									</button>
								{:else}
									<span>{column.header}</span>
								{/if}
								{#if column.filter}
									{@render column.filter()}
								{/if}
							</div>
						</th>
					{/each}
				</tr>
			</thead>

			<tbody>
				{#if loading}
					{#each [0, 1, 2, 3, 4, 5] as rowIndex (rowIndex)}
						<tr class="border-b border-slate-100">
							{#if selectable}
								<td class="px-3 py-2"><Skeleton variant="rect" class="h-4 w-4" /></td>
							{/if}
							{#each columns as column (column.key)}
								<td class="px-3 py-2">
									<Skeleton />
								</td>
							{/each}
						</tr>
					{/each}
				{:else if error}
					<tr>
						<td colspan={colSpan} class="px-3 py-8">
							<div
								class="mx-auto max-w-md rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-center text-sm text-red-700"
							>
								{error}
							</div>
						</td>
					</tr>
				{:else if isEmpty}
					<tr>
						<td colspan={colSpan} class="px-3 py-10 text-center text-sm text-slate-500">
							{emptyText}
						</td>
					</tr>
				{:else}
					{#each rows as row (getRowId(row))}
						{@const id = getRowId(row)}
						{@const selected = selectable && selectedIds.includes(id)}
						<tr
							class={[
								'border-b border-slate-100 transition-colors',
								selected ? 'bg-amber-50' : 'hover:bg-slate-50',
								onRowClick ? 'cursor-pointer' : ''
							]
								.filter(Boolean)
								.join(' ')}
							onclick={() => onRowClick?.(row)}
						>
							{#if selectable}
								<td class="px-3 py-2">
									<Checkbox
										checked={selected}
										onchange={() => toggleRow(id)}
										aria-label="Select row"
									/>
								</td>
							{/if}
							{#each columns as column (column.key)}
								<td
									class={[
										'px-3 py-2 text-slate-700',
										alignClasses[column.align ?? 'left'],
										column.width ?? ''
									]
										.filter(Boolean)
										.join(' ')}
								>
									{#if column.cell}
										{@render column.cell(row)}
									{:else if column.value}
										{column.value(row)}
									{/if}
								</td>
							{/each}
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>

	{#if footer}
		<div class={['border-t border-slate-200', zClasses.popover].filter(Boolean).join(' ')}>
			{@render footer()}
		</div>
	{/if}
</div>

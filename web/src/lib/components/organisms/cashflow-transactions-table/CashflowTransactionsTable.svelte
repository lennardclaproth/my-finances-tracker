<script lang="ts">
	import type { Snippet } from 'svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import FooterBar from '$lib/components/organisms/footer-bar/FooterBar.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import TextFilter from '$lib/components/molecules/text-filter/TextFilter.svelte';
	import SelectFilter from '$lib/components/molecules/select-filter/SelectFilter.svelte';
	import DirectionFilter from '$lib/components/molecules/direction-filter/DirectionFilter.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { SortDirection } from '$lib/components/organisms/data-table/data-table.types';
	import type { CashflowDirection, CashflowTransaction } from '$lib/api/types';

	type Props = {
		rows: CashflowTransaction[];
		loading?: boolean;
		error?: string | null;
		total?: number;
		limit?: number;
		offset?: number;
		selectedIds?: string[];
		sortKey?: string;
		sortDirection?: SortDirection;
		descriptionFilter?: string;
		tagFilter?: string[];
		directionFilter?: CashflowDirection | null;
		tagOptions?: { value: string; label: string }[];
		onSort?: (key: string, direction: SortDirection) => void;
		onPageChange?: (offset: number) => void;
		onLimitChange?: (limit: number) => void;
		onFilterChange?: () => void;
		/** Bulk actions for the footer when rows are selected. */
		bulkActions?: Snippet;
		class?: string;
	};

	let {
		rows,
		loading = false,
		error = null,
		total = 0,
		limit = $bindable(25),
		offset = $bindable(0),
		selectedIds = $bindable([]),
		sortKey = 'date',
		sortDirection = 'desc',
		descriptionFilter = $bindable(''),
		tagFilter = $bindable([]),
		directionFilter = $bindable(null),
		tagOptions = [],
		onSort,
		onPageChange,
		onLimitChange,
		onFilterChange,
		bulkActions,
		class: className = ''
	}: Props = $props();

	function fmtDate(value: string): string {
		return formatDisplayDate(value.slice(0, 10));
	}
</script>

{#snippet descriptionFilterControl()}
	<TextFilter
		bind:value={descriptionFilter}
		label="Description"
		onApply={() => onFilterChange?.()}
	/>
{/snippet}

{#snippet tagFilterControl()}
	<SelectFilter
		options={tagOptions}
		bind:selected={tagFilter}
		label="Tags"
		onApply={() => onFilterChange?.()}
	/>
{/snippet}

{#snippet directionFilterControl()}
	<DirectionFilter bind:value={directionFilter} onApply={() => onFilterChange?.()} />
{/snippet}

{#snippet tagCell(row: CashflowTransaction)}
	{#if row.tag}
		<Badge intent="neutral" variant="soft" size="sm">{row.tag}</Badge>
	{:else}
		<span class="text-slate-500">Untagged</span>
	{/if}
{/snippet}

{#snippet directionCell(row: CashflowTransaction)}
	<Badge intent={row.direction === 'in' ? 'success' : 'error'} variant="soft" size="sm">
		{row.direction === 'in' ? 'In' : 'Out'}
	</Badge>
{/snippet}

{#snippet amountCell(row: CashflowTransaction)}
	<Money amount={scaledToNumber(row.amountCents)} currency="EUR" size="sm" />
{/snippet}

<DataTable
	{rows}
	{loading}
	{error}
	bind:selectedIds
	selectable
	{sortKey}
	{sortDirection}
	{onSort}
	emptyText="No transactions match your filters"
	class={className}
	columns={[
		{
			key: 'date',
			header: 'Date',
			sortKey: 'date',
			value: (r: CashflowTransaction) => fmtDate(r.date)
		},
		{
			key: 'description',
			header: 'Description',
			sortKey: 'description',
			value: (r: CashflowTransaction) => r.description,
			filter: descriptionFilterControl
		},
		{ key: 'tag', header: 'Tag', sortKey: 'tag', cell: tagCell, filter: tagFilterControl },
		{ key: 'direction', header: 'Direction', cell: directionCell, filter: directionFilterControl },
		{ key: 'amount', header: 'Amount', sortKey: 'amount', align: 'right', cell: amountCell }
	]}
>
	{#snippet footer()}
		<FooterBar
			{total}
			bind:limit
			bind:offset
			selectedCount={selectedIds.length}
			{onPageChange}
			{onLimitChange}
			actions={bulkActions}
		/>
	{/snippet}
</DataTable>

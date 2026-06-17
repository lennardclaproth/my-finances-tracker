<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import SortableHeader from './SortableHeader.svelte';
	import { sortableHeaderAligns, type SortDirection } from './sortable-header.types';

	const { Story } = defineMeta({
		title: 'Molecules/SortableHeader',
		component: SortableHeader,
		tags: ['autodocs'],
		argTypes: {
			align: { control: 'select', options: sortableHeaderAligns },
			direction: { control: 'inline-radio', options: ['asc', 'desc'] }
		}
	});
</script>

<script lang="ts">
	// Interactive demo: clicking headers sorts in-place.
	let activeKey = $state('date');
	let direction = $state<SortDirection>('desc');

	function onSort(key: string, dir: SortDirection) {
		activeKey = key;
		direction = dir;
	}
</script>

<Story name="In a table header" asChild>
	<table class="w-full border border-slate-200 text-sm">
		<thead class="border-b border-slate-200 bg-white">
			<tr>
				<SortableHeader label="Date" sortKey="date" {activeKey} {direction} {onSort} />
				<SortableHeader
					label="Description"
					sortKey="description"
					{activeKey}
					{direction}
					{onSort}
				/>
				<SortableHeader label="Tag" sortKey="tag" {activeKey} {direction} {onSort} />
				<SortableHeader
					label="Amount"
					sortKey="amount"
					align="right"
					{activeKey}
					{direction}
					{onSort}
				/>
			</tr>
		</thead>
		<tbody class="text-slate-700">
			<tr class="border-b border-slate-100">
				<td class="px-3 py-2">2026-06-22</td>
				<td class="px-3 py-2">Albert Heijn</td>
				<td class="px-3 py-2">groceries</td>
				<td class="px-3 py-2 text-right">-74.32</td>
			</tr>
			<tr>
				<td class="px-3 py-2">2026-06-25</td>
				<td class="px-3 py-2">Monthly salary</td>
				<td class="px-3 py-2">salary</td>
				<td class="px-3 py-2 text-right">+4200.00</td>
			</tr>
		</tbody>
	</table>
</Story>

<Story name="States" asChild>
	<table class="w-full border border-slate-200 text-sm">
		<thead>
			<tr>
				<SortableHeader label="Inactive" sortKey="a" activeKey="x" />
				<SortableHeader label="Active asc" sortKey="b" activeKey="b" direction="asc" />
				<SortableHeader label="Active desc" sortKey="c" activeKey="c" direction="desc" />
			</tr>
		</thead>
	</table>
</Story>

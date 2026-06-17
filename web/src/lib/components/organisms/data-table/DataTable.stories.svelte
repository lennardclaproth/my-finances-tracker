<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import DataTable from './DataTable.svelte';
	import type { Column } from './data-table.types';

	interface Row {
		id: string;
		name: string;
		role: string;
		value: number;
	}

	const rows: Row[] = [
		{ id: '1', name: 'Alice', role: 'Admin', value: 120 },
		{ id: '2', name: 'Bob', role: 'Editor', value: 84 },
		{ id: '3', name: 'Carol', role: 'Viewer', value: 240 }
	];

	const columns: Column<Row>[] = [
		{ key: 'name', header: 'Name', sortKey: 'name', value: (r) => r.name },
		{ key: 'role', header: 'Role', value: (r) => r.role },
		{ key: 'value', header: 'Value', sortKey: 'value', align: 'right', value: (r) => r.value }
	];

	const { Story } = defineMeta({
		title: 'Organisms/DataTable',
		component: DataTable,
		tags: ['autodocs']
	});
</script>

<Story name="Default" asChild>
	<div class="h-80 rounded-2xl border border-slate-200 bg-white">
		<DataTable {rows} {columns} />
	</div>
</Story>

<Story name="Selectable" asChild>
	<div class="h-80 rounded-2xl border border-slate-200 bg-white">
		<DataTable {rows} {columns} selectable selectedIds={['2']} />
	</div>
</Story>

<Story name="Sorted" asChild>
	<div class="h-80 rounded-2xl border border-slate-200 bg-white">
		<DataTable {rows} {columns} sortKey="value" sortDirection="desc" />
	</div>
</Story>

<Story name="Loading" asChild>
	<div class="h-80 rounded-2xl border border-slate-200 bg-white">
		<DataTable rows={[]} {columns} loading />
	</div>
</Story>

<Story name="Empty" asChild>
	<div class="h-80 rounded-2xl border border-slate-200 bg-white">
		<DataTable rows={[]} {columns} />
	</div>
</Story>

<Story name="Error" asChild>
	<div class="h-80 rounded-2xl border border-slate-200 bg-white">
		<DataTable rows={[]} {columns} error="Failed to load rows" />
	</div>
</Story>

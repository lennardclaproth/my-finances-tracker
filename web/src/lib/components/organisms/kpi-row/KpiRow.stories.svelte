<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import KpiRow from './KpiRow.svelte';
	import type { KpiItem } from './kpi-row.types';

	const items: KpiItem[] = [
		{
			label: 'Market value',
			amount: 12230.4,
			currency: 'EUR',
			change: 12.31,
			trend: [9, 9.6, 9.7, 10.1, 10.8, 12.2]
		},
		{ label: 'Total P&L', amount: 1340.4, currency: 'EUR', change: 12.31 },
		{ label: 'Cost basis', amount: 10890, currency: 'EUR', change: -1.2 }
	];

	const { Story } = defineMeta({
		title: 'Organisms/KpiRow',
		component: KpiRow,
		tags: ['autodocs'],
		argTypes: { columns: { control: 'inline-radio', options: [2, 3, 4] } }
	});
</script>

<Story name="Three KPIs" asChild>
	<KpiRow {items} columns={3} />
</Story>

<Story name="Four KPIs" asChild>
	<KpiRow
		columns={4}
		items={[...items, { label: 'Day return', amount: 86, currency: 'EUR', change: 0.71 }]}
	/>
</Story>

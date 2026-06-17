<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import TimeSeriesChart from './TimeSeriesChart.svelte';
	import { cashflowMonthly } from '$lib/data/fixtures/cashflow';
	import { assetSnapshots } from '$lib/data/fixtures/assets';
	import { scaledToNumber, decimalStringToNumber } from '$lib/api/money';
	import { chartColors } from '$lib/charts/theme';

	const monthShort = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	const monthLabels = cashflowMonthly.map((p) => p.month);
	const assetLabels = assetSnapshots.map((s) => s.date);

	const { Story } = defineMeta({
		title: 'Organisms/Charts/TimeSeriesChart',
		component: TimeSeriesChart,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	let lastRange = $state('');
</script>

<Story name="Trend line (net)" asChild>
	<div class="w-[640px] max-w-full rounded-2xl border border-slate-200 bg-white p-4">
		<TimeSeriesChart
			labels={monthLabels}
			xTickFormat={monthShort}
			datasets={[
				{
					label: 'Net',
					data: cashflowMonthly.map((p) => scaledToNumber(p.net_cents)),
					color: chartColors.net,
					signed: true
				}
			]}
		/>
	</div>
</Story>

<Story name="Combo (income / expense / net)" asChild>
	<div class="w-[640px] max-w-full rounded-2xl border border-slate-200 bg-white p-4">
		<TimeSeriesChart
			labels={monthLabels}
			xTickFormat={monthShort}
			datasets={[
				{
					label: 'Incoming',
					type: 'bar',
					data: cashflowMonthly.map((p) => scaledToNumber(p.incoming_cents)),
					color: chartColors.positive
				},
				{
					label: 'Outgoing',
					type: 'bar',
					data: cashflowMonthly.map((p) => scaledToNumber(p.outgoing_cents)),
					color: chartColors.negative
				},
				{
					label: 'Net',
					type: 'line',
					data: cashflowMonthly.map((p) => scaledToNumber(p.net_cents)),
					color: chartColors.net
				}
			]}
		/>
	</div>
</Story>

<Story name="Area growth" asChild>
	<div class="w-[640px] max-w-full rounded-2xl border border-slate-200 bg-white p-4">
		<TimeSeriesChart
			labels={assetLabels}
			xTickFormat={monthShort}
			datasets={[
				{
					label: 'Total worth',
					data: assetSnapshots.map((s) => decimalStringToNumber(s.total_worth)),
					color: chartColors.positive,
					fill: true
				}
			]}
		/>
	</div>
</Story>

<Story name="Drag range-selection" asChild>
	<div class="w-[640px] max-w-full space-y-2 rounded-2xl border border-slate-200 bg-white p-4">
		<TimeSeriesChart
			labels={monthLabels}
			xTickFormat={monthShort}
			enableRangeSelect
			onRangeSelect={(from, to) => (lastRange = `${from} → ${to}`)}
			datasets={[
				{
					label: 'Net',
					data: cashflowMonthly.map((p) => scaledToNumber(p.net_cents)),
					color: chartColors.net,
					signed: true
				}
			]}
		/>
		<p class="text-xs text-slate-500">
			Click or drag across the chart. Selected: <span class="font-mono text-slate-800"
				>{lastRange || '—'}</span
			>
		</p>
	</div>
</Story>

<Story name="Loading" asChild>
	<div class="w-[640px] max-w-full rounded-2xl border border-slate-200 bg-white p-4">
		<TimeSeriesChart labels={[]} datasets={[]} loading />
	</div>
</Story>

<Story name="Empty" asChild>
	<div class="w-[640px] max-w-full rounded-2xl border border-slate-200 bg-white p-4">
		<TimeSeriesChart labels={[]} datasets={[]} />
	</div>
</Story>

<Story name="Error" asChild>
	<div class="w-[640px] max-w-full rounded-2xl border border-slate-200 bg-white p-4">
		<TimeSeriesChart labels={[]} datasets={[]} error="Failed to load chart data" />
	</div>
</Story>

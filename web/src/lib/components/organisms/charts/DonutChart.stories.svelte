<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import DonutChart from './DonutChart.svelte';
	import { cashflowTagDistribution } from '$lib/data/fixtures/cashflow';
	import { scaledToNumber } from '$lib/api/money';
	import { donutRamps } from '$lib/charts/theme';
	import type { DonutDatum } from '$lib/charts/types';

	const toData = (entries: { tag: string; totalCents: number }[]): DonutDatum[] =>
		entries.map((e) => ({ label: e.tag, value: scaledToNumber(e.totalCents) }));

	const incoming = toData(cashflowTagDistribution.incoming);
	const outgoing = toData(cashflowTagDistribution.outgoing);

	const euro = (n: number) => `€${n.toLocaleString('en', { maximumFractionDigits: 0 })}`;

	const { Story } = defineMeta({
		title: 'Organisms/Charts/DonutChart',
		component: DonutChart,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	let clicked = $state('');
</script>

<Story name="Incoming" asChild>
	<div class="w-96 rounded-2xl border border-slate-200 bg-white p-4">
		<DonutChart
			data={incoming}
			ramp={donutRamps.incoming}
			centerLabel="Incoming"
			formatValue={euro}
		/>
	</div>
</Story>

<Story name="Outgoing (with +N more)" asChild>
	<div class="w-96 rounded-2xl border border-slate-200 bg-white p-4">
		<DonutChart
			data={outgoing}
			ramp={donutRamps.outgoing}
			maxSlices={4}
			centerLabel="Outgoing"
			formatValue={euro}
		/>
	</div>
</Story>

<Story name="Click to filter" asChild>
	<div class="w-96 space-y-2 rounded-2xl border border-slate-200 bg-white p-4">
		<DonutChart
			data={outgoing}
			ramp={donutRamps.outgoing}
			formatValue={euro}
			onSliceClick={(label) => (clicked = label || 'Untagged')}
		/>
		<p class="text-xs text-slate-500">
			Clicked slice: <span class="font-mono text-slate-800">{clicked || '—'}</span>
		</p>
	</div>
</Story>

<Story name="Loading" asChild>
	<div class="w-96 rounded-2xl border border-slate-200 bg-white p-4">
		<DonutChart data={[]} loading />
	</div>
</Story>

<Story name="Empty" asChild>
	<div class="w-96 rounded-2xl border border-slate-200 bg-white p-4">
		<DonutChart data={[]} />
	</div>
</Story>

<Story name="Error" asChild>
	<div class="w-96 rounded-2xl border border-slate-200 bg-white p-4">
		<DonutChart data={[]} error="Failed to load" />
	</div>
</Story>

<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import PageContentTemplate from './PageContentTemplate.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';

	const kpis: KpiItem[] = [
		{ label: 'Market value', amount: 12230.4, currency: 'EUR', change: 12.31 },
		{ label: 'Total P&L', amount: 1340.4, currency: 'EUR', change: 12.31 },
		{ label: 'Cost basis', amount: 10890, currency: 'EUR' }
	];

	const { Story } = defineMeta({
		title: 'Templates/PageContentTemplate',
		component: PageContentTemplate,
		tags: ['autodocs']
	});
</script>

<Story name="Analytics + content + FAB" asChild>
	<div class="h-[32rem] bg-taupe-100">
		<PageContentTemplate showFab fabLabel="New">
			{#snippet analytics()}
				<KpiRow items={kpis} columns={3} />
			{/snippet}
			<div class="flex h-full items-center justify-center text-sm text-slate-500">
				Primary content panel
			</div>
		</PageContentTemplate>
	</div>
</Story>

<Story name="No analytics" asChild>
	<div class="h-[24rem] bg-taupe-100">
		<PageContentTemplate>
			<div class="flex h-full items-center justify-center text-sm text-slate-500">
				Full-height content
			</div>
		</PageContentTemplate>
	</div>
</Story>

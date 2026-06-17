<script lang="ts">
	import StatCard from '$lib/components/organisms/stat-card/StatCard.svelte';
	import type { KpiItem } from './kpi-row.types';

	type Props = {
		items: KpiItem[];
		columns?: 2 | 3 | 4;
		class?: string;
	};

	let { items, columns = 3, class: className = '' }: Props = $props();

	const colClasses = {
		2: 'sm:grid-cols-2',
		3: 'sm:grid-cols-2 lg:grid-cols-3',
		4: 'sm:grid-cols-2 lg:grid-cols-4'
	} satisfies Record<2 | 3 | 4, string>;
</script>

<div class={['grid grid-cols-1 gap-3', colClasses[columns], className].filter(Boolean).join(' ')}>
	{#each items as kpi (kpi.label)}
		<StatCard
			label={kpi.label}
			amount={kpi.amount}
			currency={kpi.currency}
			locale={kpi.locale}
			change={kpi.change}
			changeFormat={kpi.changeFormat}
			trend={kpi.trend}
		/>
	{/each}
</div>

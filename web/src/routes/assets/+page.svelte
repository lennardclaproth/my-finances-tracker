<script lang="ts">
	import { onMount } from 'svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import DonutChart from '$lib/components/organisms/charts/DonutChart.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import AssetClassDrawer from '$lib/components/organisms/asset-class-drawer/AssetClassDrawer.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import { listAssetClasses, getAssetSnapshots, getAssetClassDetails } from '$lib/services/assets';
	import { adminMode } from '$lib/stores/admin.svelte';
	import { decimalStringToNumber } from '$lib/api/money';
	import { donutRamps } from '$lib/charts/theme';
	import { DEMO_ACCOUNT_ID } from '$lib/api/config';
	import type { AssetClass, AssetClassDetails, AssetSnapshotPoint } from '$lib/api/types';

	let classes = $state<AssetClass[]>([]);
	let snapshots = $state<AssetSnapshotPoint[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let drawerOpen = $state(false);
	let details = $state<AssetClassDetails | null>(null);
	let detailsLoading = $state(false);

	const euro = (n: number) => `€${n.toLocaleString('en', { maximumFractionDigits: 0 })}`;
	const monthShort = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	const distribution = $derived(
		classes
			.filter((c) => !c.archived)
			.map((c) => ({ label: c.name, value: decimalStringToNumber(c.current_worth) }))
	);

	async function loadAll() {
		loading = true;
		error = null;
		try {
			const [cls, snaps] = await Promise.all([
				listAssetClasses({ account_id: DEMO_ACCOUNT_ID }),
				getAssetSnapshots({ account_id: DEMO_ACCOUNT_ID })
			]);
			classes = cls;
			snapshots = snaps;
		} catch {
			error = 'Failed to load assets';
		} finally {
			loading = false;
		}
	}

	onMount(loadAll);

	async function openClass(row: AssetClass) {
		drawerOpen = true;
		detailsLoading = true;
		details = null;
		try {
			details = await getAssetClassDetails(row.id, DEMO_ACCOUNT_ID);
		} finally {
			detailsLoading = false;
		}
	}
</script>

{#snippet worthCell(row: AssetClass)}
	<Money amount={decimalStringToNumber(row.current_worth)} currency="EUR" size="sm" />
{/snippet}

{#snippet growthCell(row: AssetClass)}
	{#if row.growth_pct !== null && row.growth_pct !== undefined}
		<Badge intent={row.growth_pct >= 0 ? 'success' : 'error'} variant="soft" size="sm">
			{row.growth_pct >= 0 ? '+' : ''}{row.growth_pct.toFixed(2)}%
		</Badge>
	{:else}
		<span class="text-slate-400">—</span>
	{/if}
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Assets"
			breadcrumb={[{ label: 'Home', href: '/' }, { label: 'Assets' }]}
			showDateRange
			accountName="Lennard Claproth"
			adminMode={adminMode.enabled}
			onAdminToggle={(v) => adminMode.set(v)}
		/>
	{/snippet}

	<PageContentTemplate showFab fabLabel="New asset class">
		{#snippet analytics()}
			<div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
				<div class="rounded-2xl border border-slate-200 bg-white p-3 lg:col-span-2">
					<p class="mb-1 text-xs font-medium text-slate-500">Total worth</p>
					<TimeSeriesChart
						height="h-52"
						{loading}
						labels={snapshots.map((s) => s.date)}
						xTickFormat={monthShort}
						datasets={[
							{
								label: 'Total worth',
								data: snapshots.map((s) => decimalStringToNumber(s.total_worth)),
								color: '#059669',
								fill: true
							}
						]}
					/>
				</div>
				<div class="rounded-2xl border border-slate-200 bg-white p-3">
					<p class="mb-1 text-xs font-medium text-slate-500">Distribution</p>
					<DonutChart
						data={distribution}
						ramp={donutRamps.incoming}
						{loading}
						formatValue={euro}
						centerLabel="Total"
					/>
				</div>
			</div>
		{/snippet}

		<DataTable
			rows={classes}
			{loading}
			{error}
			emptyText="No asset classes"
			onRowClick={openClass}
			columns={[
				{ key: 'name', header: 'Class', value: (r: AssetClass) => r.name },
				{ key: 'source', header: 'Source', value: (r: AssetClass) => r.source },
				{ key: 'worth', header: 'Current worth', align: 'right', cell: worthCell },
				{ key: 'growth', header: 'Growth', align: 'right', cell: growthCell }
			]}
		/>
	</PageContentTemplate>
</AppShellTemplate>

<AssetClassDrawer bind:open={drawerOpen} {details} loading={detailsLoading} />

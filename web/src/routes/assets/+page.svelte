<script lang="ts">
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import DonutChart from '$lib/components/organisms/charts/DonutChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
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

	let from = $state('');
	let to = $state('');

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
				getAssetSnapshots({
					account_id: DEMO_ACCOUNT_ID,
					from: from || undefined,
					to: to || undefined
				})
			]);
			classes = cls;
			snapshots = snaps;
		} catch {
			error = 'Failed to load assets';
		} finally {
			loading = false;
		}
	}

	// Reload (and zoom the worth chart) whenever the date range changes; also the initial load.
	$effect(() => {
		void from;
		void to;
		void loadAll();
	});

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
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Assets"
			showDateRange
			dateFrom={from || null}
			dateTo={to || null}
			onDateChange={(r) => {
				from = r.from ?? '';
				to = r.to ?? '';
			}}
			accountName="Lennard Claproth"
			adminMode={adminMode.enabled}
			onAdminToggle={(v) => adminMode.set(v)}
		/>
	{/snippet}

	<PageContentTemplate showFab fabLabel="New asset class">
		{#snippet analytics()}
			<div class="grid grid-cols-1 gap-3 lg:grid-cols-3">
				<AnalyticsCard title="Total worth" class="lg:col-span-2">
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
				</AnalyticsCard>
				<AnalyticsCard title="Distribution">
					<DonutChart
						data={distribution}
						ramp={donutRamps.incoming}
						{loading}
						formatValue={euro}
						centerLabel="Total"
					/>
				</AnalyticsCard>
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

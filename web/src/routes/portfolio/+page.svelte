<script lang="ts">
	import { onMount } from 'svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Switch from '$lib/components/atoms/switch/Switch.svelte';
	import {
		listPortfolioPositions,
		getPortfolioSnapshots,
		listPortfolioTransactions
	} from '$lib/services/portfolio';
	import { adminMode } from '$lib/stores/admin.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { chartColors } from '$lib/charts/theme';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import { DEMO_ACCOUNT_ID } from '$lib/api/config';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';
	import type {
		PortfolioPosition,
		PortfolioSnapshotPoint,
		PortfolioTransaction
	} from '$lib/api/types';

	let positions = $state<PortfolioPosition[]>([]);
	let snapshots = $state<PortfolioSnapshotPoint[]>([]);
	let transactions = $state<PortfolioTransaction[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let includeClosed = $state(false);
	let tab = $state('positions');

	const tabs = [
		{ value: 'positions', label: 'Positions' },
		{ value: 'transactions', label: 'Transactions' }
	];

	const monthShort = (iso: string) =>
		new Date(iso).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	const latest = $derived(snapshots.at(-1) ?? null);
	const kpis = $derived<KpiItem[]>(
		latest
			? [
					{
						label: 'Market value',
						amount: scaledToNumber(latest.market_value),
						currency: 'EUR',
						change: latest.total_pnl_pct
					},
					{
						label: 'Total P&L',
						amount: scaledToNumber(latest.total_pnl),
						currency: 'EUR',
						change: latest.return_vs_cost_basis_pct
					},
					{ label: 'Cost basis', amount: scaledToNumber(latest.total_cost_basis), currency: 'EUR' }
				]
			: []
	);

	async function loadPositions() {
		positions = (
			await listPortfolioPositions({ account_id: DEMO_ACCOUNT_ID, include_closed: includeClosed })
		).data;
	}

	async function loadAll() {
		loading = true;
		error = null;
		try {
			const [snaps, txs] = await Promise.all([
				getPortfolioSnapshots({ account_id: DEMO_ACCOUNT_ID }),
				listPortfolioTransactions({ account_id: DEMO_ACCOUNT_ID, limit: 25 })
			]);
			snapshots = snaps;
			transactions = txs.data;
			await loadPositions();
		} catch {
			error = 'Failed to load portfolio';
		} finally {
			loading = false;
		}
	}

	onMount(loadAll);

	$effect(() => {
		// Reads `includeClosed` synchronously, so it reloads positions when the toggle changes.
		void loadPositions();
	});
</script>

{#snippet marketValueCell(row: PortfolioPosition)}
	{#if row.market_value !== null && row.market_value !== undefined}
		<Money amount={scaledToNumber(row.market_value)} currency="EUR" size="sm" />
	{:else}
		<span class="text-slate-400">—</span>
	{/if}
{/snippet}

{#snippet pnlCell(row: PortfolioPosition)}
	{#if row.unrealized_pnl_pct !== null && row.unrealized_pnl_pct !== undefined}
		<Badge intent={row.unrealized_pnl_pct >= 0 ? 'success' : 'error'} variant="soft" size="sm">
			{row.unrealized_pnl_pct >= 0 ? '+' : ''}{row.unrealized_pnl_pct.toFixed(2)}%
		</Badge>
	{:else}
		<span class="text-slate-400">—</span>
	{/if}
{/snippet}

{#snippet txTypeCell(row: PortfolioTransaction)}
	<Badge intent="neutral" variant="soft" size="sm">{row.type}</Badge>
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Portfolio"
			breadcrumb={[{ label: 'Home', href: '/' }, { label: 'Portfolio' }]}
			accountName="Lennard Claproth"
			adminMode={adminMode.enabled}
			onAdminToggle={(v) => adminMode.set(v)}
		/>
	{/snippet}

	<PageContentTemplate showFab fabLabel="New transaction">
		{#snippet analytics()}
			<div class="flex flex-col gap-3">
				<KpiRow items={kpis} columns={3} />
				<div class="rounded-2xl border border-slate-200 bg-white p-3">
					<p class="mb-1 text-xs font-medium text-slate-500">Value vs cost basis</p>
					<TimeSeriesChart
						height="h-52"
						{loading}
						labels={snapshots.map((s) => s.occurred_at.slice(0, 10))}
						xTickFormat={monthShort}
						datasets={[
							{
								label: 'Market value',
								data: snapshots.map((s) => scaledToNumber(s.market_value)),
								color: chartColors.positive,
								fill: true
							},
							{
								label: 'Cost basis',
								data: snapshots.map((s) => scaledToNumber(s.total_cost_basis)),
								color: chartColors.net,
								dashed: true
							}
						]}
					/>
				</div>
			</div>
		{/snippet}

		<div class="flex min-h-0 flex-1 flex-col">
			<div class="flex items-center justify-between gap-3 border-b border-slate-200 px-3 py-2">
				<Tabs {tabs} bind:value={tab} ariaLabel="Portfolio view" />
				{#if tab === 'positions'}
					<label class="flex items-center gap-2 text-sm text-slate-600">
						<span>Include closed</span>
						<Switch bind:checked={includeClosed} aria-label="Include closed positions" />
					</label>
				{/if}
			</div>

			{#if tab === 'positions'}
				<DataTable
					rows={positions}
					{loading}
					{error}
					emptyText="No positions"
					columns={[
						{ key: 'symbol', header: 'Symbol', value: (r: PortfolioPosition) => r.symbol ?? '—' },
						{ key: 'name', header: 'Name', value: (r: PortfolioPosition) => r.name ?? '—' },
						{
							key: 'quantity',
							header: 'Qty',
							align: 'right',
							value: (r: PortfolioPosition) => r.quantity
						},
						{ key: 'market_value', header: 'Market value', align: 'right', cell: marketValueCell },
						{ key: 'pnl', header: 'Unrealized', align: 'right', cell: pnlCell }
					]}
				/>
			{:else}
				<DataTable
					rows={transactions}
					{loading}
					{error}
					emptyText="No transactions"
					columns={[
						{
							key: 'date',
							header: 'Date',
							value: (r: PortfolioTransaction) => formatDisplayDate(r.occurred_at.slice(0, 10))
						},
						{ key: 'type', header: 'Type', cell: txTypeCell },
						{
							key: 'symbol',
							header: 'Symbol',
							value: (r: PortfolioTransaction) => r.symbol ?? '—'
						},
						{
							key: 'quantity',
							header: 'Qty',
							align: 'right',
							value: (r: PortfolioTransaction) => r.quantity
						},
						{
							key: 'amount',
							header: 'Amount',
							align: 'right',
							value: (r: PortfolioTransaction) => r.amount
						}
					]}
				/>
			{/if}
		</div>
	</PageContentTemplate>
</AppShellTemplate>

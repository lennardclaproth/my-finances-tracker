<script lang="ts">
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Tabs from '$lib/components/molecules/tabs/Tabs.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import Switch from '$lib/components/atoms/switch/Switch.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import DatePicker from '$lib/components/molecules/date-picker/DatePicker.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import ListingSearchSelect from '$lib/components/molecules/listing-search-select/ListingSearchSelect.svelte';
	import {
		listPortfolioPositions,
		getPortfolioSnapshots,
		listPortfolioTransactions,
		createManualPortfolioTransaction,
		rebuildPortfolio
	} from '$lib/services/portfolio';
	import { listVendors } from '$lib/services/vendors';
	import { adminMode } from '$lib/stores/admin.svelte';
	import { accountStore } from '$lib/stores/account.svelte';
	import { toast } from '$lib/stores/toast.svelte';
	import { scaledToNumber } from '$lib/api/money';
	import { chartColors } from '$lib/charts/theme';
	import {
		formatDisplayDate,
		todayISO
	} from '$lib/components/molecules/calendar/calendar.utils';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';
	import type {
		Listing,
		PortfolioPosition,
		PortfolioSnapshotPoint,
		PortfolioTransaction,
		PortfolioTransactionType,
		Vendor
	} from '$lib/api/types';

	let positions = $state<PortfolioPosition[]>([]);
	let snapshots = $state<PortfolioSnapshotPoint[]>([]);
	let transactions = $state<PortfolioTransaction[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let includeClosed = $state(false);
	let tab = $state('positions');
	let from = $state('');
	let to = $state('');
	let searchQuery = $state('');

	let txOpen = $state(false);
	let vendors = $state<Vendor[]>([]);
	let vendorId = $state('');
	let txType = $state<PortfolioTransactionType>('BUY');
	let txListing = $state<Listing | null>(null);
	let txQuantity = $state('');
	let txAmount = $state('');
	let txDate = $state(todayISO());
	let txDescription = $state('');
	let creatingTx = $state(false);
	let rebuilding = $state(false);

	const tabs = [
		{ value: 'positions', label: 'Positions' },
		{ value: 'transactions', label: 'Transactions' }
	];

	const txTypeOptions = [
		{ value: 'BUY', label: 'Buy' },
		{ value: 'SELL', label: 'Sell' },
		{ value: 'DIVIDEND', label: 'Dividend' },
		{ value: 'TAX', label: 'Tax' },
		{ value: 'FEE', label: 'Fee' },
		{ value: 'CASH', label: 'Cash' }
	];

	// Backend rules: non-CASH types require a listing; BUY/SELL additionally require a quantity.
	const needsListing = $derived(txType !== 'CASH');
	const needsQuantity = $derived(txType === 'BUY' || txType === 'SELL');
	const vendorOptions = $derived(vendors.map((v) => ({ value: v.id, label: v.name })));

	const navActions: MenuItem[] = [
		{ label: 'Rebuild portfolio', icon: 'heroicons:arrow-path', onSelect: () => void handleRebuild() }
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
		await accountStore.ensureLoaded();
		positions = (
			await listPortfolioPositions({
				account_id: accountStore.activeId,
				include_closed: includeClosed
			})
		).data;
	}

	async function loadAll() {
		loading = true;
		error = null;
		try {
			await accountStore.ensureLoaded();
			const accountId = accountStore.activeId;
			const [snaps, txs] = await Promise.all([
				getPortfolioSnapshots({
					account_id: accountId,
					from: from || undefined,
					to: to || undefined
				}),
				listPortfolioTransactions({
					account_id: accountId,
					limit: 25,
					q: searchQuery || undefined,
					from: from || undefined,
					to: to || undefined
				})
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

	// Reload (zooming the value chart, filtering transactions) on date/search change; also initial load.
	$effect(() => {
		void from;
		void to;
		void searchQuery;
		void loadAll();
	});

	$effect(() => {
		// Reads `includeClosed` synchronously, so it reloads positions when the toggle changes.
		void loadPositions();
	});

	async function openTx() {
		txType = 'BUY';
		txListing = null;
		txQuantity = '';
		txAmount = '';
		txDate = todayISO();
		txDescription = '';
		txOpen = true;
		if (vendors.length === 0) {
			try {
				const all = await listVendors();
				// Manual portfolio transactions require a brokerage/portfolio vendor.
				vendors = all.filter((v) => v.active && v.type === 'portfolio');
				if (vendors.length > 0) vendorId = vendors[0].id;
			} catch {
				// Leave the vendor list empty; the form will warn on submit.
			}
		}
	}

	async function submitTx() {
		if (!vendorId || txAmount.trim() === '') {
			toast.error('Vendor and amount are required');
			return;
		}
		if (needsListing && !txListing) {
			toast.error('Select a listing');
			return;
		}
		if (needsQuantity && txQuantity.trim() === '') {
			toast.error('Quantity is required');
			return;
		}
		creatingTx = true;
		try {
			await accountStore.ensureLoaded();
			await createManualPortfolioTransaction({
				account_id: accountStore.activeId,
				vendor_id: vendorId,
				occurred_at: txDate,
				type: txType,
				listing_id: needsListing && txListing ? txListing.id : undefined,
				amount: txAmount.trim(),
				quantity: needsQuantity ? txQuantity.trim() : undefined,
				description: txDescription.trim() || undefined
			});
			txOpen = false;
			toast.success('Transaction added');
			void loadAll();
		} catch {
			toast.error('Failed to add transaction');
		} finally {
			creatingTx = false;
		}
	}

	async function handleRebuild() {
		if (rebuilding) return;
		rebuilding = true;
		try {
			await accountStore.ensureLoaded();
			await rebuildPortfolio({ account_id: accountStore.activeId });
			toast.success('Portfolio rebuild started');
		} catch {
			toast.error('Failed to start rebuild');
		} finally {
			rebuilding = false;
		}
	}
</script>

{#snippet marketValueCell(row: PortfolioPosition)}
	{#if row.market_value !== null && row.market_value !== undefined}
		<Money amount={scaledToNumber(row.market_value)} currency="EUR" size="sm" />
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

{#snippet pnlCell(row: PortfolioPosition)}
	{#if row.unrealized_pnl_pct !== null && row.unrealized_pnl_pct !== undefined}
		<Badge intent={row.unrealized_pnl_pct >= 0 ? 'success' : 'error'} variant="soft" size="sm">
			{row.unrealized_pnl_pct >= 0 ? '+' : ''}{row.unrealized_pnl_pct.toFixed(2)}%
		</Badge>
	{:else}
		<span class="text-slate-500">—</span>
	{/if}
{/snippet}

{#snippet txTypeCell(row: PortfolioTransaction)}
	<Badge intent="neutral" variant="soft" size="sm">{row.type}</Badge>
{/snippet}

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Portfolio"
			showSearch
			searchValue={searchQuery}
			searchPlaceholder="Search transactions…"
			onSearch={(q) => (searchQuery = q)}
			showDateRange
			dateFrom={from || null}
			dateTo={to || null}
			onDateChange={(r) => {
				from = r.from ?? '';
				to = r.to ?? '';
			}}
			actions={navActions}
			accountName="Lennard Claproth"
			adminMode={adminMode.enabled}
			onAdminToggle={(v) => adminMode.set(v)}
		/>
	{/snippet}

	<PageContentTemplate showFab fabLabel="New transaction" onFabClick={openTx}>
		{#snippet analytics()}
			<div class="flex flex-col gap-3">
				<KpiRow items={kpis} columns={3} />
				<AnalyticsCard title="Value vs cost basis">
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
				</AnalyticsCard>
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

<Dialog bind:open={txOpen} title="New transaction" size="md">
	<div class="space-y-3">
		<div class="grid grid-cols-2 gap-3">
			<FormField label="Date" id="ptx-date">
				<DatePicker value={txDate} onChange={(v) => (txDate = v ?? txDate)} class="w-full" />
			</FormField>
			<FormField label="Type" id="ptx-type">
				{#snippet children(ctx)}
					<Select
						id={ctx.id}
						bind:value={txType}
						options={txTypeOptions}
						ariaLabel="Transaction type"
					/>
				{/snippet}
			</FormField>
		</div>

		<FormField label="Vendor" id="ptx-vendor" hint="Brokerage account">
			{#snippet children(ctx)}
				<Select
					id={ctx.id}
					bind:value={vendorId}
					options={vendorOptions}
					placeholder="Select a vendor"
					ariaLabel="Vendor"
				/>
			{/snippet}
		</FormField>

		{#if needsListing}
			<FormField label="Listing" id="ptx-listing">
				{#snippet children()}
					<ListingSearchSelect bind:value={txListing} ariaLabel="Listing" />
				{/snippet}
			</FormField>
		{/if}

		<div class="grid grid-cols-2 gap-3">
			{#if needsQuantity}
				<FormField label="Quantity" id="ptx-quantity">
					{#snippet children(ctx)}
						<Input id={ctx.id} bind:value={txQuantity} placeholder="e.g. 10" />
					{/snippet}
				</FormField>
			{/if}
			<FormField label="Amount" id="ptx-amount">
				{#snippet children(ctx)}
					<CurrencyInput id={ctx.id} bind:value={txAmount} ariaDescribedby={ctx.describedby} />
				{/snippet}
			</FormField>
		</div>

		<FormField label="Description" id="ptx-description" hint="Optional">
			{#snippet children(ctx)}
				<Input id={ctx.id} bind:value={txDescription} />
			{/snippet}
		</FormField>
	</div>

	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (txOpen = false)}>Cancel</Button>
		<Button intent="success" onclick={submitTx} loading={creatingTx}>Save</Button>
	{/snippet}
</Dialog>

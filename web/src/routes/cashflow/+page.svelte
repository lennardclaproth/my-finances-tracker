<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import TimeSeriesChart from '$lib/components/organisms/charts/TimeSeriesChart.svelte';
	import DonutChart from '$lib/components/organisms/charts/DonutChart.svelte';
	import AnalyticsCard from '$lib/components/molecules/analytics-card/AnalyticsCard.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import CashflowTransactionsTable from '$lib/components/organisms/cashflow-transactions-table/CashflowTransactionsTable.svelte';
	import TransactionFormModal from '$lib/components/organisms/transaction-form-modal/TransactionFormModal.svelte';
	import {
		listCashflowTransactions,
		getCashflowMonthly,
		getCashflowTagDistribution,
		createCashflowTransactions,
		tagCashflowTransactionsBySelection
	} from '$lib/services/cashflow';
	import { connectRealtime } from '$lib/services/realtime';
	import { toast } from '$lib/stores/toast.svelte';
	import { adminMode } from '$lib/stores/admin.svelte';
	import { accountStore } from '$lib/stores/account.svelte';
	import type { CashflowTransactionFormValue } from '$lib/components/organisms/transaction-form-modal/transaction-form-modal.types';
	import {
		parseQuery,
		serializeQuery,
		type QuerySchema,
		type QueryState
	} from '$lib/url/routeQuery';
	import { pushQuery } from '$lib/url/queryState';
	import { scaledToNumber } from '$lib/api/money';
	import { chartColors, donutRamps } from '$lib/charts/theme';
	import type {
		CashflowDirection,
		CashflowTransaction,
		CashflowTransactionsQuery,
		CashflowMonthlyPoint,
		TagDistributionEntry
	} from '$lib/api/types';
	import type { SortDirection } from '$lib/components/organisms/data-table/data-table.types';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	const schema: QuerySchema = {
		description: { type: 'string' },
		tags: { type: 'string[]' },
		direction: { type: 'string' },
		sort_by: { type: 'string' },
		sort_order: { type: 'string' },
		limit: { type: 'number' },
		offset: { type: 'number' },
		from: { type: 'string' },
		to: { type: 'string' }
	};

	const initial = parseQuery(page.url.searchParams, schema);

	let descriptionFilter = $state((initial.description as string) || '');
	let tagFilter = $state(initial.tags as string[]);
	let directionFilter = $state(((initial.direction as string) || null) as CashflowDirection | null);
	let sortKey = $state((initial.sort_by as string) || 'date');
	let sortDirection = $state(((initial.sort_order as string) || 'desc') as SortDirection);
	let limit = $state((initial.limit as number) || 25);
	let offset = $state((initial.offset as number) || 0);
	let from = $state((initial.from as string) || '');
	let to = $state((initial.to as string) || '');

	let rows = $state<CashflowTransaction[]>([]);
	let total = $state(0);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let selectedIds = $state<string[]>([]);

	let monthly = $state<CashflowMonthlyPoint[]>([]);
	let incoming = $state<TagDistributionEntry[]>([]);
	let outgoing = $state<TagDistributionEntry[]>([]);
	let analyticsLoading = $state(true);

	let createOpen = $state(false);
	let creating = $state(false);
	let createError = $state<string | null>(null);

	const tagOptions = $derived(
		incoming
			.concat(outgoing)
			.map((entry) => entry.tag)
			.filter((tag, index, all) => tag !== '' && all.indexOf(tag) === index)
			.map((tag) => ({ value: tag, label: tag }))
	);

	const euro = (n: number) => `€${n.toLocaleString('en', { maximumFractionDigits: 0 })}`;
	const monthShort = (iso: string) =>
		new Date(`${iso}T00:00:00Z`).toLocaleDateString('en', { month: 'short', timeZone: 'UTC' });

	function currentQuery(): CashflowTransactionsQuery {
		return {
			description: descriptionFilter || undefined,
			tags: tagFilter.join(',') || undefined,
			direction: directionFilter ?? undefined,
			sort_by: (sortKey || 'date') as CashflowTransactionsQuery['sort_by'],
			sort_order: sortDirection,
			limit,
			offset,
			from: from || undefined,
			to: to || undefined,
			hide_ignored: true
		};
	}

	function urlState(): QueryState {
		return {
			description: descriptionFilter,
			tags: tagFilter,
			direction: directionFilter ?? '',
			sort_by: sortKey,
			sort_order: sortDirection,
			limit,
			offset,
			from,
			to
		};
	}

	async function load(query: CashflowTransactionsQuery) {
		loading = true;
		error = null;
		try {
			const result = await listCashflowTransactions(query);
			rows = result.data;
			total = result.pagination.total;
		} catch {
			error = 'Failed to load transactions';
		} finally {
			loading = false;
		}
	}

	function syncUrl() {
		void pushQuery('/cashflow', serializeQuery(urlState(), schema), { replace: true });
	}

	// Reload whenever the working state changes (initial + every filter/sort/page change).
	$effect(() => {
		void load(currentQuery());
	});

	async function loadAnalytics() {
		analyticsLoading = true;
		try {
			const range = { from: from || undefined, to: to || undefined };
			const [monthlyRes, dist] = await Promise.all([
				getCashflowMonthly(range),
				getCashflowTagDistribution(range)
			]);
			monthly = monthlyRes.data;
			incoming = dist.incoming;
			outgoing = dist.outgoing;
		} finally {
			analyticsLoading = false;
		}
	}

	// Re-fetch the trend + donuts whenever the date range changes (also covers the initial load).
	$effect(() => {
		void from;
		void to;
		void loadAnalytics();
	});

	onMount(() => {
		let realtime: { disconnect: () => void } | null = null;
		void accountStore.ensureLoaded().then(() => {
			realtime = connectRealtime({
				accountId: accountStore.activeId,
				events: ['import.completed', 'bulk_tag.completed'],
				onRefresh: () => {
					void load(currentQuery());
					void loadAnalytics();
				}
			});
		});
		return () => realtime?.disconnect();
	});

	function onSort(key: string, direction: SortDirection) {
		sortKey = key;
		sortDirection = direction;
		syncUrl();
	}
	function onPageChange() {
		syncUrl();
	}
	function onLimitChange() {
		offset = 0;
		syncUrl();
	}
	function onFilterChange() {
		offset = 0;
		syncUrl();
	}
	function onRangeSelect(rangeFrom: string, rangeTo: string) {
		from = rangeFrom;
		to = rangeTo;
		offset = 0;
		syncUrl();
		toast.info(`Filtered to ${rangeFrom} – ${rangeTo}`);
	}

	async function handleCreate(value: CashflowTransactionFormValue) {
		creating = true;
		createError = null;
		try {
			await accountStore.ensureLoaded();
			await createCashflowTransactions({
				account_id: accountStore.activeId,
				transactions: [
					{
						date: value.date,
						amount: value.amount,
						type: value.type,
						description: value.description,
						// The backend requires a non-blank note and tag per row.
						note: value.note || value.description,
						tag: value.tag || 'Uncategorized'
					}
				]
			});
			createOpen = false;
			toast.success('Transaction created');
			void load(currentQuery());
			void loadAnalytics();
		} catch {
			createError = 'Failed to create transaction';
			toast.error('Failed to create transaction');
		} finally {
			creating = false;
		}
	}

	async function handleBulkTag() {
		const tag = window.prompt('Tag for the selected transactions')?.trim();
		if (!tag) return;
		const ids = selectedIds;
		try {
			await tagCashflowTransactionsBySelection({ tag, ids });
			toast.success(`Tagged ${ids.length} transactions`);
			selectedIds = [];
			void load(currentQuery());
			void loadAnalytics();
		} catch {
			toast.error('Failed to tag transactions');
		}
	}

	const navActions: MenuItem[] = [
		{ label: 'New transaction', icon: 'heroicons:plus', onSelect: () => (createOpen = true) },
		{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' }
	];
	const accountItems: MenuItem[] = [
		{ label: 'Account settings', icon: 'heroicons:cog-6-tooth' },
		{ label: 'Sign out', icon: 'heroicons:arrow-right-on-rectangle', intent: 'danger' }
	];
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Cashflow"
			showSearch
			searchValue={descriptionFilter}
			searchPlaceholder="Search description…"
			onSearch={(q) => {
				descriptionFilter = q;
				onFilterChange();
			}}
			showDateRange
			dateFrom={from || null}
			dateTo={to || null}
			onDateChange={(r) => {
				from = r.from ?? '';
				to = r.to ?? '';
				offset = 0;
				syncUrl();
			}}
			actions={navActions}
			accountName="Lennard Claproth"
			accountEmail="lennard@example.com"
			adminMode={adminMode.enabled}
			onAdminToggle={(v) => adminMode.set(v)}
			{accountItems}
		/>
	{/snippet}

	<PageContentTemplate showFab fabLabel="New transaction" onFabClick={() => (createOpen = true)}>
		{#snippet analytics()}
			<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
				<AnalyticsCard title="Net trend" class="lg:col-span-2">
					<TimeSeriesChart
						height="h-44"
						labels={monthly.map((m) => m.month)}
						xTickFormat={monthShort}
						loading={analyticsLoading}
						enableRangeSelect
						{onRangeSelect}
						datasets={[
							{
								label: 'Net',
								data: monthly.map((m) => scaledToNumber(m.net_cents)),
								color: chartColors.net,
								signed: true
							}
						]}
					/>
				</AnalyticsCard>
				<AnalyticsCard title="Incoming">
					<DonutChart
						data={incoming.map((e) => ({ label: e.tag, value: scaledToNumber(e.totalCents) }))}
						ramp={donutRamps.incoming}
						loading={analyticsLoading}
						formatValue={euro}
						centerLabel="In"
					/>
				</AnalyticsCard>
				<AnalyticsCard title="Outgoing">
					<DonutChart
						data={outgoing.map((e) => ({ label: e.tag, value: scaledToNumber(e.totalCents) }))}
						ramp={donutRamps.outgoing}
						loading={analyticsLoading}
						formatValue={euro}
						centerLabel="Out"
					/>
				</AnalyticsCard>
			</div>
		{/snippet}

		<CashflowTransactionsTable
			{rows}
			{loading}
			{error}
			{total}
			bind:limit
			bind:offset
			bind:selectedIds
			{sortKey}
			{sortDirection}
			bind:descriptionFilter
			bind:tagFilter
			bind:directionFilter
			{tagOptions}
			{onSort}
			{onPageChange}
			{onLimitChange}
			{onFilterChange}
		>
			{#snippet bulkActions()}
				<Button size="sm" variant="ghost" intent="secondary" onclick={handleBulkTag}>Tag</Button>
			{/snippet}
		</CashflowTransactionsTable>
	</PageContentTemplate>
</AppShellTemplate>

<TransactionFormModal
	bind:open={createOpen}
	onSubmit={handleCreate}
	submitting={creating}
	error={createError}
/>

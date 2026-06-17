<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import AppShellTemplate from './AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import KpiRow from '$lib/components/organisms/kpi-row/KpiRow.svelte';
	import CashflowTransactionsTable from '$lib/components/organisms/cashflow-transactions-table/CashflowTransactionsTable.svelte';
	import { cashflowTransactions } from '$lib/data/fixtures/cashflow';
	import type { KpiItem } from '$lib/components/organisms/kpi-row/kpi-row.types';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	const kpis: KpiItem[] = [
		{ label: 'Incoming', amount: 5400, currency: 'EUR', change: 4.9 },
		{ label: 'Outgoing', amount: 3815.5, currency: 'EUR', change: -2.1 },
		{ label: 'Net', amount: 1584.5, currency: 'EUR', change: 12.3 }
	];
	const actions: MenuItem[] = [{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' }];
	const tagOptions = [
		{ value: 'salary', label: 'salary' },
		{ value: 'rent', label: 'rent' },
		{ value: 'groceries', label: 'groceries' }
	];

	const { Story } = defineMeta({
		title: 'Templates/AppShellTemplate',
		component: AppShellTemplate,
		tags: ['autodocs']
	});
</script>

<Story name="Cashflow page" asChild>
	<AppShellTemplate>
		{#snippet top()}
			<TopNavbar
				title="Cashflow"
				showSearch
				showDateRange
				{actions}
				accountName="Lennard Claproth"
				accountEmail="lennard@example.com"
			/>
		{/snippet}

		<PageContentTemplate showFab fabLabel="New transaction">
			{#snippet analytics()}
				<KpiRow items={kpis} columns={3} />
			{/snippet}
			<CashflowTransactionsTable
				rows={cashflowTransactions.slice(0, 10)}
				total={cashflowTransactions.length}
				{tagOptions}
			/>
		</PageContentTemplate>
	</AppShellTemplate>
</Story>

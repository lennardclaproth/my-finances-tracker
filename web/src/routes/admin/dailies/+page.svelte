<script lang="ts">
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import ListingSearchSelect from '$lib/components/molecules/listing-search-select/ListingSearchSelect.svelte';
	import { getEOD } from '$lib/services/marketdata';
	import { scaledToNumber } from '$lib/api/money';
	import { adminMode } from '$lib/stores/admin.svelte';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { EOD, Listing } from '$lib/api/types';

	let selected = $state<Listing | null>(null);
	let rows = $state<EOD[]>([]);
	let loading = $state(false);
	let error = $state<string | null>(null);

	const currency = $derived(selected?.currency ?? 'EUR');

	const price = (value: number) =>
		scaledToNumber(value).toLocaleString('en', { style: 'currency', currency });

	async function selectListing(listing: Listing) {
		selected = listing;
		loading = true;
		error = null;
		try {
			rows = (await getEOD({ listing_id: listing.id, sort_order: 'desc' })).Data;
		} catch {
			error = 'Failed to load daily data';
		} finally {
			loading = false;
		}
	}
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Dailies"
			breadcrumb={[{ label: 'Admin', href: '/admin' }, { label: 'Dailies' }]}
			accountName="Admin User"
			adminMode={adminMode.enabled}
			onAdminToggle={(v) => adminMode.set(v)}
		/>
	{/snippet}

	<PageContentTemplate>
		<div class="flex min-h-0 flex-1 flex-col">
			<div class="border-b border-slate-200 p-3">
				<div class="w-full max-w-sm">
					<ListingSearchSelect
						value={selected}
						placeholder="Search a listing…"
						onSelect={selectListing}
					/>
				</div>
			</div>

			{#if !selected}
				<div class="flex flex-1 items-center justify-center p-8 text-sm text-slate-400">
					Search for a listing to view its daily OHLCV data.
				</div>
			{:else}
				<DataTable
					{rows}
					{loading}
					{error}
					getRowId={(r: EOD) => r.ID}
					emptyText="No daily data for this listing"
					columns={[
						{
							key: 'date',
							header: 'Date',
							value: (r: EOD) => formatDisplayDate(r.Date.slice(0, 10))
						},
						{ key: 'open', header: 'Open', align: 'right', value: (r: EOD) => price(r.Open) },
						{ key: 'high', header: 'High', align: 'right', value: (r: EOD) => price(r.High) },
						{ key: 'low', header: 'Low', align: 'right', value: (r: EOD) => price(r.Low) },
						{ key: 'close', header: 'Close', align: 'right', value: (r: EOD) => price(r.Close) },
						{
							key: 'volume',
							header: 'Volume',
							align: 'right',
							value: (r: EOD) => r.Volume.toLocaleString('en')
						}
					]}
				/>
			{/if}
		</div>
	</PageContentTemplate>
</AppShellTemplate>

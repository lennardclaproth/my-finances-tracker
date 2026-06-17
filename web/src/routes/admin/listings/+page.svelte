<script lang="ts">
	import { onMount } from 'svelte';
	import AppShellTemplate from '$lib/components/templates/app-shell/AppShellTemplate.svelte';
	import PageContentTemplate from '$lib/components/templates/page-content/PageContentTemplate.svelte';
	import TopNavbar from '$lib/components/organisms/top-navbar/TopNavbar.svelte';
	import DataTable from '$lib/components/organisms/data-table/DataTable.svelte';
	import Dialog from '$lib/components/molecules/dialog/Dialog.svelte';
	import FormField from '$lib/components/molecules/form-field/FormField.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { listListings } from '$lib/services/marketdata';
	import { toast } from '$lib/stores/toast.svelte';
	import { adminMode } from '$lib/stores/admin.svelte';
	import type { Listing } from '$lib/api/types';

	let listings = $state<Listing[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let createOpen = $state(false);
	let name = $state('');
	let symbol = $state('');
	let source = $state('');

	async function load() {
		loading = true;
		error = null;
		try {
			listings = await listListings();
		} catch {
			error = 'Failed to load listings';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function createListing() {
		if (name.trim() === '' || symbol.trim() === '') return;
		createOpen = false;
		toast.success(`Listing ${symbol.toUpperCase()} created`);
		name = '';
		symbol = '';
		source = '';
		void load();
	}
</script>

<AppShellTemplate>
	{#snippet top()}
		<TopNavbar
			title="Listings"
			breadcrumb={[{ label: 'Admin', href: '/admin' }, { label: 'Listings' }]}
			accountName="Admin User"
			adminMode={adminMode.enabled}
			onAdminToggle={(v) => adminMode.set(v)}
		/>
	{/snippet}

	<PageContentTemplate showFab fabLabel="New listing" onFabClick={() => (createOpen = true)}>
		<DataTable
			rows={listings}
			{loading}
			{error}
			emptyText="No listings"
			columns={[
				{ key: 'symbol', header: 'Symbol', value: (r: Listing) => r.symbol },
				{ key: 'name', header: 'Name', value: (r: Listing) => r.name },
				{ key: 'source', header: 'Source', value: (r: Listing) => r.source },
				{ key: 'exchange', header: 'Exchange', value: (r: Listing) => r.exchange ?? '—' },
				{ key: 'currency', header: 'Currency', value: (r: Listing) => r.currency ?? '—' },
				{ key: 'type', header: 'Type', value: (r: Listing) => r.type ?? '—' }
			]}
		/>
	</PageContentTemplate>
</AppShellTemplate>

<Dialog bind:open={createOpen} title="New listing" size="md">
	<div class="space-y-3">
		<FormField label="Symbol" id="listing-symbol">
			{#snippet children(ctx)}
				<Input id={ctx.id} bind:value={symbol} placeholder="e.g. VWRL" />
			{/snippet}
		</FormField>
		<FormField label="Name" id="listing-name">
			{#snippet children(ctx)}
				<Input id={ctx.id} bind:value={name} placeholder="e.g. Vanguard FTSE All-World" />
			{/snippet}
		</FormField>
		<FormField label="Source" id="listing-source" hint="Optional">
			{#snippet children(ctx)}
				<Input id={ctx.id} bind:value={source} placeholder="e.g. marketstack" />
			{/snippet}
		</FormField>
	</div>
	{#snippet footer()}
		<Button variant="ghost" intent="secondary" onclick={() => (createOpen = false)}>Cancel</Button>
		<Button onclick={createListing}>Create</Button>
	{/snippet}
</Dialog>

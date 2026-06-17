<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import TopNavbar from './TopNavbar.svelte';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	const actions: MenuItem[] = [
		{ label: 'New transaction', icon: 'heroicons:plus' },
		{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' }
	];
	const accountItems: MenuItem[] = [
		{ label: 'Account settings', icon: 'heroicons:cog-6-tooth' },
		{ label: 'Sign out', icon: 'heroicons:arrow-right-on-rectangle', intent: 'danger' }
	];

	const { Story } = defineMeta({
		title: 'Organisms/TopNavbar',
		component: TopNavbar,
		tags: ['autodocs']
	});
</script>

<Story name="Full (search + date + actions)" asChild>
	<TopNavbar
		title="Cashflow"
		showSearch
		showDateRange
		{actions}
		accountName="Lennard Claproth"
		accountEmail="lennard@example.com"
		{accountItems}
	/>
</Story>

<Story name="Admin (minimal)" asChild>
	<TopNavbar title="Listings" accountName="Admin User" adminMode={true} />
</Story>

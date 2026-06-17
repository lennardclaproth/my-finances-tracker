<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import AccountMenu from './AccountMenu.svelte';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	const items: MenuItem[] = [
		{ label: 'Account settings', icon: 'heroicons:cog-6-tooth' },
		{ label: 'Switch account', icon: 'heroicons:arrows-right-left' },
		{ label: 'Sign out', icon: 'heroicons:arrow-right-on-rectangle', intent: 'danger' }
	];

	const { Story } = defineMeta({
		title: 'Molecules/AccountMenu',
		component: AccountMenu,
		tags: ['autodocs'],
		args: { name: 'Lennard Claproth', email: 'lennard@example.com', items }
	});
</script>

<Story name="Playground" asChild>
	<div class="flex min-h-72 items-start justify-end">
		<AccountMenu name="Lennard Claproth" email="lennard@example.com" {items} />
	</div>
</Story>

<Story name="Admin enabled, no extra items" asChild>
	<div class="flex min-h-56 items-start justify-end">
		<AccountMenu name="Demo User" adminMode={true} />
	</div>
</Story>

<Story name="Without admin toggle" asChild>
	<div class="flex min-h-56 items-start justify-end">
		<AccountMenu name="Read Only" email="ro@example.com" showAdminToggle={false} {items} />
	</div>
</Story>

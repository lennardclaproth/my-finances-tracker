<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import NavMenu from './NavMenu.svelte';
	import type { NavItem } from './nav-menu.types';

	const items: NavItem[] = [
		{ label: 'Cashflow', href: '/cashflow', icon: 'heroicons:banknotes' },
		{ label: 'Assets', href: '/assets', icon: 'heroicons:building-library' },
		{ label: 'Portfolio', href: '/portfolio', icon: 'heroicons:chart-pie' },
		{ label: 'Listings', href: '/admin/listings', icon: 'heroicons:cog-6-tooth', divider: true }
	];

	const { Story } = defineMeta({
		title: 'Molecules/NavMenu',
		component: NavMenu,
		tags: ['autodocs'],
		args: { items, currentPath: '/cashflow' }
	});
</script>

<Story name="Playground" asChild>
	<div class="flex min-h-64 items-start justify-start">
		<NavMenu {items} currentPath="/cashflow" />
	</div>
</Story>

<Story name="Assets active" asChild>
	<div class="flex min-h-64 items-start justify-start">
		<NavMenu {items} currentPath="/assets" />
	</div>
</Story>

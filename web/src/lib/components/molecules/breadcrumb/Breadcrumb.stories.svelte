<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import type { ComponentProps } from 'svelte';
	import Breadcrumb from './Breadcrumb.svelte';
	import type { BreadcrumbItem } from './breadcrumb.types';

	type BreadcrumbProps = ComponentProps<typeof Breadcrumb>;

	const items: BreadcrumbItem[] = [
		{ label: 'Home', href: '/' },
		{ label: 'Admin', href: '/admin' },
		{ label: 'Listings', href: '/admin/listings' },
		{ label: 'VWRL' }
	];

	const { Story } = defineMeta({
		title: 'Molecules/Breadcrumb',
		component: Breadcrumb,
		tags: ['autodocs'],
		args: { items }
	});
</script>

{#snippet playground(args: BreadcrumbProps)}
	<Breadcrumb {...args} />
{/snippet}

<Story name="Playground" args={{ items }} template={playground} />

<Story name="Two levels" asChild>
	<Breadcrumb items={[{ label: 'Portfolio', href: '/portfolio' }, { label: 'Transactions' }]} />
</Story>

<Story name="Single (current page only)" asChild>
	<Breadcrumb items={[{ label: 'Cashflow' }]} />
</Story>

<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import type { ComponentProps } from 'svelte';
	import Tabs from './Tabs.svelte';
	import { tabsSizes, type TabItem } from './tabs.types';

	type TabsProps = ComponentProps<typeof Tabs>;

	const positions: TabItem[] = [
		{ value: 'positions', label: 'Positions' },
		{ value: 'transactions', label: 'Transactions' }
	];

	const withDisabled: TabItem[] = [
		{ value: 'overview', label: 'Overview' },
		{ value: 'analytics', label: 'Analytics' },
		{ value: 'archived', label: 'Archived', disabled: true }
	];

	const { Story } = defineMeta({
		title: 'Molecules/Tabs',
		component: Tabs,
		tags: ['autodocs'],
		args: { tabs: positions, size: 'md' },
		argTypes: {
			size: { control: 'select', options: tabsSizes }
		}
	});
</script>

{#snippet playground(args: TabsProps)}
	<Tabs {...args} />
{/snippet}

<Story
	name="Playground"
	args={{ tabs: positions, value: 'positions', size: 'md' }}
	template={playground}
/>

<Story name="Sizes" asChild>
	<div class="flex flex-col items-start gap-3">
		{#each tabsSizes as size (size)}
			<Tabs tabs={positions} value="positions" {size} ariaLabel={`Size ${size}`} />
		{/each}
	</div>
</Story>

<Story name="With disabled tab" asChild>
	<Tabs tabs={withDisabled} value="overview" ariaLabel="Sections" />
</Story>

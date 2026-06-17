<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ActionMenu from './ActionMenu.svelte';
	import type { MenuItem } from './menu.types';

	const items: MenuItem[] = [
		{ label: 'New transaction', icon: 'heroicons:plus' },
		{ label: 'Import CSV', icon: 'heroicons:cloud-arrow-up' },
		{ label: 'Rebuild portfolio', icon: 'heroicons:arrow-path' },
		{ label: 'Delete account', icon: 'heroicons:trash', intent: 'danger', divider: true }
	];

	const { Story } = defineMeta({
		title: 'Molecules/ActionMenu',
		component: ActionMenu,
		tags: ['autodocs'],
		args: { items }
	});
</script>

<Story name="Playground" asChild>
	<div class="flex min-h-64 items-start justify-center">
		<ActionMenu {items} />
	</div>
</Story>

<Story name="With disabled item" asChild>
	<div class="flex min-h-64 items-start justify-center">
		<ActionMenu
			label="Manage"
			items={[
				{ label: 'Edit', icon: 'heroicons:pencil-square' },
				{ label: 'Duplicate', icon: 'heroicons:document-duplicate', disabled: true },
				{ label: 'Remove', icon: 'heroicons:trash', intent: 'danger', divider: true }
			]}
		/>
	</div>
</Story>

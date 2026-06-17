<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import VisibilityFilter from './VisibilityFilter.svelte';

	const columns = [
		{ key: 'date', label: 'Date' },
		{ key: 'description', label: 'Description' },
		{ key: 'note', label: 'Note' },
		{ key: 'tag', label: 'Tag' },
		{ key: 'amount', label: 'Amount' }
	];

	const { Story } = defineMeta({
		title: 'Molecules/Filters/VisibilityFilter',
		component: VisibilityFilter,
		tags: ['autodocs'],
		args: { columns }
	});
</script>

<Story name="All visible" asChild>
	<div class="flex min-h-64 items-start">
		<VisibilityFilter {columns} />
	</div>
</Story>

<Story name="Some hidden" asChild>
	<div class="flex min-h-64 items-start">
		<VisibilityFilter {columns} hidden={['note', 'tag']} />
	</div>
</Story>

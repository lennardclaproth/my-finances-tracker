<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import DirectionFilter from './DirectionFilter.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/Filters/DirectionFilter',
		component: DirectionFilter,
		tags: ['autodocs']
	});
</script>

<Story name="Inactive (All)" asChild>
	<div class="flex min-h-48 items-start">
		<DirectionFilter />
	</div>
</Story>

<Story name="Active (Incoming)" asChild>
	<div class="flex min-h-48 items-start">
		<DirectionFilter value="in" />
	</div>
</Story>

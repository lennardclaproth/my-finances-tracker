<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import TextFilter from './TextFilter.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/Filters/TextFilter',
		component: TextFilter,
		tags: ['autodocs']
	});
</script>

<Story name="Inactive" asChild>
	<div class="flex min-h-56 items-start">
		<TextFilter label="Description" placeholder="Search description…" />
	</div>
</Story>

<Story name="Active" asChild>
	<div class="flex min-h-56 items-start">
		<TextFilter label="Description" value="Albert Heijn" />
	</div>
</Story>

<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import SelectFilter from './SelectFilter.svelte';

	const tagOptions = [
		{ value: 'salary', label: 'Salary' },
		{ value: 'rent', label: 'Rent' },
		{ value: 'groceries', label: 'Groceries' },
		{ value: 'dining', label: 'Dining' },
		{ value: 'utilities', label: 'Utilities' },
		{ value: 'transport', label: 'Transport' }
	];

	const { Story } = defineMeta({
		title: 'Molecules/Filters/SelectFilter',
		component: SelectFilter,
		tags: ['autodocs'],
		args: { options: tagOptions, label: 'Tags' }
	});
</script>

<Story name="Inactive" asChild>
	<div class="flex min-h-72 items-start">
		<SelectFilter options={tagOptions} label="Tags" />
	</div>
</Story>

<Story name="Active" asChild>
	<div class="flex min-h-72 items-start">
		<SelectFilter options={tagOptions} label="Tags" selected={['rent', 'groceries']} />
	</div>
</Story>

<Story name="Empty options" asChild>
	<div class="flex min-h-48 items-start">
		<SelectFilter options={[]} label="Tags" />
	</div>
</Story>

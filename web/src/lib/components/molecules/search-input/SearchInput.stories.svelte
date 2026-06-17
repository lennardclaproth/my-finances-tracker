<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import type { ComponentProps } from 'svelte';
	import SearchInput from './SearchInput.svelte';
	import { inputSizes, inputShapes } from '$lib/components/atoms/input/input.types';

	type SearchInputProps = ComponentProps<typeof SearchInput>;

	const { Story } = defineMeta({
		title: 'Molecules/SearchInput',
		component: SearchInput,
		tags: ['autodocs'],
		argTypes: {
			size: { control: 'select', options: inputSizes },
			shape: { control: 'select', options: inputShapes },
			debounceMs: { control: 'number' },
			disabled: { control: 'boolean' }
		}
	});
</script>

<script lang="ts">
	let lastSearch = $state('');
	let value = $state('groceries');
</script>

{#snippet playground(args: SearchInputProps)}
	<div class="w-80">
		<SearchInput {...args} />
	</div>
{/snippet}

<Story
	name="Playground"
	args={{ placeholder: 'Search transactions…', size: 'md', shape: 'rounded', debounceMs: 300 }}
	template={playground}
/>

<Story name="Debounced search" asChild>
	<div class="w-80 space-y-2">
		<SearchInput bind:value placeholder="Type to search…" onSearch={(q) => (lastSearch = q)} />
		<p class="text-xs text-slate-500">
			Debounced query: <span class="font-mono text-slate-800">{lastSearch || '—'}</span>
		</p>
	</div>
</Story>

<Story name="Sizes" asChild>
	<div class="flex w-80 flex-col gap-3">
		{#each inputSizes as size (size)}
			<SearchInput {size} value="VWRL" ariaLabel={`Search ${size}`} />
		{/each}
	</div>
</Story>

<Story name="Disabled" asChild>
	<div class="w-80">
		<SearchInput value="locked" disabled />
	</div>
</Story>

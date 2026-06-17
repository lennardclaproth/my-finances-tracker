<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import type { ComponentProps } from 'svelte';
	import DatePicker from './DatePicker.svelte';

	type DatePickerProps = ComponentProps<typeof DatePicker>;

	const { Story } = defineMeta({
		title: 'Molecules/DatePicker',
		component: DatePicker,
		tags: ['autodocs'],
		argTypes: {
			size: { control: 'select', options: ['sm', 'md', 'lg'] },
			disabled: { control: 'boolean' }
		}
	});
</script>

<script lang="ts">
	let value = $state<string | null>('2026-06-12');
</script>

{#snippet playground(args: DatePickerProps)}
	<div class="flex min-h-72 items-start">
		<DatePicker {...args} />
	</div>
{/snippet}

<Story name="Playground" args={{ size: 'md', placeholder: 'Select date' }} template={playground} />

<Story name="With value" asChild>
	<div class="flex min-h-72 items-start gap-3">
		<DatePicker bind:value />
		<span class="text-sm text-slate-500">value: <code>{value ?? 'null'}</code></span>
	</div>
</Story>

<Story name="Min / max bounded" asChild>
	<div class="flex min-h-72 items-start">
		<DatePicker value="2026-06-15" min="2026-06-05" max="2026-06-24" />
	</div>
</Story>

<Story name="Empty + disabled" asChild>
	<div class="flex items-start gap-3">
		<DatePicker placeholder="Pick a day" />
		<DatePicker value="2026-06-12" disabled />
	</div>
</Story>

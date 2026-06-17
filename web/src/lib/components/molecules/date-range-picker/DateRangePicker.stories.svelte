<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import DateRangePicker from './DateRangePicker.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/DateRangePicker',
		component: DateRangePicker,
		tags: ['autodocs'],
		argTypes: {
			size: { control: 'select', options: ['sm', 'md', 'lg'] },
			showPresets: { control: 'boolean' },
			disabled: { control: 'boolean' }
		}
	});
</script>

<script lang="ts">
	let from = $state<string | null>('2026-06-09');
	let to = $state<string | null>('2026-06-20');
</script>

<Story name="With range" asChild>
	<div class="flex min-h-96 items-start gap-3">
		<DateRangePicker bind:from bind:to />
		<span class="text-sm text-slate-500">
			<code>{from ?? 'null'}</code> → <code>{to ?? 'null'}</code>
		</span>
	</div>
</Story>

<Story name="Empty (presets only)" asChild>
	<div class="flex min-h-96 items-start">
		<DateRangePicker placeholder="Pick a range" />
	</div>
</Story>

<Story name="No presets" asChild>
	<div class="flex min-h-96 items-start">
		<DateRangePicker from="2026-05-01" to="2026-05-15" showPresets={false} />
	</div>
</Story>

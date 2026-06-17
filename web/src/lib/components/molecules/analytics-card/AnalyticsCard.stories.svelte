<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import AnalyticsCard from './AnalyticsCard.svelte';

	const { Story } = defineMeta({
		title: 'Molecules/AnalyticsCard',
		component: AnalyticsCard,
		tags: ['autodocs'],
		argTypes: {
			title: { control: 'text' }
		}
	});
</script>

<Story name="Playground" asChild>
	<div class="max-w-sm">
		<AnalyticsCard title="Net trend">
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">
				Chart goes here
			</div>
		</AnalyticsCard>
	</div>
</Story>

<Story name="Without title" asChild>
	<div class="max-w-sm">
		<AnalyticsCard>
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">Untitled card</div>
		</AnalyticsCard>
	</div>
</Story>

<Story name="In a grid" asChild>
	<div class="grid grid-cols-1 gap-3 lg:grid-cols-4">
		<AnalyticsCard title="Net trend" class="lg:col-span-2">
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">Wide chart</div>
		</AnalyticsCard>
		<AnalyticsCard title="Incoming">
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">Donut</div>
		</AnalyticsCard>
		<AnalyticsCard title="Outgoing">
			<div class="flex h-32 items-center justify-center text-sm text-slate-400">Donut</div>
		</AnalyticsCard>
	</div>
</Story>

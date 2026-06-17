<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Calendar from './Calendar.svelte';
	import { startOfMonthUTC } from './calendar.utils';

	const june2026 = startOfMonthUTC(new Date(Date.UTC(2026, 5, 1)));

	const { Story } = defineMeta({
		title: 'Molecules/Calendar',
		component: Calendar,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	let single = $state('2026-06-12');
	let rangeStart = $state<string | null>('2026-06-09');
	let rangeEnd = $state<string | null>('2026-06-20');
</script>

<Story name="Single select" asChild>
	<div class="inline-block rounded-xl border border-slate-300 bg-white p-3">
		<Calendar month={june2026} mode="single" selected={single} onSelect={(iso) => (single = iso)} />
	</div>
</Story>

<Story name="Range with selection" asChild>
	<div class="inline-block rounded-xl border border-slate-300 bg-white p-3">
		<Calendar month={june2026} mode="range" {rangeStart} {rangeEnd} />
	</div>
</Story>

<Story name="Min / max disabled" asChild>
	<div class="inline-block rounded-xl border border-slate-300 bg-white p-3">
		<Calendar
			month={june2026}
			mode="single"
			selected="2026-06-15"
			min="2026-06-05"
			max="2026-06-24"
		/>
	</div>
</Story>

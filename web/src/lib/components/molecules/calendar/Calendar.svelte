<script lang="ts">
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import type { CalendarMode } from './calendar.types';
	import {
		buildMonthGrid,
		addMonthsUTC,
		isDisabled,
		isWithinRange,
		monthLabel,
		startOfMonthUTC,
		weekdayLabels
	} from './calendar.utils';

	type Props = {
		/** Displayed month (bindable so the built-in nav can advance it). */
		month?: Date;
		mode?: CalendarMode;
		/** Selected day in single mode ("YYYY-MM-DD"). */
		selected?: string | null;
		/** Range endpoints in range mode. */
		rangeStart?: string | null;
		rangeEnd?: string | null;
		/** Hovered day used to preview the pending range half. */
		hoverDate?: string | null;
		min?: string | null;
		max?: string | null;
		/** Show the month-navigation header. */
		showNav?: boolean;
		onSelect?: (iso: string) => void;
		onHover?: (iso: string | null) => void;
		onMonthChange?: (month: Date) => void;
		class?: string;
	};

	let {
		month = $bindable(startOfMonthUTC(new Date())),
		mode = 'single',
		selected = null,
		rangeStart = null,
		rangeEnd = null,
		hoverDate = null,
		min = null,
		max = null,
		showNav = true,
		onSelect,
		onHover,
		onMonthChange,
		class: className = ''
	}: Props = $props();

	const grid = $derived(buildMonthGrid(month));

	function shiftMonth(delta: number) {
		month = addMonthsUTC(month, delta);
		onMonthChange?.(month);
	}

	function cellClasses(iso: string, inMonth: boolean, isToday: boolean): string {
		const disabled = isDisabled(iso, min, max);
		const isStart = mode === 'range' && rangeStart === iso;
		const isEnd = mode === 'range' && rangeEnd === iso;
		const isSelected = mode === 'single' ? selected === iso : isStart || isEnd;
		const inRange =
			mode === 'range' && !isSelected && isWithinRange(iso, rangeStart, rangeEnd ?? hoverDate);

		const classes = [
			'flex h-9 w-9 items-center justify-center rounded-lg text-sm transition-colors'
		];
		if (disabled) {
			classes.push('pointer-events-none text-slate-300 opacity-40');
		} else if (isSelected) {
			classes.push('bg-slate-600 font-medium text-amber-200');
		} else if (inRange) {
			classes.push('bg-slate-100 text-slate-800');
		} else if (!inMonth) {
			classes.push('text-slate-300 hover:bg-slate-50');
		} else {
			classes.push('text-slate-700 hover:bg-slate-100');
		}
		if (isToday && !isSelected) classes.push('ring-1 ring-slate-300');
		return classes.join(' ');
	}
</script>

<div class={['w-fit select-none', className].filter(Boolean).join(' ')}>
	{#if showNav}
		<div class="mb-2 flex items-center justify-between">
			<button
				type="button"
				aria-label="Previous month"
				class="inline-flex size-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
				onclick={() => shiftMonth(-1)}
			>
				<Icon icon="heroicons:chevron-left" size="sm" />
			</button>
			<span class="text-sm font-medium text-slate-800">{monthLabel(month)}</span>
			<button
				type="button"
				aria-label="Next month"
				class="inline-flex size-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
				onclick={() => shiftMonth(1)}
			>
				<Icon icon="heroicons:chevron-right" size="sm" />
			</button>
		</div>
	{/if}

	<div class="grid grid-cols-7 gap-0.5">
		{#each weekdayLabels as label (label)}
			<div class="flex h-7 w-9 items-center justify-center text-[11px] font-medium text-slate-500">
				{label}
			</div>
		{/each}

		{#each grid as cell (cell.iso)}
			{@const disabled = isDisabled(cell.iso, min, max)}
			<button
				type="button"
				class={cellClasses(cell.iso, cell.inMonth, cell.isToday)}
				{disabled}
				aria-label={cell.iso}
				aria-current={cell.isToday ? 'date' : undefined}
				aria-pressed={mode === 'single' ? selected === cell.iso : undefined}
				onclick={() => onSelect?.(cell.iso)}
				onmouseenter={() => onHover?.(cell.iso)}
				onmouseleave={() => onHover?.(null)}
			>
				{cell.day}
			</button>
		{/each}
	</div>
</div>

<script lang="ts">
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Popover from '$lib/components/molecules/popover/Popover.svelte';
	import Calendar from '$lib/components/molecules/calendar/Calendar.svelte';
	import {
		addDaysISO,
		addMonthsUTC,
		formatDisplayDate,
		monthLabel,
		parseISODate,
		startOfMonthUTC,
		toISODate,
		todayISO
	} from '$lib/components/molecules/calendar/calendar.utils';

	type Size = 'sm' | 'md' | 'lg';

	type Props = {
		/** Bindable range endpoints ("YYYY-MM-DD") or null. */
		from?: string | null;
		to?: string | null;
		placeholder?: string;
		min?: string | null;
		max?: string | null;
		size?: Size;
		disabled?: boolean;
		/** Show the quick-range preset column. */
		showPresets?: boolean;
		ariaLabel?: string;
		onChange?: (range: { from: string | null; to: string | null }) => void;
		class?: string;
	};

	let {
		from = $bindable(null),
		to = $bindable(null),
		placeholder = 'Select range',
		min = null,
		max = null,
		size = 'md',
		disabled = false,
		showPresets = true,
		ariaLabel = 'Select date range',
		onChange,
		class: className = ''
	}: Props = $props();

	let open = $state(false);
	let leftMonth = $state(startOfMonthUTC(parseISODate(from) ?? new Date()));
	let draftStart = $state<string | null>(from);
	let draftEnd = $state<string | null>(to);
	let hover = $state<string | null>(null);

	const rightMonth = $derived(addMonthsUTC(leftMonth, 1));

	// Reset the draft to the committed range each time the popover opens.
	$effect(() => {
		if (open) {
			draftStart = from;
			draftEnd = to;
			hover = null;
			leftMonth = startOfMonthUTC(parseISODate(from) ?? new Date());
		}
	});

	function handleSelect(iso: string) {
		if (!draftStart || (draftStart && draftEnd)) {
			draftStart = iso;
			draftEnd = null;
			return;
		}
		// Second pick: order the endpoints.
		if (iso < draftStart) {
			draftEnd = draftStart;
			draftStart = iso;
		} else {
			draftEnd = iso;
		}
	}

	function shift(delta: number) {
		leftMonth = addMonthsUTC(leftMonth, delta);
	}

	function apply() {
		const start = draftStart;
		const end = draftEnd ?? draftStart;
		from = start;
		to = end;
		onChange?.({ from, to });
		open = false;
	}

	function clear() {
		draftStart = null;
		draftEnd = null;
		from = null;
		to = null;
		onChange?.({ from: null, to: null });
		open = false;
	}

	const sizeClasses = {
		sm: 'h-8 px-2.5 text-sm',
		md: 'h-10 px-3 text-sm',
		lg: 'h-12 px-3 text-base'
	} satisfies Record<Size, string>;

	const label = $derived(
		from && to
			? `${formatDisplayDate(from)} – ${formatDisplayDate(to)}`
			: from
				? formatDisplayDate(from)
				: placeholder
	);

	const presets = $derived.by(() => {
		const today = todayISO();
		const monthStart = toISODate(startOfMonthUTC(parseISODate(today) ?? new Date()));
		const yearStart = `${today.slice(0, 4)}-01-01`;
		return [
			{ label: 'Last 7 days', from: addDaysISO(today, -6), to: today },
			{ label: 'Last 30 days', from: addDaysISO(today, -29), to: today },
			{ label: 'This month', from: monthStart, to: today },
			{
				label: 'Last 3 months',
				from: toISODate(addMonthsUTC(parseISODate(today) ?? new Date(), -3)),
				to: today
			},
			{ label: 'Year to date', from: yearStart, to: today }
		];
	});

	function applyPreset(preset: { from: string; to: string }) {
		draftStart = preset.from;
		draftEnd = preset.to;
		leftMonth = startOfMonthUTC(parseISODate(preset.from) ?? new Date());
	}
</script>

<Popover bind:open placement="bottom-start" class="p-3">
	{#snippet trigger(api)}
		<button
			type="button"
			{disabled}
			aria-label={ariaLabel}
			aria-expanded={api.open}
			class={[
				'inline-flex items-center gap-2 rounded-xl border bg-white whitespace-nowrap',
				'transition-colors hover:bg-slate-50',
				'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
				'disabled:pointer-events-none disabled:opacity-50',
				api.open ? 'border-slate-400' : 'border-slate-300',
				from ? 'text-slate-800' : 'text-slate-500',
				sizeClasses[size],
				className
			].join(' ')}
			onclick={api.toggle}
		>
			<Icon icon="heroicons:calendar-days" size="sm" class="text-slate-500" />
			<span>{label}</span>
		</button>
	{/snippet}

	<div class="flex gap-3">
		{#if showPresets}
			<div class="flex w-36 flex-col gap-1 border-r border-slate-200 pr-3">
				{#each presets as preset (preset.label)}
					<button
						type="button"
						class="rounded-lg px-2 py-1.5 text-left text-sm text-slate-600 transition-colors hover:bg-slate-100 hover:text-slate-900 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
						onclick={() => applyPreset(preset)}
					>
						{preset.label}
					</button>
				{/each}
			</div>
		{/if}

		<div class="flex flex-col gap-3">
			<div class="flex items-center justify-between">
				<button
					type="button"
					aria-label="Previous month"
					class="inline-flex size-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
					onclick={() => shift(-1)}
				>
					<Icon icon="heroicons:chevron-left" size="sm" />
				</button>
				<div class="flex flex-1 justify-around text-sm font-medium text-slate-800">
					<span>{monthLabel(leftMonth)}</span>
					<span>{monthLabel(rightMonth)}</span>
				</div>
				<button
					type="button"
					aria-label="Next month"
					class="inline-flex size-8 items-center justify-center rounded-lg text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
					onclick={() => shift(1)}
				>
					<Icon icon="heroicons:chevron-right" size="sm" />
				</button>
			</div>

			<div class="flex gap-4">
				<Calendar
					month={leftMonth}
					mode="range"
					rangeStart={draftStart}
					rangeEnd={draftEnd}
					hoverDate={hover}
					{min}
					{max}
					showNav={false}
					onSelect={handleSelect}
					onHover={(iso) => (hover = iso)}
				/>
				<Calendar
					month={rightMonth}
					mode="range"
					rangeStart={draftStart}
					rangeEnd={draftEnd}
					hoverDate={hover}
					{min}
					{max}
					showNav={false}
					onSelect={handleSelect}
					onHover={(iso) => (hover = iso)}
				/>
			</div>

			<div class="flex items-center justify-between border-t border-slate-200 pt-3">
				<span class="text-xs text-slate-500">
					{draftStart ? formatDisplayDate(draftStart) : '—'}
					{draftEnd ? `→ ${formatDisplayDate(draftEnd)}` : ''}
				</span>
				<div class="flex gap-2">
					<Button size="sm" variant="ghost" intent="secondary" onclick={clear}>Clear</Button>
					<Button size="sm" onclick={apply} disabled={!draftStart}>Apply</Button>
				</div>
			</div>
		</div>
	</div>
</Popover>

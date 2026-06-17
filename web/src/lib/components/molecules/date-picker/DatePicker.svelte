<script lang="ts">
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Popover from '$lib/components/molecules/popover/Popover.svelte';
	import Calendar from '$lib/components/molecules/calendar/Calendar.svelte';
	import {
		formatDisplayDate,
		parseISODate,
		startOfMonthUTC
	} from '$lib/components/molecules/calendar/calendar.utils';

	type Size = 'sm' | 'md' | 'lg';

	type Props = {
		/** Bindable selected date ("YYYY-MM-DD") or null. */
		value?: string | null;
		placeholder?: string;
		min?: string | null;
		max?: string | null;
		size?: Size;
		disabled?: boolean;
		ariaLabel?: string;
		onChange?: (value: string | null) => void;
		class?: string;
	};

	let {
		value = $bindable(null),
		placeholder = 'Select date',
		min = null,
		max = null,
		size = 'md',
		disabled = false,
		ariaLabel = 'Select date',
		onChange,
		class: className = ''
	}: Props = $props();

	let open = $state(false);
	let month = $state(startOfMonthUTC(parseISODate(value) ?? new Date()));

	// Re-center the calendar on the selected value whenever the popover opens.
	$effect(() => {
		if (open) month = startOfMonthUTC(parseISODate(value) ?? new Date());
	});

	const sizeClasses = {
		sm: 'h-8 px-2.5 text-sm',
		md: 'h-10 px-3 text-sm',
		lg: 'h-12 px-3 text-base'
	} satisfies Record<Size, string>;

	const label = $derived(value ? formatDisplayDate(value) : placeholder);

	function handleSelect(iso: string) {
		value = iso;
		onChange?.(iso);
		open = false;
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
				value ? 'text-slate-800' : 'text-slate-400',
				sizeClasses[size],
				className
			].join(' ')}
			onclick={api.toggle}
		>
			<Icon icon="heroicons:calendar-days" size="sm" class="text-slate-400" />
			<span>{label}</span>
		</button>
	{/snippet}

	<Calendar bind:month mode="single" selected={value} {min} {max} onSelect={handleSelect} />
</Popover>

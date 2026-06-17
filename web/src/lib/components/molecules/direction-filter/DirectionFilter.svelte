<script lang="ts">
	import FilterPopover from '$lib/components/molecules/filter-popover/FilterPopover.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';

	type Direction = 'in' | 'out';

	type Props = {
		/** Bindable selected direction; `null` means "all" (inactive). */
		value?: Direction | null;
		label?: string;
		onApply?: (value: Direction | null) => void;
	};

	let { value = $bindable(null), label = 'Direction', onApply }: Props = $props();

	let open = $state(false);

	const active = $derived(value !== null);

	const options: { value: Direction | null; label: string }[] = [
		{ value: null, label: 'All' },
		{ value: 'in', label: 'Incoming' },
		{ value: 'out', label: 'Outgoing' }
	];

	function select(next: Direction | null) {
		value = next;
		onApply?.(next);
		open = false;
	}
</script>

<FilterPopover
	bind:open
	{active}
	label={`Filter by ${label}`}
	icon="heroicons:arrows-up-down"
	showFooter={false}
>
	<div class="flex w-44 flex-col gap-0.5">
		{#each options as option (option.label)}
			<button
				type="button"
				class={[
					'flex items-center justify-between rounded-md px-2 py-1.5 text-left text-sm transition-colors',
					value === option.value
						? 'bg-slate-100 font-medium text-slate-900'
						: 'text-slate-700 hover:bg-slate-50'
				].join(' ')}
				onclick={() => select(option.value)}
			>
				{option.label}
				{#if value === option.value}
					<Icon icon="heroicons:check" size="sm" class="text-slate-600" />
				{/if}
			</button>
		{/each}
	</div>
</FilterPopover>

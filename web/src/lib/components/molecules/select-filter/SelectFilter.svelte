<script lang="ts">
	import FilterPopover from '$lib/components/molecules/filter-popover/FilterPopover.svelte';
	import Checkbox from '$lib/components/atoms/checkbox/Checkbox.svelte';

	type Option = { value: string; label: string };

	type Props = {
		options: Option[];
		/** Bindable selected values (empty = inactive). */
		selected?: string[];
		label?: string;
		onApply?: (selected: string[]) => void;
	};

	let { options, selected = $bindable([]), label = 'Options', onApply }: Props = $props();

	let open = $state(false);
	let draft = $state<string[]>([...selected]);

	$effect(() => {
		if (open) draft = [...selected];
	});

	const active = $derived(selected.length > 0);

	function toggle(value: string) {
		draft = draft.includes(value) ? draft.filter((v) => v !== value) : [...draft, value];
	}

	function apply() {
		selected = [...draft];
		onApply?.(selected);
	}

	function clear() {
		draft = [];
		selected = [];
		onApply?.([]);
	}
</script>

<FilterPopover
	bind:open
	{active}
	count={selected.length}
	label={`Filter by ${label}`}
	onApply={apply}
	onClear={clear}
>
	<div class="max-h-56 w-52 space-y-0.5 overflow-y-auto pr-1">
		{#if options.length === 0}
			<p class="px-1 py-2 text-sm text-slate-500">No options</p>
		{/if}
		{#each options as option (option.value)}
			<label
				class="flex cursor-pointer items-center gap-2 rounded-md px-1 py-1 text-sm text-slate-700 hover:bg-slate-50"
			>
				<Checkbox
					size="sm"
					checked={draft.includes(option.value)}
					onchange={() => toggle(option.value)}
				/>
				<span>{option.label}</span>
			</label>
		{/each}
	</div>
</FilterPopover>

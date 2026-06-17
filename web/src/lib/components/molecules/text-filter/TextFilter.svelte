<script lang="ts">
	import FilterPopover from '$lib/components/molecules/filter-popover/FilterPopover.svelte';
	import Input from '$lib/components/atoms/input/Input.svelte';

	type Props = {
		/** Bindable committed filter value (empty string = inactive). */
		value?: string;
		label?: string;
		placeholder?: string;
		onApply?: (value: string) => void;
	};

	let {
		value = $bindable(''),
		label = 'Value',
		placeholder = 'Contains…',
		onApply
	}: Props = $props();

	let open = $state(false);
	let draft = $state(value);

	// Sync the draft to the committed value each time the popover opens.
	$effect(() => {
		if (open) draft = value;
	});

	const active = $derived(value.trim().length > 0);

	function apply() {
		value = draft.trim();
		onApply?.(value);
	}

	function clear() {
		draft = '';
		value = '';
		onApply?.('');
	}
</script>

<FilterPopover bind:open {active} label={`Filter by ${label}`} onApply={apply} onClear={clear}>
	<div class="w-56 space-y-1.5">
		<span class="block text-xs font-medium text-slate-500">{label}</span>
		<Input bind:value={draft} {placeholder} size="sm" ariaLabel={label} />
	</div>
</FilterPopover>

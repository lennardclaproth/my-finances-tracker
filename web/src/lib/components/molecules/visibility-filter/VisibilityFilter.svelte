<script lang="ts">
	import FilterPopover from '$lib/components/molecules/filter-popover/FilterPopover.svelte';
	import Switch from '$lib/components/atoms/switch/Switch.svelte';

	type Column = { key: string; label: string };

	type Props = {
		columns: Column[];
		/** Bindable list of hidden column keys (non-empty = active). */
		hidden?: string[];
		label?: string;
		onChange?: (hidden: string[]) => void;
	};

	let { columns, hidden = $bindable([]), label = 'Columns', onChange }: Props = $props();

	let open = $state(false);

	const active = $derived(hidden.length > 0);

	function toggle(key: string) {
		hidden = hidden.includes(key) ? hidden.filter((k) => k !== key) : [...hidden, key];
		onChange?.(hidden);
	}
</script>

<FilterPopover
	bind:open
	{active}
	count={hidden.length}
	{label}
	icon="heroicons:eye"
	showFooter={false}
>
	<div class="w-52 space-y-0.5">
		{#each columns as column (column.key)}
			<div class="flex items-center justify-between gap-3 px-1 py-1 text-sm text-slate-700">
				<span>{column.label}</span>
				<Switch checked={!hidden.includes(column.key)} onchange={() => toggle(column.key)} />
			</div>
		{/each}
	</div>
</FilterPopover>

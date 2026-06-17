<script lang="ts">
	import type { Snippet } from 'svelte';
	import Select from '$lib/components/atoms/select/Select.svelte';
	import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';

	type Props = {
		total: number;
		/** Bindable page size. */
		limit?: number;
		/** Bindable row offset. */
		offset?: number;
		pageSizes?: number[];
		/** When > 0, the bar shows selection + bulk actions instead of the range. */
		selectedCount?: number;
		onPageChange?: (offset: number) => void;
		onLimitChange?: (limit: number) => void;
		/** Bulk actions, shown when selectedCount > 0. */
		actions?: Snippet;
		class?: string;
	};

	let {
		total,
		limit = $bindable(25),
		offset = $bindable(0),
		pageSizes = [10, 25, 50, 100],
		selectedCount = 0,
		onPageChange,
		onLimitChange,
		actions,
		class: className = ''
	}: Props = $props();

	const from = $derived(total === 0 ? 0 : offset + 1);
	const to = $derived(Math.min(offset + limit, total));
	const canPrev = $derived(offset > 0);
	const canNext = $derived(offset + limit < total);
	const sizeOptions = $derived(pageSizes.map((n) => ({ value: String(n), label: String(n) })));

	function prev() {
		if (!canPrev) return;
		offset = Math.max(0, offset - limit);
		onPageChange?.(offset);
	}

	function next() {
		if (!canNext) return;
		offset = offset + limit;
		onPageChange?.(offset);
	}

	function changeLimit(event: Event) {
		const value = Number((event.target as HTMLSelectElement).value);
		limit = value;
		offset = 0;
		onLimitChange?.(value);
	}
</script>

<div
	class={[
		'flex items-center justify-between gap-3 bg-white px-3 py-2 text-sm text-slate-600',
		className
	]
		.filter(Boolean)
		.join(' ')}
>
	<div class="flex min-w-0 items-center gap-3">
		{#if selectedCount > 0}
			<span class="font-medium text-slate-800">{selectedCount} selected</span>
			{#if actions}
				{@render actions()}
			{/if}
		{:else}
			<span class="text-slate-500">{from}–{to} of {total}</span>
		{/if}
	</div>

	<div class="flex shrink-0 items-center gap-2">
		<span class="hidden text-slate-500 sm:inline">Rows</span>
		<div class="w-20">
			<Select
				value={String(limit)}
				options={sizeOptions}
				size="sm"
				ariaLabel="Rows per page"
				onchange={changeLimit}
			/>
		</div>
		<IconButton
			icon="heroicons:chevron-left"
			ariaLabel="Previous page"
			size="sm"
			variant="outline"
			disabled={!canPrev}
			onclick={prev}
		/>
		<IconButton
			icon="heroicons:chevron-right"
			ariaLabel="Next page"
			size="sm"
			variant="outline"
			disabled={!canNext}
			onclick={next}
		/>
	</div>
</div>

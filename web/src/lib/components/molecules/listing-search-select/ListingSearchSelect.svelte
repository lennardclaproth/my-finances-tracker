<script lang="ts">
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { zClasses } from '$lib/styles/z-index';
	import { searchListings } from '$lib/services/marketdata';
	import type { Listing } from '$lib/api/types';

	type Props = {
		/** Bindable selected listing. */
		value?: Listing | null;
		/** Bindable query text. */
		query?: string;
		placeholder?: string;
		/** Minimum characters before searching. */
		minChars?: number;
		debounceMs?: number;
		limit?: number;
		disabled?: boolean;
		/** Injectable search fn (defaults to the marketdata service, so it works on mocks or live API). */
		search?: (query: string) => Promise<Listing[]>;
		onSelect?: (listing: Listing) => void;
		ariaLabel?: string;
		class?: string;
	};

	let {
		value = $bindable(null),
		query = $bindable(''),
		placeholder = 'Search listings…',
		minChars = 2,
		debounceMs = 300,
		limit = 8,
		disabled = false,
		search,
		onSelect,
		ariaLabel = 'Search listings',
		class: className = ''
	}: Props = $props();

	const listboxId = $props.id();

	const defaultSearch = async (q: string): Promise<Listing[]> =>
		(await searchListings({ q, limit })).data;
	const runSearch = $derived(search ?? defaultSearch);

	let open = $state(false);
	let loading = $state(false);
	let error = $state(false);
	let results = $state<Listing[]>([]);
	let activeIndex = $state(-1);
	let inputEl = $state<HTMLInputElement | null>(null);

	let requestId = 0;
	let suppress = false;

	// Debounced, race-safe search driven by the query text.
	$effect(() => {
		const q = query.trim();
		if (suppress) {
			suppress = false;
			return;
		}
		if (q.length < minChars) {
			results = [];
			open = false;
			loading = false;
			error = false;
			return;
		}
		loading = true;
		error = false;
		open = true;
		const id = ++requestId;
		const timer = setTimeout(async () => {
			try {
				const found = await runSearch(q);
				if (id !== requestId) return;
				results = found;
				activeIndex = found.length ? 0 : -1;
			} catch {
				if (id !== requestId) return;
				error = true;
				results = [];
			} finally {
				if (id === requestId) loading = false;
			}
		}, debounceMs);
		return () => clearTimeout(timer);
	});

	function select(listing: Listing) {
		suppress = true;
		value = listing;
		query = listing.symbol;
		results = [];
		open = false;
		activeIndex = -1;
		onSelect?.(listing);
	}

	function clear() {
		value = null;
		query = '';
		results = [];
		open = false;
		inputEl?.focus();
	}

	function onInput() {
		// Typing invalidates any prior selection.
		value = null;
	}

	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'ArrowDown') {
			event.preventDefault();
			if (results.length) activeIndex = Math.min(activeIndex + 1, results.length - 1);
		} else if (event.key === 'ArrowUp') {
			event.preventDefault();
			if (results.length) activeIndex = Math.max(activeIndex - 1, 0);
		} else if (event.key === 'Enter') {
			if (open && activeIndex >= 0 && results[activeIndex]) {
				event.preventDefault();
				select(results[activeIndex]);
			}
		} else if (event.key === 'Escape') {
			open = false;
		}
	}

	function onFocus() {
		if (results.length > 0 || error) open = true;
	}
</script>

<div class={['relative', className].filter(Boolean).join(' ')}>
	<span
		class="pointer-events-none absolute inset-y-0 left-0 flex w-9 items-center justify-center text-slate-400"
		aria-hidden="true"
	>
		{#if loading}
			<Icon icon="heroicons:arrow-path" size="sm" class="animate-spin" />
		{:else}
			<Icon icon="heroicons:magnifying-glass" size="sm" />
		{/if}
	</span>

	<input
		bind:this={inputEl}
		bind:value={query}
		type="search"
		role="combobox"
		aria-expanded={open}
		aria-controls={listboxId}
		aria-autocomplete="list"
		aria-label={ariaLabel}
		{placeholder}
		{disabled}
		class="h-10 w-full rounded-xl border border-slate-300 bg-white pr-9 pl-9 text-sm text-slate-800 transition-colors placeholder:text-slate-400 focus:border-slate-400 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none disabled:opacity-50"
		oninput={onInput}
		onkeydown={onKeydown}
		onfocus={onFocus}
		onblur={() => (open = false)}
	/>

	{#if query.length > 0 && !disabled}
		<button
			type="button"
			aria-label="Clear"
			class="absolute inset-y-0 right-0 flex w-9 items-center justify-center text-slate-400 transition-colors hover:text-slate-700 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
			onclick={clear}
		>
			<Icon icon="heroicons:x-mark" size="sm" />
		</button>
	{/if}

	{#if open}
		<ul
			id={listboxId}
			role="listbox"
			class={[
				'absolute top-full right-0 left-0 mt-1 max-h-64 overflow-y-auto rounded-xl border border-slate-300 bg-white py-1 shadow-md',
				zClasses.asyncDropdown
			].join(' ')}
		>
			{#if loading}
				<li class="px-3 py-2 text-sm text-slate-400">Searching…</li>
			{:else if error}
				<li class="px-3 py-2 text-sm text-red-600">Could not load results. Try again.</li>
			{:else if results.length === 0}
				<li class="px-3 py-2 text-sm text-slate-400">No listings found</li>
			{:else}
				{#each results as listing, index (listing.id)}
					<li role="option" aria-selected={index === activeIndex}>
						<button
							type="button"
							class={[
								'flex w-full items-center justify-between gap-3 px-3 py-1.5 text-left text-sm transition-colors',
								index === activeIndex ? 'bg-slate-100' : 'hover:bg-slate-50'
							].join(' ')}
							onmousedown={(event) => event.preventDefault()}
							onclick={() => select(listing)}
							onmouseenter={() => (activeIndex = index)}
						>
							<span class="flex min-w-0 items-center gap-2">
								{#if value?.id === listing.id}
									<Icon icon="heroicons:check" size="sm" class="shrink-0 text-slate-600" />
								{/if}
								<span class="font-medium text-slate-900">{listing.symbol}</span>
								<span class="text-slate-500">{listing.name}</span>
							</span>
							{#if listing.exchange}
								<span class="shrink-0 text-xs text-slate-400">{listing.exchange}</span>
							{/if}
						</button>
					</li>
				{/each}
			{/if}
		</ul>
	{/if}
</div>

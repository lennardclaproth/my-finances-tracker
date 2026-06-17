<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ListingSearchSelect from './ListingSearchSelect.svelte';
	import { listings } from '$lib/data/fixtures/marketdata';
	import type { Listing } from '$lib/api/types';

	// Deterministic, offline search fns injected so stories don't depend on the service/network.
	const stubSearch = (q: string): Promise<Listing[]> =>
		Promise.resolve(
			listings.filter(
				(l) =>
					l.symbol.toLowerCase().includes(q.toLowerCase()) ||
					l.name.toLowerCase().includes(q.toLowerCase())
			)
		);
	const neverResolves = (): Promise<Listing[]> => new Promise(() => {});
	const empty = (): Promise<Listing[]> => Promise.resolve([]);
	const fails = (): Promise<Listing[]> => Promise.reject(new Error('boom'));

	const { Story } = defineMeta({
		title: 'Molecules/ListingSearchSelect',
		component: ListingSearchSelect,
		tags: ['autodocs']
	});
</script>

<Story name="Results" asChild>
	<div class="min-h-72 w-80">
		<ListingSearchSelect search={stubSearch} query="a" debounceMs={0} />
	</div>
</Story>

<Story name="Loading" asChild>
	<div class="min-h-48 w-80">
		<ListingSearchSelect search={neverResolves} query="aapl" debounceMs={0} />
	</div>
</Story>

<Story name="Empty" asChild>
	<div class="min-h-48 w-80">
		<ListingSearchSelect search={empty} query="zzz" debounceMs={0} />
	</div>
</Story>

<Story name="Error" asChild>
	<div class="min-h-48 w-80">
		<ListingSearchSelect search={fails} query="err" debounceMs={0} />
	</div>
</Story>

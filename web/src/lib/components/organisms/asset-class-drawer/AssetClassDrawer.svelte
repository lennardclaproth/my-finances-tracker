<script lang="ts">
	import Drawer from '$lib/components/organisms/drawer/Drawer.svelte';
	import Money from '$lib/components/atoms/money/Money.svelte';
	import Badge from '$lib/components/atoms/badge/Badge.svelte';
	import { decimalStringToNumber } from '$lib/api/money';
	import { formatDisplayDate } from '$lib/components/molecules/calendar/calendar.utils';
	import type { AssetClassDetails } from '$lib/api/types';

	type Props = {
		open?: boolean;
		details?: AssetClassDetails | null;
		loading?: boolean;
		error?: string | null;
		onClose?: () => void;
	};

	let {
		open = $bindable(false),
		details = null,
		loading = false,
		error = null,
		onClose
	}: Props = $props();

	const growth = $derived(details?.class.growth_pct ?? null);
</script>

<Drawer bind:open title={details?.class.name ?? 'Asset class'} width="max-w-lg" {onClose}>
	{#if loading}
		<div class="space-y-3">
			{#each [0, 1, 2, 3] as i (i)}
				<div class="h-10 w-full animate-pulse rounded-lg bg-slate-100"></div>
			{/each}
		</div>
	{:else if error}
		<div class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
			{error}
		</div>
	{:else if details}
		<div class="space-y-6">
			<div class="flex items-end justify-between gap-3">
				<div>
					<p class="text-xs text-slate-500">Current worth</p>
					<Money
						amount={decimalStringToNumber(details.class.current_worth)}
						currency="EUR"
						size="xl"
						weight="semibold"
					/>
				</div>
				{#if growth !== null}
					<Badge intent={growth >= 0 ? 'success' : 'error'} variant="soft">
						{growth >= 0 ? '+' : ''}{growth.toFixed(2)}%
					</Badge>
				{/if}
			</div>

			<section>
				<h3 class="mb-2 text-sm font-semibold text-slate-900">Assets</h3>
				{#if details.assets.length === 0}
					<p class="text-sm text-slate-400">No assets in this class</p>
				{:else}
					<ul class="divide-y divide-slate-100 rounded-xl border border-slate-200">
						{#each details.assets as asset (asset.id)}
							<li class="flex items-center justify-between gap-3 px-3 py-2 text-sm">
								<span class="min-w-0 truncate text-slate-700">{asset.name}</span>
								<Money
									amount={decimalStringToNumber(asset.current_worth)}
									currency="EUR"
									size="sm"
								/>
							</li>
						{/each}
					</ul>
				{/if}
			</section>

			<section>
				<h3 class="mb-2 text-sm font-semibold text-slate-900">Recent changes</h3>
				{#if details.mutations.length === 0}
					<p class="text-sm text-slate-400">No recorded changes</p>
				{:else}
					<ul class="space-y-2">
						{#each details.mutations as mutation (mutation.id)}
							<li class="flex items-center justify-between gap-3 text-sm">
								<span class="min-w-0">
									<span class="text-slate-700 capitalize">{mutation.change_type}</span>
									<span class="ml-2 text-xs text-slate-400"
										>{formatDisplayDate(mutation.effective_date)}</span
									>
								</span>
								<Money
									amount={decimalStringToNumber(mutation.new_worth)}
									currency="EUR"
									size="sm"
								/>
							</li>
						{/each}
					</ul>
				{/if}
			</section>
		</div>
	{:else}
		<p class="text-sm text-slate-400">Select an asset class to see details.</p>
	{/if}
</Drawer>

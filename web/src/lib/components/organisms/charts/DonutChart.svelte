<script lang="ts">
	import { onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import type { ChartConfiguration } from 'chart.js';
	import { Chart, ensureChartsRegistered } from '$lib/charts/setup';
	import { donutOptions } from '$lib/charts/config';
	import { chartColors, chartStructure } from '$lib/charts/theme';
	import type { DonutDatum } from '$lib/charts/types';
	import Popover from '$lib/components/molecules/popover/Popover.svelte';

	type Props = {
		data: DonutDatum[];
		loading?: boolean;
		error?: string | null;
		/** Slices beyond this count collapse into an "Other" slice + a "+N more" popover. */
		maxSlices?: number;
		/** Color ramp used when a datum has no explicit color. */
		ramp?: readonly string[];
		/** Format a slice value for the legend (e.g. a currency formatter). */
		formatValue?: (value: number) => string;
		centerLabel?: string;
		onSliceClick?: (label: string) => void;
		ariaLabel?: string;
		class?: string;
	};

	const defaultRamp = ['#059669', '#0284c7', '#334155', '#dc2626', '#f59e0b', '#10b981', '#38bdf8'];

	let {
		data,
		loading = false,
		error = null,
		maxSlices = 6,
		ramp = defaultRamp,
		formatValue = (value: number) => value.toLocaleString(),
		centerLabel,
		onSliceClick,
		ariaLabel = 'Distribution chart',
		class: className = ''
	}: Props = $props();

	let canvasEl = $state<HTMLCanvasElement | null>(null);
	let chart: Chart | null = null;
	let moreOpen = $state(false);

	const sorted = $derived([...data].sort((a, b) => b.value - a.value));
	const visible = $derived(sorted.slice(0, maxSlices));
	const rest = $derived(sorted.slice(maxSlices));
	const slices = $derived(
		rest.length > 0
			? [...visible, { label: 'Other', value: rest.reduce((sum, d) => sum + d.value, 0) }]
			: visible
	);
	const sliceColors = $derived(slices.map((d, i) => d.color ?? ramp[i % ramp.length]));
	const total = $derived(data.reduce((sum, d) => sum + d.value, 0));

	const isEmpty = $derived(!loading && !error && (data.length === 0 || total === 0));
	const showCanvas = $derived(!loading && !error && !isEmpty);

	function buildConfig(): ChartConfiguration {
		const current = slices;
		return {
			type: 'doughnut',
			data: {
				labels: current.map((s) => s.label),
				datasets: [
					{
						data: current.map((s) => s.value),
						backgroundColor: sliceColors,
						borderColor: chartColors.sliceBorder,
						borderWidth: 2,
						hoverOffset: chartStructure.hoverOffset
					}
				]
			},
			options: donutOptions((index) => onSliceClick?.(current[index].label))
		} as unknown as ChartConfiguration;
	}

	function create() {
		if (!canvasEl) return;
		ensureChartsRegistered();
		chart = new Chart(canvasEl, buildConfig());
	}

	function destroy() {
		chart?.destroy();
		chart = null;
	}

	$effect(() => {
		if (!browser) return;
		if (!showCanvas || !canvasEl) {
			destroy();
			return;
		}
		if (!chart) {
			create();
			return;
		}
		const config = buildConfig();
		chart.data = config.data;
		chart.options = config.options ?? {};
		chart.update();
	});

	onDestroy(destroy);
</script>

<div class={['flex items-center gap-4', className].filter(Boolean).join(' ')}>
	<div class="relative h-40 w-40 shrink-0" role="img" aria-label={ariaLabel}>
		{#if loading}
			<div class="h-full w-full animate-pulse rounded-full bg-slate-100"></div>
		{:else if error}
			<div
				class="flex h-full w-full items-center justify-center rounded-full border border-red-200 bg-red-50 p-3 text-center text-xs text-red-700"
			>
				{error}
			</div>
		{:else if isEmpty}
			<div
				class="flex h-full w-full items-center justify-center rounded-full border border-dashed border-slate-200 text-xs text-slate-400"
			>
				No data
			</div>
		{:else}
			<canvas bind:this={canvasEl}></canvas>
			{#if centerLabel}
				<div
					class="pointer-events-none absolute inset-0 flex flex-col items-center justify-center text-center"
				>
					<span class="text-sm font-semibold text-slate-900">{formatValue(total)}</span>
					<span class="text-[11px] text-slate-500">{centerLabel}</span>
				</div>
			{/if}
		{/if}
	</div>

	{#if showCanvas}
		<ul class="min-w-0 flex-1 space-y-1 text-sm">
			{#each visible as slice, index (slice.label)}
				<li>
					<button
						type="button"
						class="flex w-full items-center gap-2 rounded-md px-1 py-0.5 text-left transition-colors hover:bg-slate-50"
						onclick={() => onSliceClick?.(slice.label)}
					>
						<span class="size-2.5 shrink-0 rounded-full" style={`background:${sliceColors[index]}`}
						></span>
						<span class="min-w-0 flex-1 truncate text-slate-700">{slice.label || 'Untagged'}</span>
						<span class="shrink-0 text-slate-500">{formatValue(slice.value)}</span>
					</button>
				</li>
			{/each}

			{#if rest.length > 0}
				<li>
					<Popover bind:open={moreOpen} placement="bottom-start">
						{#snippet trigger(api)}
							<button
								type="button"
								class="rounded-md px-1 py-0.5 text-xs font-medium text-slate-500 transition-colors hover:bg-slate-50 hover:text-slate-800"
								aria-expanded={api.open}
								onclick={api.toggle}
							>
								+{rest.length} more
							</button>
						{/snippet}
						<ul class="max-h-56 w-48 space-y-1 overflow-y-auto p-2 text-sm">
							{#each rest as slice (slice.label)}
								<li>
									<button
										type="button"
										class="flex w-full items-center justify-between gap-2 rounded-md px-1 py-0.5 text-left text-slate-700 transition-colors hover:bg-slate-50"
										onclick={() => {
											onSliceClick?.(slice.label);
											moreOpen = false;
										}}
									>
										<span class="min-w-0 truncate">{slice.label || 'Untagged'}</span>
										<span class="shrink-0 text-slate-500">{formatValue(slice.value)}</span>
									</button>
								</li>
							{/each}
						</ul>
					</Popover>
				</li>
			{/if}
		</ul>
	{/if}
</div>

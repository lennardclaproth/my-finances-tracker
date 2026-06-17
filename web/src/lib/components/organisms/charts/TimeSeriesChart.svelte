<script lang="ts">
	import { onDestroy } from 'svelte';
	import { browser } from '$app/environment';
	import type {
		ChartConfiguration,
		ScriptableContext,
		ScriptableLineSegmentContext
	} from 'chart.js';
	import { Chart, ensureChartsRegistered } from '$lib/charts/setup';
	import { timeSeriesOptions } from '$lib/charts/config';
	import { gridZeroLinePlugin } from '$lib/charts/plugins';
	import { rangeSelectPlugin } from '$lib/charts/rangeSelect';
	import { areaGradient } from '$lib/charts/gradient';
	import { chartColors } from '$lib/charts/theme';
	import type { SeriesDataset } from '$lib/charts/types';

	type Props = {
		labels: string[];
		datasets: SeriesDataset[];
		loading?: boolean;
		error?: string | null;
		/** Tailwind height utility for the chart frame. */
		height?: string;
		/** Enable click-or-drag range selection (emits via onRangeSelect). */
		enableRangeSelect?: boolean;
		onRangeSelect?: (from: string, to: string) => void;
		/** Format an ISO label for x-axis display (the raw ISO label is still used for range emission). */
		xTickFormat?: (label: string) => string;
		ariaLabel?: string;
		class?: string;
	};

	let {
		labels,
		datasets,
		loading = false,
		error = null,
		height = 'h-56',
		enableRangeSelect = false,
		onRangeSelect,
		xTickFormat,
		ariaLabel = 'Chart',
		class: className = ''
	}: Props = $props();

	let canvasEl = $state<HTMLCanvasElement | null>(null);
	let chart: Chart | null = null;

	const isEmpty = $derived(
		!loading && !error && (datasets.length === 0 || datasets.every((d) => d.data.length === 0))
	);
	const showCanvas = $derived(!loading && !error && !isEmpty);

	function buildConfig(): ChartConfiguration {
		const mapped = datasets.map((d) => ({
			type: d.type ?? 'line',
			label: d.label,
			data: d.data,
			borderColor: d.color,
			backgroundColor: d.fill
				? (ctx: ScriptableContext<'line'>) => areaGradient(ctx.chart, d.color)
				: d.color,
			fill: d.fill ? 'origin' : false,
			tension: 0.35,
			borderWidth: d.type === 'bar' ? 0 : 2,
			borderDash: d.dashed ? [6, 4] : undefined,
			pointRadius: 0,
			pointHoverRadius: 4,
			borderRadius: d.type === 'bar' ? 4 : 0,
			maxBarThickness: 28,
			...(d.signed
				? {
						segment: {
							borderColor: (c: ScriptableLineSegmentContext) =>
								(c.p1.parsed.y ?? 0) < 0 ? chartColors.negative : d.color
						}
					}
				: {})
		}));

		const options = timeSeriesOptions();
		(options.plugins as Record<string, unknown>).rangeSelect = {
			enabled: enableRangeSelect,
			labels,
			onSelect: (from: string, to: string) => onRangeSelect?.(from, to)
		};
		if (xTickFormat) {
			const scales = options.scales as Record<string, { ticks?: Record<string, unknown> }>;
			if (scales.x?.ticks) {
				scales.x.ticks.callback = (_value: unknown, index: number) =>
					xTickFormat(labels[index] ?? '');
			}
		}

		const baseType = datasets.some((d) => d.type === 'bar') ? 'bar' : 'line';
		return {
			type: baseType,
			data: { labels, datasets: mapped },
			options,
			plugins: [gridZeroLinePlugin, rangeSelectPlugin]
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

	// Single client-only lifecycle: create when the canvas is shown, update on data change, tear down
	// when hidden. Reads labels/datasets/etc. through buildConfig so changes re-run this effect.
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

<div
	class={['relative w-full', height, className].filter(Boolean).join(' ')}
	role="img"
	aria-label={ariaLabel}
>
	{#if loading}
		<div class="h-full w-full animate-pulse rounded-xl bg-slate-100"></div>
	{:else if error}
		<div
			class="flex h-full w-full items-center justify-center rounded-xl border border-red-200 bg-red-50 px-4 text-center text-sm text-red-700"
		>
			{error}
		</div>
	{:else if isEmpty}
		<div class="flex h-full w-full items-center justify-center text-sm text-slate-400">No data</div>
	{:else}
		<canvas bind:this={canvasEl}></canvas>
	{/if}
</div>

/**
 * Click-or-drag range selection (DESIGN_PLAN §5.1) — defined once and shared by all three time-series
 * charts. Implemented as a Chart.js plugin (the Chart.js-native equivalent of the planned
 * `use:rangeSelect` action): it reads the category x-scale to map pixels → indices, draws the live
 * overlay (brand-tinted band + two dashed vertical guides + endpoint markers), and on pointer-up emits
 * a `{from, to}` range of ISO date labels. Travel under the pixel threshold is treated as a single
 * point (`from === to`).
 *
 * Options are read from `chart.options.plugins.rangeSelect`.
 */
import type { Chart, ChartType, Plugin } from 'chart.js';
import { chartColors, chartStructure } from './theme';
import { hexToRgba } from './gradient';

export interface RangeSelectOptions {
	enabled?: boolean;
	/** ISO date labels (one per category index); falls back to `chart.data.labels`. */
	labels?: string[];
	pointThresholdPx?: number;
	onSelect?: (from: string, to: string) => void;
}

interface DragState {
	dragging: boolean;
	startX: number;
	currentX: number;
	cleanup: () => void;
}

const states = new WeakMap<Chart, DragState>();

function readOptions(chart: Chart): RangeSelectOptions | undefined {
	return (chart.options.plugins as unknown as { rangeSelect?: RangeSelectOptions } | undefined)
		?.rangeSelect;
}

function relativeX(chart: Chart, clientX: number): number {
	return clientX - chart.canvas.getBoundingClientRect().left;
}

function indexAt(chart: Chart, px: number): number {
	const area = chart.chartArea;
	const clamped = Math.max(area.left, Math.min(px, area.right));
	const raw = chart.scales.x.getValueForPixel(clamped) ?? 0;
	const max = (chart.data.labels?.length ?? 1) - 1;
	return Math.max(0, Math.min(Math.round(raw), max));
}

export const rangeSelectPlugin: Plugin<ChartType> = {
	id: 'rangeSelect',

	afterInit(chart) {
		const options = readOptions(chart);
		if (!options?.enabled) return;
		const canvas = chart.canvas;

		const onPointerDown = (event: PointerEvent) => {
			if (event.button !== 0) return;
			const x = relativeX(chart, event.clientX);
			const area = chart.chartArea;
			if (x < area.left || x > area.right) return;
			const state = states.get(chart);
			if (!state) return;
			state.dragging = true;
			state.startX = x;
			state.currentX = x;
			canvas.setPointerCapture?.(event.pointerId);
			chart.render();
		};

		const onPointerMove = (event: PointerEvent) => {
			const state = states.get(chart);
			if (!state?.dragging) return;
			state.currentX = relativeX(chart, event.clientX);
			chart.render();
		};

		const finish = (event: PointerEvent) => {
			const state = states.get(chart);
			if (!state?.dragging) return;
			state.dragging = false;

			const opts = readOptions(chart);
			const labels = (opts?.labels ?? (chart.data.labels as string[]) ?? []) as string[];
			const threshold = opts?.pointThresholdPx ?? chartStructure.dragPointThresholdPx;
			const startIdx = indexAt(chart, state.startX);
			const endIdx = indexAt(chart, state.currentX);

			let from: string | undefined;
			let to: string | undefined;
			if (Math.abs(state.currentX - state.startX) < threshold) {
				from = to = labels[startIdx];
			} else {
				const lo = Math.min(startIdx, endIdx);
				const hi = Math.max(startIdx, endIdx);
				from = labels[lo];
				to = labels[hi];
			}

			chart.render();
			try {
				canvas.releasePointerCapture?.(event.pointerId);
			} catch {
				// pointer may already be released
			}
			if (from && to) opts?.onSelect?.(from, to);
		};

		canvas.addEventListener('pointerdown', onPointerDown);
		canvas.addEventListener('pointermove', onPointerMove);
		canvas.addEventListener('pointerup', finish);
		canvas.addEventListener('pointercancel', finish);
		canvas.style.touchAction = 'none';

		states.set(chart, {
			dragging: false,
			startX: 0,
			currentX: 0,
			cleanup() {
				canvas.removeEventListener('pointerdown', onPointerDown);
				canvas.removeEventListener('pointermove', onPointerMove);
				canvas.removeEventListener('pointerup', finish);
				canvas.removeEventListener('pointercancel', finish);
			}
		});
	},

	afterDraw(chart) {
		const state = states.get(chart);
		if (!state?.dragging) return;
		const { ctx, chartArea } = chart;
		const x1 = Math.min(state.startX, state.currentX);
		const x2 = Math.max(state.startX, state.currentX);

		ctx.save();
		// Selection band — the palette's info blue (src/lib/charts/theme.ts).
		ctx.fillStyle = hexToRgba(chartColors.info, 0.1);
		ctx.fillRect(x1, chartArea.top, x2 - x1, chartArea.bottom - chartArea.top);

		// Two dashed vertical guides.
		ctx.setLineDash(chartStructure.gridDash);
		ctx.strokeStyle = chartColors.info;
		ctx.lineWidth = 1;
		for (const gx of [x1, x2]) {
			ctx.beginPath();
			ctx.moveTo(gx, chartArea.top);
			ctx.lineTo(gx, chartArea.bottom);
			ctx.stroke();
		}

		// Circular endpoint markers on the first dataset's line, when available.
		ctx.setLineDash([]);
		const meta = chart.getDatasetMeta(0);
		for (const gx of [x1, x2]) {
			const idx = indexAt(chart, gx);
			const point = meta?.data?.[idx] as { y?: number } | undefined;
			const py = point?.y ?? chartArea.top;
			ctx.beginPath();
			ctx.arc(gx, py, 4, 0, Math.PI * 2);
			ctx.fillStyle = '#ffffff';
			ctx.fill();
			ctx.lineWidth = 2;
			ctx.strokeStyle = chartColors.info;
			ctx.stroke();
		}
		ctx.restore();
	},

	afterDestroy(chart) {
		states.get(chart)?.cleanup();
		states.delete(chart);
	}
};

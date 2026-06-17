import type { Chart } from 'chart.js';

/** Convert a #rrggbb hex to an rgba() string with the given alpha. */
export function hexToRgba(hex: string, alpha: number): string {
	const normalized = hex.replace('#', '');
	const r = parseInt(normalized.slice(0, 2), 16);
	const g = parseInt(normalized.slice(2, 4), 16);
	const b = parseInt(normalized.slice(4, 6), 16);
	return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

/**
 * Build a vertical top→bottom area-fill gradient for `hex` (DESIGN_PLAN §5.2: ~0.2 at top fading to
 * ~0.02 at the bottom). Returns a flat translucent color before the chart area is laid out.
 */
export function areaGradient(
	chart: Chart,
	hex: string,
	topAlpha = 0.2,
	bottomAlpha = 0.02
): CanvasGradient | string {
	const { ctx, chartArea } = chart;
	if (!chartArea) return hexToRgba(hex, topAlpha);
	const gradient = ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
	gradient.addColorStop(0, hexToRgba(hex, topAlpha));
	gradient.addColorStop(1, hexToRgba(hex, bottomAlpha));
	return gradient;
}

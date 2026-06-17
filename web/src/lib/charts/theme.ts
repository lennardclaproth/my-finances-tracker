/**
 * Centralized chart theme (DESIGN_PLAN §2.1, §5.2). This is the single source of chart styling,
 * mirroring the Vue reference's `chartHelpers.ts` but recolored to the target brand: the reference's
 * indigo/rose/cyan are re-mapped to emerald/red/sky/slate. It is intentionally framework-agnostic
 * (plain data + small helpers) so the Chart.js wrappers added in Phase 4 consume these tokens rather
 * than hardcoding hexes. Keep all chart hexes here — components must not hardcode their own.
 */

/** Semantic series colors. Derived from the Tailwind tokens used across the design system. */
export const chartColors = {
	/** income / positive — emerald-600 */
	positive: '#059669',
	/** expense / negative — red-600 / red-700 */
	negative: '#dc2626',
	negativeStrong: '#b91c1c',
	/** net / secondary line — slate-700 */
	net: '#334155',
	/** info — sky-600 */
	info: '#0284c7',
	/** time-range drag-selection on charts — blue-600 (band, guides, endpoint markers) */
	selection: '#2563eb',
	/** hover marker border — brand amber-200 over slate (not the reference's indigo) */
	markerBorder: '#fde68a',
	/** grid + axis text */
	grid: '#e2e8f0',
	zeroLine: '#94a3b8',
	axisText: '#64748b',
	/** dark tooltip — slate-900 surface, slate-50 text */
	tooltipBg: '#0f172a',
	tooltipText: '#f8fafc',
	sliceBorder: '#ffffff'
} as const;

/**
 * Donut ramps (DESIGN_PLAN §2.1): incoming becomes emerald+sky, outgoing becomes red+amber. Slices
 * cycle through the ramp; "+N more" collapses the tail.
 */
export const donutRamps = {
	incoming: ['#059669', '#0ea5e9', '#10b981', '#38bdf8', '#34d399', '#7dd3fc', '#6ee7b7'],
	outgoing: ['#b91c1c', '#f59e0b', '#dc2626', '#fbbf24', '#ef4444', '#fcd34d', '#f87171']
} as const;

/** Structural constants reproduced from the reference (recolor only — geometry unchanged). */
export const chartStructure = {
	gridDash: [4, 4] as number[],
	zeroLineWidth: 1.75,
	areaGradientTop: 0.2,
	areaGradientBottom: 0.02,
	donutCutout: '60%',
	hoverOffset: 3,
	/** click-or-drag range selection: travel under this many px is treated as a single point */
	dragPointThresholdPx: 6
} as const;

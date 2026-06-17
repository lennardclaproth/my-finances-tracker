import type { ChartOptions } from 'chart.js';
import { chartColors, chartStructure } from './theme';

const darkTooltip = {
	backgroundColor: chartColors.tooltipBg,
	titleColor: chartColors.tooltipText,
	bodyColor: chartColors.tooltipText,
	borderWidth: 0,
	padding: 10,
	cornerRadius: 8
};

/**
 * Base options for the time-series (line/bar/area) charts. The built-in y grid is disabled because
 * `gridZeroLinePlugin` draws the dashed grid + solid zero line; the x grid is off (matching the
 * reference). Plugin-specific `rangeSelect` options are merged in by the component.
 */
export function timeSeriesOptions(): ChartOptions {
	return {
		responsive: true,
		maintainAspectRatio: false,
		animation: { duration: 250 },
		interaction: { mode: 'index', intersect: false },
		plugins: {
			legend: { display: false },
			tooltip: { ...darkTooltip, displayColors: true, usePointStyle: true }
		},
		scales: {
			x: {
				grid: { display: false },
				border: { display: false },
				ticks: {
					color: chartColors.axisText,
					font: { size: 11 },
					maxRotation: 0,
					autoSkip: true,
					maxTicksLimit: 8
				}
			},
			y: {
				grid: { display: false },
				border: { display: false },
				ticks: { color: chartColors.axisText, font: { size: 11 }, maxTicksLimit: 6 }
			}
		}
	} as ChartOptions;
}

/** Base options for the donut charts. `onSliceClick` receives the clicked slice index. */
export function donutOptions(onSliceClick?: (index: number) => void): ChartOptions<'doughnut'> {
	return {
		responsive: true,
		maintainAspectRatio: false,
		cutout: chartStructure.donutCutout,
		plugins: {
			legend: { display: false },
			tooltip: { ...darkTooltip }
		},
		onClick: (_event, elements) => {
			if (onSliceClick && elements.length > 0) onSliceClick(elements[0].index);
		}
	} as ChartOptions<'doughnut'>;
}

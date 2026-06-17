import type { ChartType, Plugin } from 'chart.js';
import { chartColors, chartStructure } from './theme';

/**
 * Draws the reference's structural grid: dashed horizontal gridlines at the y ticks plus a solid,
 * darker zero line (width 1.75) when zero is in range. The component disables Chart.js's built-in y
 * grid so this is the single source of horizontal lines.
 */
export const gridZeroLinePlugin: Plugin<ChartType> = {
	id: 'gridZeroLine',
	beforeDatasetsDraw(chart) {
		const y = chart.scales.y;
		if (!y) return;
		const { ctx, chartArea } = chart;
		ctx.save();

		ctx.strokeStyle = chartColors.grid;
		ctx.lineWidth = 1;
		ctx.setLineDash(chartStructure.gridDash);
		for (const tick of y.ticks) {
			const py = y.getPixelForValue(tick.value);
			ctx.beginPath();
			ctx.moveTo(chartArea.left, py);
			ctx.lineTo(chartArea.right, py);
			ctx.stroke();
		}

		if (y.min < 0 && y.max > 0) {
			const pz = y.getPixelForValue(0);
			ctx.setLineDash([]);
			ctx.strokeStyle = chartColors.zeroLine;
			ctx.lineWidth = chartStructure.zeroLineWidth;
			ctx.beginPath();
			ctx.moveTo(chartArea.left, pz);
			ctx.lineTo(chartArea.right, pz);
			ctx.stroke();
		}

		ctx.restore();
	}
};

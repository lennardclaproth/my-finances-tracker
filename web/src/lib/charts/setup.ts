/**
 * Tree-shakeable Chart.js registration. Call `ensureChartsRegistered()` from a chart component's
 * `onMount` (client-only) before constructing a Chart so SSR never imports/executes canvas code.
 */
import {
	Chart,
	ArcElement,
	BarController,
	BarElement,
	CategoryScale,
	DoughnutController,
	Filler,
	Legend,
	LinearScale,
	LineController,
	LineElement,
	PointElement,
	Tooltip
} from 'chart.js';

let registered = false;

export function ensureChartsRegistered(): void {
	if (registered) return;
	Chart.register(
		LineController,
		BarController,
		DoughnutController,
		LineElement,
		PointElement,
		BarElement,
		ArcElement,
		CategoryScale,
		LinearScale,
		Tooltip,
		Filler,
		Legend
	);
	registered = true;
}

export { Chart };

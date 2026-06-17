import type { TrendIndicatorFormat } from '$lib/components/molecules/trend-indicator/trend-indicator.types';

/** One KPI rendered as a stat-card in the KPI row. */
export interface KpiItem {
	label: string;
	amount: number;
	currency?: string;
	locale?: string;
	change?: number;
	changeFormat?: TrendIndicatorFormat;
	trend?: number[];
}

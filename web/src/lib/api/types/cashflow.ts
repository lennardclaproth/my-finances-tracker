import type { PaginatedResponse } from './common';

/** Cashflow direction. */
export type CashflowDirection = 'in' | 'out';

/**
 * One cashflow transaction. Mirrors `cashflow.CreateTransactionResponse`.
 * `amountCents` is 1e6-scaled (see api/money.ts), not hundredths.
 */
export interface CashflowTransaction {
	id: string;
	description: string;
	note: string;
	source: string;
	amountCents: number;
	direction: CashflowDirection;
	/** RFC3339 timestamp. */
	date: string;
	tag: string;
	ignored: boolean;
}

/** `GET /cashflow/transactions` — mirrors `cashflow.GetTransactionsResponse`. */
export type CashflowTransactionsResponse = PaginatedResponse<CashflowTransaction>;

/** One month of aggregated cashflow. Mirrors `cashflow.MonthlyAnalyticsPointResponse`. */
export interface CashflowMonthlyPoint {
	/** "YYYY-MM-DD" (first of month). */
	month: string;
	incoming_cents: number;
	outgoing_cents: number;
	net_cents: number;
}

/** `GET /cashflow/analytics/monthly` — mirrors `cashflow.CashflowMonthlyAnalyticsResponse`. */
export interface CashflowMonthlyAnalyticsResponse {
	data: CashflowMonthlyPoint[];
}

/** One tag total. Mirrors `cashflow.TagDistributionEntryResponse`. */
export interface TagDistributionEntry {
	tag: string;
	totalCents: number;
}

/** `GET /cashflow/analytics/tags` — mirrors `cashflow.TagDistributionResponse`. */
export interface TagDistributionResponse {
	combined: TagDistributionEntry[];
	incoming: TagDistributionEntry[];
	outgoing: TagDistributionEntry[];
}

/** Query filters for `GET /cashflow/transactions`. */
export interface CashflowTransactionsQuery {
	limit?: number;
	offset?: number;
	sort_by?: 'date' | 'description' | 'note' | 'tag' | 'source' | 'amount';
	sort_order?: 'asc' | 'desc';
	q?: string;
	description?: string;
	note?: string;
	source?: string;
	direction?: CashflowDirection;
	/** comma-separated tags */
	tags?: string;
	untagged?: boolean;
	hide_ignored?: boolean;
	from?: string;
	to?: string;
}

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

/** One manual cashflow transaction to create. Mirrors `cashflow.CreateManualCashflowTransactionRequest`. */
export interface CreateManualCashflowTransaction {
	/** "YYYY-MM-DD". */
	date: string;
	/** Non-negative decimal string. */
	amount: string;
	/** Direction the backend accepts: `income`/`in` or `expense`/`out`. */
	type: 'income' | 'expense' | 'in' | 'out';
	description: string;
	note: string;
	tag: string;
	vendor?: string;
}

/** `POST /cashflow/transactions/manual` request — mirrors `cashflow.CreateTransactionsRequest`. */
export interface CreateCashflowTransactionsRequest {
	account_id: string;
	transactions: CreateManualCashflowTransaction[];
}

/** `POST /cashflow/transactions/manual` — mirrors `cashflow.TransactionsResponse`. */
export interface CreateCashflowTransactionsResponse {
	created_count: number;
	data: CashflowTransaction[];
}

/** Shared bulk-mutation filter body. Mirrors `cashflow.TransactionFilters`. */
export interface CashflowTransactionFilters {
	q?: string;
	description?: string;
	note?: string;
	source?: string;
	direction?: string;
	/** comma-separated tags */
	tags?: string;
	untagged?: boolean;
	hide_ignored?: boolean;
	from?: string;
	to?: string;
}

/** `POST /cashflow/transactions/tag` request (single). */
export interface TagTransactionRequest {
	id: string;
	tag: string;
}

/** `POST /cashflow/transactions/tag/selection` request. */
export interface TagTransactionsBySelectionRequest {
	tag: string;
	ids: string[];
}

/** `POST /cashflow/transactions/tag/filter` request. */
export interface TagTransactionsByFilterRequest {
	tag: string;
	account_id?: string;
	filters: CashflowTransactionFilters;
}

/** `POST /cashflow/transactions/ignore/selection` request. */
export interface IgnoreTransactionsBySelectionRequest {
	/** Defaults to true on the backend when omitted. */
	ignored?: boolean;
	ids: string[];
}

/** `POST /cashflow/transactions/ignore/filter` request. */
export interface IgnoreTransactionsByFilterRequest {
	ignored?: boolean;
	filters: CashflowTransactionFilters;
}

/** Result of a bulk tag/ignore mutation. */
export interface CashflowBulkMutationResponse {
	updated_count: number;
	status: string;
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

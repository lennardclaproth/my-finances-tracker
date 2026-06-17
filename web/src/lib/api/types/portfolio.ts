import type { PaginatedResponse } from './common';

/** Portfolio transaction type. */
export type PortfolioTransactionType = 'BUY' | 'SELL' | 'DIVIDEND' | 'TAX' | 'FEE' | 'CASH';

/** Portfolio transaction origin. */
export type PortfolioTransactionOrigin = 'IMPORT' | 'MANUAL';

/**
 * One position with its latest snapshot metrics. Mirrors `portfolio.PositionResponse`.
 * Integer money fields (`cost_basis`, `realized_pnl`, `market_value`) are 1e6-scaled.
 */
export interface PortfolioPosition {
	id: string;
	symbol?: string | null;
	name?: string | null;
	quantity: number;
	cost_basis: number;
	realized_pnl: number;
	market_value?: number | null;
	unrealized_pnl_pct?: number | null;
	/** RFC3339 timestamp. */
	last_snapshot_at?: string | null;
	open_date: string;
	close_date?: string | null;
	is_closed: boolean;
}

/** `GET /portfolio/positions` — mirrors `portfolio.PositionsResponse`. */
export interface PortfolioPositionsResponse {
	include_closed: boolean;
	data: PortfolioPosition[];
}

/**
 * One snapshot point for the portfolio value timeline. Mirrors `portfolio.SnapshotPointResponse`.
 * `market_value`, `total_pnl`, `total_cost_basis` are 1e6-scaled; the `*_pct`/index fields are plain numbers.
 */
export interface PortfolioSnapshotPoint {
	/** RFC3339 timestamp. */
	occurred_at: string;
	market_value: number;
	total_pnl: number;
	total_pnl_pct: number;
	total_cost_basis: number;
	return_vs_cost_basis_pct: number;
	daily_return_pct: number;
	time_weighted_return_pct: number;
	value_index: number;
}

/** `GET /portfolio/snapshots` returns a bare array of snapshot points. */
export type PortfolioSnapshotsResponse = PortfolioSnapshotPoint[];

/**
 * One portfolio transaction. Mirrors `portfolio.PortfolioTransactionResponse`.
 * Money fields (`amount`, `quantity`, `unit_price`) are decimal strings, not scaled integers.
 */
export interface PortfolioTransaction {
	id: string;
	account_id: string;
	origin: PortfolioTransactionOrigin;
	source: string;
	/** RFC3339 timestamp. */
	occurred_at: string;
	type: PortfolioTransactionType;
	listing_id?: string | null;
	isin?: string | null;
	symbol?: string | null;
	description: string;
	amount: string;
	quantity: string;
	unit_price: string;
	created_at: string;
	updated_at: string;
}

/** `GET /portfolio/transactions` — mirrors `portfolio.PortfolioTransactionsResponse`. */
export type PortfolioTransactionsResponse = PaginatedResponse<PortfolioTransaction>;

/** Query filters for `GET /portfolio/positions`. */
export interface PortfolioPositionsQuery {
	account_id: string;
	include_closed?: boolean;
}

/** Query filters for `GET /portfolio/transactions`. */
export interface PortfolioTransactionsQuery {
	account_id: string;
	from?: string;
	to?: string;
	limit?: 10 | 25 | 50 | 100;
	offset?: number;
	sort_by?: 'date';
	sort_order?: 'asc' | 'desc';
	q?: string;
	type?: PortfolioTransactionType;
	origin?: PortfolioTransactionOrigin;
	source?: string;
	listing?: string;
}

/** Query filters for `GET /portfolio/snapshots`. */
export interface PortfolioSnapshotsQuery {
	account_id: string;
	from?: string;
	to?: string;
}

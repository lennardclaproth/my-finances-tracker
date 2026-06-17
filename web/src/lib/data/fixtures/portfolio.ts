import { numberToScaled as s } from '$lib/api/money';
import type {
	PortfolioPosition,
	PortfolioSnapshotPoint,
	PortfolioTransaction
} from '$lib/api/types';
import { DEMO_ACCOUNT_ID } from '$lib/api/config';

/**
 * Current + one closed position. Integer money fields (`cost_basis`, `realized_pnl`, `market_value`)
 * are 1e6-scaled; `unrealized_pnl_pct` is a percentage number (e.g. 18.42 ⇒ +18.42%).
 */
export const portfolioPositions: PortfolioPosition[] = [
	{
		id: 'pos-vwrl',
		symbol: 'VWRL',
		name: 'Vanguard FTSE All-World',
		quantity: 42,
		cost_basis: s(4200),
		realized_pnl: s(0),
		market_value: s(4973.4),
		unrealized_pnl_pct: 18.41,
		last_snapshot_at: '2026-06-16T22:00:00Z',
		open_date: '2025-02-11T10:00:00Z',
		close_date: null,
		is_closed: false
	},
	{
		id: 'pos-iwda',
		symbol: 'IWDA',
		name: 'iShares Core MSCI World',
		quantity: 30,
		cost_basis: s(2700),
		realized_pnl: s(0),
		market_value: s(3105),
		unrealized_pnl_pct: 15.0,
		last_snapshot_at: '2026-06-16T22:00:00Z',
		open_date: '2025-03-20T10:00:00Z',
		close_date: null,
		is_closed: false
	},
	{
		id: 'pos-aapl',
		symbol: 'AAPL',
		name: 'Apple Inc.',
		quantity: 12,
		cost_basis: s(1980),
		realized_pnl: s(0),
		market_value: s(2256),
		unrealized_pnl_pct: 13.94,
		last_snapshot_at: '2026-06-16T22:00:00Z',
		open_date: '2025-05-02T14:30:00Z',
		close_date: null,
		is_closed: false
	},
	{
		id: 'pos-asml',
		symbol: 'ASML',
		name: 'ASML Holding',
		quantity: 3,
		cost_basis: s(2010),
		realized_pnl: s(0),
		market_value: s(1896),
		unrealized_pnl_pct: -5.67,
		last_snapshot_at: '2026-06-16T22:00:00Z',
		open_date: '2025-09-15T09:00:00Z',
		close_date: null,
		is_closed: false
	},
	{
		id: 'pos-tsla',
		symbol: 'TSLA',
		name: 'Tesla Inc.',
		quantity: 0,
		cost_basis: s(0),
		realized_pnl: s(312.5),
		market_value: null,
		unrealized_pnl_pct: null,
		last_snapshot_at: null,
		open_date: '2025-01-22T15:00:00Z',
		close_date: '2026-02-28T16:00:00Z',
		is_closed: true
	}
];

function monthlySnapshots(): PortfolioSnapshotPoint[] {
	// Deterministic 12-point monthly value curve, gently rising with one dip.
	const values = [9000, 9180, 9050, 9420, 9610, 9880, 9740, 10120, 10450, 10380, 10790, 12230.4];
	const costBasis = [8400, 8550, 8550, 8800, 8800, 9000, 9000, 9300, 9600, 9600, 9890, 10890];
	const start = new Date(Date.UTC(2025, 6, 1)); // 2025-07
	return values.map((value, i) => {
		const occurred = new Date(start);
		occurred.setUTCMonth(start.getUTCMonth() + i);
		const cb = costBasis[i];
		const pnl = value - cb;
		return {
			occurred_at: occurred.toISOString().replace('.000', ''),
			market_value: s(value),
			total_pnl: s(pnl),
			total_pnl_pct: Number(((pnl / cb) * 100).toFixed(2)),
			total_cost_basis: s(cb),
			return_vs_cost_basis_pct: Number(((pnl / cb) * 100).toFixed(2)),
			daily_return_pct: Number(
				(((value - values[Math.max(0, i - 1)]) / values[Math.max(0, i - 1)]) * 100).toFixed(2)
			),
			time_weighted_return_pct: Number(((value / values[0] - 1) * 100).toFixed(2)),
			value_index: Number(((value / values[0]) * 100).toFixed(2))
		};
	});
}

/** Portfolio value timeline (12 monthly points). Money fields 1e6-scaled; `value_index` starts at 100. */
export const portfolioSnapshots: PortfolioSnapshotPoint[] = monthlySnapshots();

/** Portfolio transactions (decimal-string money). Newest first; enough rows to page at 10/25. */
export const portfolioTransactions: PortfolioTransaction[] = [
	{
		id: 'ptx-001',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'degiro',
		occurred_at: '2026-06-10T09:31:00Z',
		type: 'BUY',
		listing_id: 'lst-vwrl',
		isin: 'IE00B3RBWM25',
		symbol: 'VWRL',
		description: 'Vanguard FTSE All-World',
		amount: '592.000000',
		quantity: '5',
		unit_price: '118.400000',
		created_at: '2026-06-10T09:31:05Z',
		updated_at: '2026-06-10T09:31:05Z'
	},
	{
		id: 'ptx-002',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'degiro',
		occurred_at: '2026-06-03T11:02:00Z',
		type: 'DIVIDEND',
		listing_id: 'lst-iwda',
		isin: 'IE00B4L5Y983',
		symbol: 'IWDA',
		description: 'Dividend IWDA',
		amount: '24.180000',
		quantity: '0',
		unit_price: '0',
		created_at: '2026-06-03T11:02:05Z',
		updated_at: '2026-06-03T11:02:05Z'
	},
	{
		id: 'ptx-003',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'MANUAL',
		source: 'manual',
		occurred_at: '2026-05-28T16:00:00Z',
		type: 'SELL',
		listing_id: 'lst-tsla',
		isin: 'US88160R1014',
		symbol: 'TSLA',
		description: 'Close Tesla position',
		amount: '2312.500000',
		quantity: '10',
		unit_price: '231.250000',
		created_at: '2026-05-28T16:00:05Z',
		updated_at: '2026-05-28T16:00:05Z'
	},
	{
		id: 'ptx-004',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'degiro',
		occurred_at: '2026-05-15T10:15:00Z',
		type: 'BUY',
		listing_id: 'lst-aapl',
		isin: 'US0378331005',
		symbol: 'AAPL',
		description: 'Apple Inc.',
		amount: '660.000000',
		quantity: '4',
		unit_price: '165.000000',
		created_at: '2026-05-15T10:15:05Z',
		updated_at: '2026-05-15T10:15:05Z'
	},
	{
		id: 'ptx-005',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'degiro',
		occurred_at: '2026-05-02T09:00:00Z',
		type: 'FEE',
		listing_id: null,
		isin: null,
		symbol: null,
		description: 'Connectivity fee',
		amount: '2.500000',
		quantity: '0',
		unit_price: '0',
		created_at: '2026-05-02T09:00:05Z',
		updated_at: '2026-05-02T09:00:05Z'
	},
	{
		id: 'ptx-006',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'brandnewday',
		occurred_at: '2026-04-25T08:00:00Z',
		type: 'CASH',
		listing_id: null,
		isin: null,
		symbol: null,
		description: 'Monthly deposit',
		amount: '300.000000',
		quantity: '0',
		unit_price: '0',
		created_at: '2026-04-25T08:00:05Z',
		updated_at: '2026-04-25T08:00:05Z'
	},
	{
		id: 'ptx-007',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'degiro',
		occurred_at: '2026-04-18T13:45:00Z',
		type: 'BUY',
		listing_id: 'lst-asml',
		isin: 'NL0010273215',
		symbol: 'ASML',
		description: 'ASML Holding',
		amount: '2010.000000',
		quantity: '3',
		unit_price: '670.000000',
		created_at: '2026-04-18T13:45:05Z',
		updated_at: '2026-04-18T13:45:05Z'
	},
	{
		id: 'ptx-008',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'degiro',
		occurred_at: '2026-04-04T10:00:00Z',
		type: 'TAX',
		listing_id: 'lst-iwda',
		isin: 'IE00B4L5Y983',
		symbol: 'IWDA',
		description: 'Dividend tax',
		amount: '3.620000',
		quantity: '0',
		unit_price: '0',
		created_at: '2026-04-04T10:00:05Z',
		updated_at: '2026-04-04T10:00:05Z'
	},
	{
		id: 'ptx-009',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'degiro',
		occurred_at: '2026-03-20T09:30:00Z',
		type: 'BUY',
		listing_id: 'lst-iwda',
		isin: 'IE00B4L5Y983',
		symbol: 'IWDA',
		description: 'iShares Core MSCI World',
		amount: '900.000000',
		quantity: '10',
		unit_price: '90.000000',
		created_at: '2026-03-20T09:30:05Z',
		updated_at: '2026-03-20T09:30:05Z'
	},
	{
		id: 'ptx-010',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'degiro',
		occurred_at: '2026-02-11T10:00:00Z',
		type: 'BUY',
		listing_id: 'lst-vwrl',
		isin: 'IE00B3RBWM25',
		symbol: 'VWRL',
		description: 'Vanguard FTSE All-World',
		amount: '4200.000000',
		quantity: '42',
		unit_price: '100.000000',
		created_at: '2026-02-11T10:00:05Z',
		updated_at: '2026-02-11T10:00:05Z'
	},
	{
		id: 'ptx-011',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'MANUAL',
		source: 'manual',
		occurred_at: '2026-01-30T12:00:00Z',
		type: 'CASH',
		listing_id: null,
		isin: null,
		symbol: null,
		description: 'Opening deposit',
		amount: '5000.000000',
		quantity: '0',
		unit_price: '0',
		created_at: '2026-01-30T12:00:05Z',
		updated_at: '2026-01-30T12:00:05Z'
	},
	{
		id: 'ptx-012',
		account_id: DEMO_ACCOUNT_ID,
		origin: 'IMPORT',
		source: 'degiro',
		occurred_at: '2026-01-22T15:00:00Z',
		type: 'BUY',
		listing_id: 'lst-tsla',
		isin: 'US88160R1014',
		symbol: 'TSLA',
		description: 'Tesla Inc.',
		amount: '2000.000000',
		quantity: '10',
		unit_price: '200.000000',
		created_at: '2026-01-22T15:00:05Z',
		updated_at: '2026-01-22T15:00:05Z'
	}
];

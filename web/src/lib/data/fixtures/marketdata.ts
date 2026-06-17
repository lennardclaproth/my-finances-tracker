import { numberToScaled as s } from '$lib/api/money';
import type { EOD, Listing } from '$lib/api/types';

/** Listing master data. Symbols/ids line up with the portfolio fixtures. */
export const listings: Listing[] = [
	{
		id: 'lst-vwrl',
		symbol: 'VWRL',
		name: 'Vanguard FTSE All-World UCITS ETF',
		source: 'marketstack',
		description: 'Global all-cap equity ETF',
		exchange: 'AEX',
		region: 'Netherlands',
		currency: 'EUR',
		isin: 'IE00B3RBWM25',
		ticker: 'VWRL.AS',
		type: 'etf',
		created_at: '2025-01-04T09:00:00Z',
		updated_at: '2026-06-16T22:00:00Z'
	},
	{
		id: 'lst-iwda',
		symbol: 'IWDA',
		name: 'iShares Core MSCI World UCITS ETF',
		source: 'marketstack',
		description: 'Developed-markets equity ETF',
		exchange: 'AEX',
		region: 'Netherlands',
		currency: 'EUR',
		isin: 'IE00B4L5Y983',
		ticker: 'IWDA.AS',
		type: 'etf',
		created_at: '2025-01-04T09:00:00Z',
		updated_at: '2026-06-16T22:00:00Z'
	},
	{
		id: 'lst-aapl',
		symbol: 'AAPL',
		name: 'Apple Inc.',
		source: 'alphavantage',
		description: null,
		exchange: 'NASDAQ',
		region: 'United States',
		currency: 'USD',
		isin: 'US0378331005',
		ticker: 'AAPL',
		type: 'stock',
		created_at: '2025-01-04T09:00:00Z',
		updated_at: '2026-06-16T22:00:00Z'
	},
	{
		id: 'lst-asml',
		symbol: 'ASML',
		name: 'ASML Holding NV',
		source: 'marketstack',
		description: null,
		exchange: 'AEX',
		region: 'Netherlands',
		currency: 'EUR',
		isin: 'NL0010273215',
		ticker: 'ASML.AS',
		type: 'stock',
		created_at: '2025-01-04T09:00:00Z',
		updated_at: '2026-06-16T22:00:00Z'
	},
	{
		id: 'lst-tsla',
		symbol: 'TSLA',
		name: 'Tesla Inc.',
		source: 'alphavantage',
		description: null,
		exchange: 'NASDAQ',
		region: 'United States',
		currency: 'USD',
		isin: 'US88160R1014',
		ticker: 'TSLA',
		type: 'stock',
		created_at: '2025-01-04T09:00:00Z',
		updated_at: '2026-06-16T22:00:00Z'
	}
];

/** Build a deterministic 30-business-day OHLCV series for a listing (PascalCase keys, 1e6-scaled prices). */
function eodSeries(listingId: string, symbol: string, startClose: number): EOD[] {
	const rows: EOD[] = [];
	let close = startClose;
	const day = new Date(Date.UTC(2026, 4, 4)); // 2026-05-04, a Monday
	for (let i = 0; i < 30; i++) {
		// Deterministic pseudo-walk (no Math.random): a smooth sine drift + small alternating step.
		const drift = Math.sin(i / 4) * (startClose * 0.012);
		const step = (i % 2 === 0 ? 1 : -1) * (startClose * 0.004);
		const open = close;
		close = Number((startClose + drift + step).toFixed(2));
		const high = Number(Math.max(open, close) * 1.006).toFixed(2);
		const low = Number(Math.min(open, close) * 0.994).toFixed(2);
		const date = new Date(day);
		date.setUTCDate(day.getUTCDate() + i);
		rows.push({
			ID: `eod-${listingId}-${i}`,
			ListingID: listingId,
			Symbol: symbol,
			Date: date.toISOString().replace('.000', ''),
			Open: s(open),
			Close: s(close),
			High: s(Number(high)),
			Low: s(Number(low)),
			Volume: 100000 + i * 1234,
			CreatedAt: '2026-06-16T22:00:00Z',
			UpdatedAt: '2026-06-16T22:00:00Z'
		});
	}
	return rows;
}

/** EOD rows keyed by listing id (dailies admin page + charts). */
export const eodByListing: Record<string, EOD[]> = {
	'lst-vwrl': eodSeries('lst-vwrl', 'VWRL', 118.4),
	'lst-aapl': eodSeries('lst-aapl', 'AAPL', 188.0),
	'lst-asml': eodSeries('lst-asml', 'ASML', 632.0)
};

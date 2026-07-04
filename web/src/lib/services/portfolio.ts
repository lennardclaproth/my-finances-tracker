import { apiGet, apiSend } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	CreateManualPortfolioTransactionRequest,
	PortfolioPositionsQuery,
	PortfolioPositionsResponse,
	PortfolioSnapshotsQuery,
	PortfolioSnapshotsResponse,
	PortfolioTransaction,
	PortfolioTransactionsQuery,
	PortfolioTransactionsResponse,
	RebuildPortfolioRequest
} from '$lib/api/types';
import {
	portfolioPositions,
	portfolioSnapshots,
	portfolioTransactions
} from '$lib/data/fixtures/portfolio';
import { clone, contains, delay } from './_mock';

/** `GET /portfolio/positions` */
export async function listPortfolioPositions(
	query: PortfolioPositionsQuery
): Promise<PortfolioPositionsResponse> {
	if (useMocks) {
		await delay();
		const data = clone(
			portfolioPositions.filter((p) => (query.include_closed ? true : !p.is_closed))
		);
		return { include_closed: Boolean(query.include_closed), data };
	}
	return apiGet<PortfolioPositionsResponse>('/portfolio/positions', { ...query });
}

/** `GET /portfolio/snapshots` */
export async function getPortfolioSnapshots(
	query: PortfolioSnapshotsQuery
): Promise<PortfolioSnapshotsResponse> {
	if (useMocks) {
		await delay();
		return clone(
			portfolioSnapshots.filter(
				(p) =>
					(!query.from || p.occurred_at.slice(0, 10) >= query.from) &&
					(!query.to || p.occurred_at.slice(0, 10) <= query.to)
			)
		);
	}
	return apiGet<PortfolioSnapshotsResponse>('/portfolio/snapshots', { ...query });
}

/** `GET /portfolio/transactions` */
export async function listPortfolioTransactions(
	query: PortfolioTransactionsQuery
): Promise<PortfolioTransactionsResponse> {
	if (useMocks) {
		await delay();
		let rows = portfolioTransactions.filter((tx) => {
			if (query.type && tx.type !== query.type) return false;
			if (query.origin && tx.origin !== query.origin) return false;
			if (!contains(tx.source, query.source)) return false;
			if (query.from && tx.occurred_at.slice(0, 10) < query.from) return false;
			if (query.to && tx.occurred_at.slice(0, 10) > query.to) return false;
			if (query.listing) {
				const l = query.listing.toLowerCase();
				const hit =
					(tx.symbol ?? '').toLowerCase().includes(l) || (tx.isin ?? '').toLowerCase().includes(l);
				if (!hit) return false;
			}
			if (query.q) {
				const q = query.q.toLowerCase();
				const hit =
					tx.description.toLowerCase().includes(q) ||
					tx.source.toLowerCase().includes(q) ||
					(tx.symbol ?? '').toLowerCase().includes(q) ||
					(tx.isin ?? '').toLowerCase().includes(q);
				if (!hit) return false;
			}
			return true;
		});

		const dir = query.sort_order === 'asc' ? 1 : -1;
		rows = rows.slice().sort((a, b) => a.occurred_at.localeCompare(b.occurred_at) * dir);

		const total = rows.length;
		const limit = query.limit ?? 25;
		const offset = query.offset ?? 0;
		const data = clone(rows.slice(offset, offset + limit));
		return { pagination: { limit, offset, count: data.length, total }, data };
	}
	return apiGet<PortfolioTransactionsResponse>('/portfolio/transactions', { ...query });
}

/** `POST /portfolio/transactions/manual` */
export async function createManualPortfolioTransaction(
	body: CreateManualPortfolioTransactionRequest
): Promise<PortfolioTransaction> {
	if (useMocks) {
		await delay();
		const now = new Date().toISOString();
		return {
			id: crypto.randomUUID(),
			account_id: body.account_id,
			origin: 'MANUAL',
			source: 'manual',
			occurred_at: body.occurred_at,
			type: body.type,
			listing_id: body.listing_id ?? null,
			isin: null,
			symbol: null,
			description: body.description ?? '',
			amount: body.amount,
			quantity: body.quantity ?? '0',
			unit_price: '0',
			created_at: now,
			updated_at: now
		};
	}
	return apiSend<PortfolioTransaction>('POST', '/portfolio/transactions/manual', body);
}

/** `POST /portfolio/rebuild` */
export async function rebuildPortfolio(body: RebuildPortfolioRequest): Promise<void> {
	if (useMocks) {
		await delay();
		return;
	}
	await apiSend<unknown>('POST', '/portfolio/rebuild', body);
}

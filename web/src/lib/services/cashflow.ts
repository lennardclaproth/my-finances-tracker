import { apiGet } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	CashflowMonthlyAnalyticsResponse,
	CashflowTransaction,
	CashflowTransactionsQuery,
	CashflowTransactionsResponse,
	TagDistributionResponse
} from '$lib/api/types';
import {
	cashflowMonthly,
	cashflowTagDistribution,
	cashflowTransactions
} from '$lib/data/fixtures/cashflow';
import { clone, contains, delay } from './_mock';

function compare(a: CashflowTransaction, b: CashflowTransaction, sortBy: string): number {
	switch (sortBy) {
		case 'amount':
			return a.amountCents - b.amountCents;
		case 'description':
			return a.description.localeCompare(b.description);
		case 'note':
			return a.note.localeCompare(b.note);
		case 'tag':
			return a.tag.localeCompare(b.tag);
		case 'source':
			return a.source.localeCompare(b.source);
		case 'date':
		default:
			return a.date.localeCompare(b.date);
	}
}

function mockTransactions(query: CashflowTransactionsQuery): CashflowTransactionsResponse {
	const tags = (query.tags ?? '')
		.split(',')
		.map((t) => t.trim())
		.filter(Boolean);

	let rows = cashflowTransactions.filter((tx) => {
		if (query.direction && tx.direction !== query.direction) return false;
		if (query.hide_ignored && tx.ignored) return false;
		if (query.untagged && tx.tag !== '') return false;
		if (tags.length > 0 && !tags.includes(tx.tag)) return false;
		if (!contains(tx.description, query.description)) return false;
		if (!contains(tx.note, query.note)) return false;
		if (!contains(tx.source, query.source)) return false;
		if (query.from && tx.date.slice(0, 10) < query.from) return false;
		if (query.to && tx.date.slice(0, 10) > query.to) return false;
		if (query.q) {
			const q = query.q.toLowerCase();
			const hit =
				tx.description.toLowerCase().includes(q) ||
				tx.note.toLowerCase().includes(q) ||
				tx.tag.toLowerCase().includes(q);
			if (!hit) return false;
		}
		return true;
	});

	const sortBy = query.sort_by ?? 'date';
	const dir = query.sort_order === 'asc' ? 1 : -1;
	rows = rows.slice().sort((a, b) => compare(a, b, sortBy) * dir);

	const total = rows.length;
	const limit = query.limit ?? 100;
	const offset = query.offset ?? 0;
	const data = clone(rows.slice(offset, offset + limit));

	return { pagination: { limit, offset, count: data.length, total }, data };
}

/** `GET /cashflow/transactions` */
export async function listCashflowTransactions(
	query: CashflowTransactionsQuery = {}
): Promise<CashflowTransactionsResponse> {
	if (useMocks) {
		await delay();
		return mockTransactions(query);
	}
	return apiGet<CashflowTransactionsResponse>('/cashflow/transactions', { ...query });
}

/** `GET /cashflow/analytics/monthly` */
export async function getCashflowMonthly(
	query: { from?: string; to?: string; include_ignored?: boolean } = {}
): Promise<CashflowMonthlyAnalyticsResponse> {
	if (useMocks) {
		await delay();
		const data = clone(
			cashflowMonthly.filter(
				(p) => (!query.from || p.month >= query.from) && (!query.to || p.month <= query.to)
			)
		);
		return { data };
	}
	return apiGet<CashflowMonthlyAnalyticsResponse>('/cashflow/analytics/monthly', { ...query });
}

/** `GET /cashflow/analytics/tags` */
export async function getCashflowTagDistribution(
	query: { from?: string; to?: string; include_ignored?: boolean } = {}
): Promise<TagDistributionResponse> {
	if (useMocks) {
		await delay();
		return clone(cashflowTagDistribution);
	}
	return apiGet<TagDistributionResponse>('/cashflow/analytics/tags', { ...query });
}

import { apiGet } from '$lib/api/client';
import { useMocks } from '$lib/api/config';
import type {
	EODQuery,
	EODResponse,
	ListingsResponse,
	ListingsSearchQuery,
	ListingsSearchResponse
} from '$lib/api/types';
import { eodByListing, listings } from '$lib/data/fixtures/marketdata';
import { clone, delay } from './_mock';

/** `GET /marketdata/listings` */
export async function listListings(): Promise<ListingsResponse> {
	if (useMocks) {
		await delay();
		return clone(listings.slice().sort((a, b) => a.symbol.localeCompare(b.symbol)));
	}
	return apiGet<ListingsResponse>('/marketdata/listings');
}

/** `GET /marketdata/listings/search` */
export async function searchListings(query: ListingsSearchQuery): Promise<ListingsSearchResponse> {
	if (useMocks) {
		await delay();
		const q = query.q.toLowerCase();
		const matched = listings.filter(
			(l) =>
				l.symbol.toLowerCase().includes(q) ||
				l.name.toLowerCase().includes(q) ||
				(l.isin ?? '').toLowerCase().includes(q)
		);
		const limit = query.limit ?? 25;
		const offset = query.offset ?? 0;
		const data = clone(matched.slice(offset, offset + limit));
		return { pagination: { limit, offset, count: data.length, total: matched.length }, data };
	}
	return apiGet<ListingsSearchResponse>('/marketdata/listings/search', { ...query });
}

/** `GET /marketdata/eods` */
export async function getEOD(query: EODQuery): Promise<EODResponse> {
	if (useMocks) {
		await delay();
		const listing = query.listing_id
			? listings.find((l) => l.id === query.listing_id)
			: listings.find((l) => l.symbol.toLowerCase() === (query.symbol ?? '').toLowerCase());
		const all = (listing && eodByListing[listing.id]) ?? [];
		const filtered = all.filter(
			(e) =>
				(!query.from || e.Date.slice(0, 10) >= query.from) &&
				(!query.to || e.Date.slice(0, 10) <= query.to)
		);
		const dir = query.sort_order === 'desc' ? -1 : 1;
		const sorted = filtered.slice().sort((a, b) => a.Date.localeCompare(b.Date) * dir);
		const limit = query.limit ?? 100;
		const offset = query.offset ?? 0;
		const data = clone(sorted.slice(offset, offset + limit));
		return {
			Data: data,
			Metadata: { Message: '', ResultCount: data.length, TotalCount: sorted.length }
		};
	}
	return apiGet<EODResponse>('/marketdata/eods', { ...query });
}

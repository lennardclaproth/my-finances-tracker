import type { PaginatedResponse } from './common';

/** One market listing. Mirrors `marketdata.ListingResponse`. */
export interface Listing {
	id: string;
	symbol: string;
	name: string;
	source: string;
	description?: string | null;
	exchange?: string | null;
	region?: string | null;
	currency?: string | null;
	isin?: string | null;
	ticker?: string | null;
	type?: string | null;
	created_at: string;
	updated_at: string;
}

/** `GET /marketdata/listings` returns a bare array of listings. */
export type ListingsResponse = Listing[];

/** `GET /marketdata/listings/search` — mirrors `marketdata.ListingsSearchResponse`. */
export type ListingsSearchResponse = PaginatedResponse<Listing>;

/**
 * One end-of-day OHLCV row. Mirrors the Go `marketdata.EOD` struct, which is serialized **without**
 * json tags — so the keys are PascalCase and prices are raw 1e6-scaled integers.
 */
export interface EOD {
	ID: string;
	ListingID: string;
	Symbol: string;
	/** RFC3339 timestamp. */
	Date: string;
	Open: number;
	Close: number;
	High: number;
	Low: number;
	Volume: number;
	CreatedAt: string;
	UpdatedAt: string;
}

/** EOD retrieval metadata. Mirrors `marketdata.GetEODMetadataResponse` (also PascalCase). */
export interface EODMetadata {
	Message: string;
	ResultCount: number;
	TotalCount: number;
}

/** `GET /marketdata/eods` — mirrors `marketdata.GetEODResponse` (PascalCase envelope). */
export interface EODResponse {
	Data: EOD[];
	Metadata: EODMetadata;
}

/** `POST /marketdata/listing` request — mirrors `marketdata.CreateListingRequest`. */
export interface CreateListingRequest {
	name: string;
	symbol: string;
	source: string;
	description?: string;
	exchange?: string;
	region?: string;
	/** ISO currency code; validated by the backend if present. */
	currency?: string;
	isin?: string;
	ticker?: string;
	type?: string;
}

/** `PATCH /marketdata/listing` request — mirrors `marketdata.UpdateListingFieldsRequest`. */
export interface UpdateListingFieldsRequest {
	id: string;
	description?: string;
	exchange?: string;
	region?: string;
	currency?: string;
	isin?: string;
	ticker?: string;
	type?: string;
}

/** Query filters for `GET /marketdata/listings/search`. */
export interface ListingsSearchQuery {
	q: string;
	limit?: number;
	offset?: number;
}

/** Query filters for `GET /marketdata/eods`. */
export interface EODQuery {
	listing_id?: string;
	symbol?: string;
	from?: string;
	to?: string;
	sort_order?: 'asc' | 'desc';
	limit?: number;
	offset?: number;
}

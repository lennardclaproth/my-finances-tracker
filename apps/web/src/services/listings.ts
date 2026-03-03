import type { CreateListingPayload, Listing, ListingDto, ListingsSearchResponse } from "../types/listings";
import { requestJson } from "./http";

function toListing(dto: ListingDto): Listing {
  return {
    id: dto.id,
    symbol: dto.symbol,
    name: dto.name,
    source: dto.source,
    description: dto.description,
    exchange: dto.exchange,
    region: dto.region,
    currency: dto.currency,
    isin: dto.isin,
    ticker: dto.ticker,
    type: dto.type,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  };
}

function normalizeOptional(value?: string): string | undefined {
  if (value === undefined) {
    return undefined;
  }
  const trimmed = value.trim();
  return trimmed === "" ? undefined : trimmed;
}

export async function fetchListings(): Promise<Listing[]> {
  const payload = await requestJson<ListingDto[]>("/marketdata/listings", {
    method: "GET",
  });
  return payload.map(toListing);
}

export async function createListing(input: CreateListingPayload): Promise<Listing> {
  const body: CreateListingPayload = {
    name: input.name.trim(),
    symbol: input.symbol.trim(),
    source: input.source.trim(),
    description: normalizeOptional(input.description),
    exchange: normalizeOptional(input.exchange),
    region: normalizeOptional(input.region),
    currency: normalizeOptional(input.currency),
    isin: normalizeOptional(input.isin),
    ticker: normalizeOptional(input.ticker),
    type: normalizeOptional(input.type),
  };

  const dto = await requestJson<ListingDto>("/marketdata/listing", {
    method: "POST",
    body: JSON.stringify(body),
  });
  return toListing(dto);
}

export async function searchListings(
  q: string,
  limit = 25,
  offset = 0,
): Promise<{ pagination: ListingsSearchResponse["pagination"]; data: Listing[] }> {
  const payload = await requestJson<ListingsSearchResponse>("/marketdata/listings/search", {
    method: "GET",
    query: {
      q: q.trim(),
      limit,
      offset,
    },
  });

  return {
    pagination: payload.pagination,
    data: payload.data.map(toListing),
  };
}

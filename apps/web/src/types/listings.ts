export interface ListingDto {
  id: string;
  symbol: string;
  name: string;
  source: string;
  description?: string;
  exchange?: string;
  region?: string;
  currency?: string;
  isin?: string;
  ticker?: string;
  type?: string;
  created_at: string;
  updated_at: string;
}

export interface Listing {
  id: string;
  symbol: string;
  name: string;
  source: string;
  description?: string;
  exchange?: string;
  region?: string;
  currency?: string;
  isin?: string;
  ticker?: string;
  type?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateListingPayload {
  name: string;
  symbol: string;
  source: string;
  description?: string;
  exchange?: string;
  region?: string;
  currency?: string;
  isin?: string;
  ticker?: string;
  type?: string;
}

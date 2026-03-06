export interface PortfolioSnapshotPointDto {
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

export interface PortfolioPositionDto {
  id: string;
  symbol?: string;
  name?: string;
  quantity: number;
  cost_basis: number;
  realized_pnl: number;
  market_value?: number;
  unrealized_pnl_pct?: number;
  last_snapshot_at?: string;
  open_date: string;
  close_date?: string;
  is_closed: boolean;
}

export interface PortfolioPositionsResponseDto {
  include_closed: boolean;
  data: PortfolioPositionDto[];
}

export interface PortfolioSnapshotPoint {
  occurredAt: string;
  marketValue: number;
  totalPnL: number;
  totalPnLPct: number;
  totalCostBasis: number;
  returnVsCostBasisPct: number;
  dailyReturnPct: number;
  timeWeightedReturnPct: number;
  valueIndex: number;
}

export interface PortfolioPosition {
  id: string;
  symbol?: string;
  name?: string;
  quantity: number;
  costBasis: number;
  realizedPnL: number;
  marketValue?: number;
  unrealizedPnLPct?: number;
  lastSnapshotAt?: string;
  openDate: string;
  closeDate?: string;
  isClosed: boolean;
}

export interface PortfolioGrowthPoint {
  occurredAt: string;
  timeWeightedReturnPct: number;
  returnVsCostBasisPct: number;
}

export interface PortfolioTransactionDto {
  id: string;
  account_id: string;
  origin: "IMPORT" | "MANUAL";
  source: string;
  occurred_at: string;
  type: "BUY" | "SELL" | "DIVIDEND" | "TAX" | "FEE" | "CASH";
  listing_id?: string;
  isin?: string;
  symbol?: string;
  description: string;
  amount: string;
  quantity: string;
  unit_price: string;
  created_at: string;
  updated_at: string;
}

export interface PortfolioTransactionsResponseDto {
  pagination: {
    limit: number;
    offset: number;
    count: number;
    total: number;
  };
  data: PortfolioTransactionDto[];
}

export type PortfolioTransactionSortBy = "date";
export type PortfolioTransactionSortOrder = "asc" | "desc";
export type PortfolioTransactionTypeFilter = "" | "BUY" | "SELL" | "DIVIDEND" | "TAX" | "FEE" | "CASH";
export type PortfolioTransactionOriginFilter = "" | "IMPORT" | "MANUAL";

export interface PortfolioTransaction {
  id: string;
  accountId: string;
  origin: "IMPORT" | "MANUAL";
  source: string;
  occurredAt: string;
  type: "BUY" | "SELL" | "DIVIDEND" | "TAX" | "FEE" | "CASH";
  listingId?: string;
  isin?: string;
  symbol?: string;
  description: string;
  amount: string;
  quantity: string;
  unitPrice: string;
  createdAt: string;
  updatedAt: string;
}

export interface PortfolioTransactionsPagination {
  limit: number;
  offset: number;
  count: number;
  total: number;
}

export interface PortfolioTransactionsResponse {
  pagination: PortfolioTransactionsPagination;
  data: PortfolioTransaction[];
}

export interface PortfolioTransactionsQuery {
  accountId: string;
  from?: string;
  to?: string;
  limit: number;
  offset: number;
  sortBy: PortfolioTransactionSortBy;
  sortOrder: PortfolioTransactionSortOrder;
  q: string;
  type: PortfolioTransactionTypeFilter;
  origin: PortfolioTransactionOriginFilter;
  source: string;
  listing: string;
}

export interface CreateManualPortfolioTransactionPayload {
  account_id: string;
  vendor_id: string;
  occurred_at: string;
  type: "BUY" | "SELL" | "DIVIDEND" | "TAX" | "FEE" | "CASH";
  listing_id?: string;
  amount: string;
  quantity?: string;
  description?: string;
}

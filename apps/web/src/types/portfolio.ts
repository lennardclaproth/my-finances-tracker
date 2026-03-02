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

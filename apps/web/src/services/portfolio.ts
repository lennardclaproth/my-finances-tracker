import { requestJson } from "./http";
import type {
  CreateManualPortfolioTransactionPayload,
  PortfolioPosition,
  PortfolioPositionDto,
  PortfolioPositionsResponseDto,
  PortfolioTransactionsQuery,
  PortfolioTransactionsResponse,
  PortfolioSnapshotPoint,
  PortfolioSnapshotPointDto,
  PortfolioTransaction,
  PortfolioTransactionDto,
  PortfolioTransactionsResponseDto,
} from "../types/portfolio";

function toSnapshotPoint(dto: PortfolioSnapshotPointDto): PortfolioSnapshotPoint {
  return {
    occurredAt: dto.occurred_at,
    marketValue: dto.market_value,
    totalPnL: dto.total_pnl,
    totalPnLPct: dto.total_pnl_pct,
    totalCostBasis: dto.total_cost_basis,
    returnVsCostBasisPct: dto.return_vs_cost_basis_pct,
    dailyReturnPct: dto.daily_return_pct,
    timeWeightedReturnPct: dto.time_weighted_return_pct,
    valueIndex: dto.value_index,
  };
}

function toPosition(dto: PortfolioPositionDto): PortfolioPosition {
  return {
    id: dto.id,
    symbol: dto.symbol,
    name: dto.name,
    quantity: dto.quantity,
    costBasis: dto.cost_basis,
    realizedPnL: dto.realized_pnl,
    marketValue: dto.market_value,
    unrealizedPnLPct: dto.unrealized_pnl_pct,
    lastSnapshotAt: dto.last_snapshot_at,
    openDate: dto.open_date,
    closeDate: dto.close_date,
    isClosed: dto.is_closed,
  };
}

function toPortfolioTransaction(dto: PortfolioTransactionDto): PortfolioTransaction {
  return {
    id: dto.id,
    accountId: dto.account_id,
    origin: dto.origin,
    source: dto.source,
    occurredAt: dto.occurred_at,
    type: dto.type,
    listingId: dto.listing_id,
    isin: dto.isin,
    symbol: dto.symbol,
    description: dto.description,
    amount: dto.amount,
    quantity: dto.quantity,
    unitPrice: dto.unit_price,
    createdAt: dto.created_at,
    updatedAt: dto.updated_at,
  };
}

export async function fetchPortfolioSnapshots(
  accountId: string,
  from?: string,
  to?: string,
): Promise<PortfolioSnapshotPoint[]> {
  const payload = await requestJson<PortfolioSnapshotPointDto[]>("/portfolio/snapshots", {
    method: "GET",
    query: {
      account_id: accountId,
      from,
      to,
    },
  });
  return payload.map(toSnapshotPoint);
}

export async function fetchPortfolioPositions(
  accountId: string,
  includeClosed?: boolean,
): Promise<{ includeClosed: boolean; data: PortfolioPosition[] }> {
  const payload = await requestJson<PortfolioPositionsResponseDto>("/portfolio/positions", {
    method: "GET",
    query: {
      account_id: accountId,
      include_closed: includeClosed ?? false,
    },
  });
  return {
    includeClosed: payload.include_closed,
    data: payload.data.map(toPosition),
  };
}

interface AsyncEventAcceptedResponseDto {
  message_id: string;
  topic: string;
  account_id: string;
}

export async function requestPortfolioRebuild(accountId: string): Promise<{
  messageId: string;
  topic: string;
  accountId: string;
}> {
  const payload = await requestJson<AsyncEventAcceptedResponseDto>("/portfolio/rebuild", {
    method: "POST",
    body: JSON.stringify({
      account_id: accountId,
    }),
  });
  return {
    messageId: payload.message_id,
    topic: payload.topic,
    accountId: payload.account_id,
  };
}

export async function fetchPortfolioTransactions(
  query: PortfolioTransactionsQuery,
): Promise<PortfolioTransactionsResponse> {
  const payload = await requestJson<PortfolioTransactionsResponseDto>("/portfolio/transactions", {
    method: "GET",
    query: {
      account_id: query.accountId,
      from: query.from,
      to: query.to,
      limit: query.limit,
      offset: query.offset,
      sort_by: query.sortBy,
      sort_order: query.sortOrder,
      q: query.q || undefined,
      type: query.type || undefined,
      origin: query.origin || undefined,
      source: query.source || undefined,
      listing: query.listing || undefined,
    },
  });
  return {
    pagination: {
      limit: payload.pagination.limit,
      offset: payload.pagination.offset,
      count: payload.pagination.count,
      total: payload.pagination.total,
    },
    data: payload.data.map(toPortfolioTransaction),
  };
}

export async function createManualPortfolioTransaction(
  input: CreateManualPortfolioTransactionPayload,
): Promise<PortfolioTransaction> {
  const payload = await requestJson<PortfolioTransactionDto>("/portfolio/transactions/manual", {
    method: "POST",
    body: JSON.stringify(input),
  });
  return toPortfolioTransaction(payload);
}

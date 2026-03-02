import { requestJson } from "./http";
import type {
  PortfolioPosition,
  PortfolioPositionDto,
  PortfolioPositionsResponseDto,
  PortfolioSnapshotPoint,
  PortfolioSnapshotPointDto,
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

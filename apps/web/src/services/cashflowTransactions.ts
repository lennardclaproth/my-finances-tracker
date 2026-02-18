import type {
  CashflowAnalyticsQuery,
  CashflowMonthlyAnalyticsResponse,
  CashflowTagFilters,
  CashflowTagDistributionResponse,
  CashflowTransactionsQuery,
  CashflowTransactionsResponse,
  TagTransactionsResponse,
} from "../types/cashflow";
import { requestJson } from "./http";

function toQueryParams(query: CashflowTransactionsQuery): Record<string, string | number | boolean | undefined> {
  return {
    limit: query.limit,
    offset: query.offset,
    sort_by: query.sortBy,
    sort_order: query.sortOrder,
    q: query.q || undefined,
    description: query.description || undefined,
    note: query.note || undefined,
    source: query.source || undefined,
    direction: query.direction || undefined,
    tags: query.tags || undefined,
    untagged: query.untagged || undefined,
    hide_ignored: query.hideIgnored || undefined,
    from: query.from || undefined,
    to: query.to || undefined,
  };
}

function toAnalyticsQueryParams(
  query: CashflowAnalyticsQuery,
): Record<string, string | number | boolean | undefined> {
  return {
    from: query.from,
    to: query.to,
    include_ignored: query.includeIgnored ? true : undefined,
  };
}

export function filtersFromQuery(query: CashflowTransactionsQuery): CashflowTagFilters {
  return {
    ...(query.q ? { q: query.q } : {}),
    ...(query.description ? { description: query.description } : {}),
    ...(query.note ? { note: query.note } : {}),
    ...(query.source ? { source: query.source } : {}),
    ...(query.direction ? { direction: query.direction } : {}),
    ...(query.tags ? { tags: query.tags } : {}),
    ...(query.untagged ? { untagged: true } : {}),
    ...(query.hideIgnored ? { hide_ignored: true } : {}),
    ...(query.from ? { from: query.from } : {}),
    ...(query.to ? { to: query.to } : {}),
  };
}

export async function fetchCashflowTransactions(
  query: CashflowTransactionsQuery,
): Promise<CashflowTransactionsResponse> {
  return requestJson<CashflowTransactionsResponse>("/cashflow/transactions", {
    method: "GET",
    query: toQueryParams(query),
  });
}

export async function fetchCashflowMonthlyAnalytics(
  query: CashflowAnalyticsQuery,
): Promise<CashflowMonthlyAnalyticsResponse> {
  return requestJson<CashflowMonthlyAnalyticsResponse>("/cashflow/analytics/monthly", {
    method: "GET",
    query: toAnalyticsQueryParams(query),
  });
}

export async function fetchCashflowTagDistribution(
  query: CashflowAnalyticsQuery,
): Promise<CashflowTagDistributionResponse> {
  return requestJson<CashflowTagDistributionResponse>("/cashflow/analytics/tags", {
    method: "GET",
    query: toAnalyticsQueryParams(query),
  });
}

export async function tagTransactionsBySelection(
  ids: string[],
  tag: string,
): Promise<TagTransactionsResponse> {
  return requestJson<TagTransactionsResponse>("/cashflow/transactions/tag/selection", {
    method: "POST",
    body: JSON.stringify({ ids, tag }),
  });
}

export async function tagTransactionsByFilter(
  filters: CashflowTagFilters,
  tag: string,
): Promise<TagTransactionsResponse> {
  return requestJson<TagTransactionsResponse>("/cashflow/transactions/tag/filter", {
    method: "POST",
    body: JSON.stringify({ filters, tag }),
  });
}

export async function ignoreTransactionsBySelection(
  ids: string[],
  ignored: boolean,
): Promise<TagTransactionsResponse> {
  return requestJson<TagTransactionsResponse>("/cashflow/transactions/ignore/selection", {
    method: "POST",
    body: JSON.stringify({ ids, ignored }),
  });
}

export async function ignoreTransactionsByFilter(
  filters: CashflowTagFilters,
  ignored: boolean,
): Promise<TagTransactionsResponse> {
  return requestJson<TagTransactionsResponse>("/cashflow/transactions/ignore/filter", {
    method: "POST",
    body: JSON.stringify({ filters, ignored }),
  });
}

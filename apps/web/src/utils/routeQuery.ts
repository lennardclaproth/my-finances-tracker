import type { LocationQuery, LocationQueryRaw } from "vue-router";
import type { CashflowTransactionsQuery, DirectionFilter, SortBy, SortOrder } from "../types/cashflow";
import type {
  PortfolioTransactionOriginFilter,
  PortfolioTransactionsQuery,
  PortfolioTransactionSortBy,
  PortfolioTransactionSortOrder,
  PortfolioTransactionTypeFilter,
} from "../types/portfolio";

export const LIMIT_OPTIONS = [10, 25, 50, 100] as const;

const DEFAULT_LIMIT = 25;
const DEFAULT_OFFSET = 0;
const DEFAULT_SORT_BY: SortBy = "date";
const DEFAULT_SORT_ORDER: SortOrder = "desc";
const SORTABLE_FIELDS: SortBy[] = ["date", "description", "note", "tag", "source", "amount"];
const DEFAULT_PORTFOLIO_SORT_BY: PortfolioTransactionSortBy = "date";
const DEFAULT_PORTFOLIO_SORT_ORDER: PortfolioTransactionSortOrder = "desc";

function firstValue(value: LocationQuery[string]): string {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }
  return value ?? "";
}

function parseLimit(value: string): number {
  const parsed = Number.parseInt(value, 10);
  if (LIMIT_OPTIONS.includes(parsed as (typeof LIMIT_OPTIONS)[number])) {
    return parsed;
  }
  return DEFAULT_LIMIT;
}

function parseOffset(value: string): number {
  const parsed = Number.parseInt(value, 10);
  if (Number.isNaN(parsed) || parsed < 0) {
    return DEFAULT_OFFSET;
  }
  return parsed;
}

function parseSortBy(value: string): SortBy {
  if (SORTABLE_FIELDS.includes(value as SortBy)) {
    return value as SortBy;
  }
  return DEFAULT_SORT_BY;
}

function parseSortOrder(value: string): SortOrder {
  if (value === "asc" || value === "desc") {
    return value;
  }
  return DEFAULT_SORT_ORDER;
}

function parsePortfolioSortBy(value: string): PortfolioTransactionSortBy {
  if (value === "date") {
    return "date";
  }
  return DEFAULT_PORTFOLIO_SORT_BY;
}

function parsePortfolioSortOrder(value: string): PortfolioTransactionSortOrder {
  if (value === "asc" || value === "desc") {
    return value;
  }
  return DEFAULT_PORTFOLIO_SORT_ORDER;
}

function parsePortfolioTransactionType(value: string): PortfolioTransactionTypeFilter {
  const normalized = value.trim().toUpperCase();
  if (normalized === "BUY" || normalized === "SELL" || normalized === "DIVIDEND" || normalized === "TAX" || normalized === "FEE" || normalized === "CASH") {
    return normalized;
  }
  return "";
}

function parsePortfolioTransactionOrigin(value: string): PortfolioTransactionOriginFilter {
  const normalized = value.trim().toUpperCase();
  if (normalized === "IMPORT" || normalized === "MANUAL") {
    return normalized;
  }
  return "";
}

function parseDirection(value: string): DirectionFilter {
  const normalized = value.trim().toLowerCase();
  if (normalized === "in" || normalized === "out") {
    return normalized;
  }
  return "";
}

function cleanString(value: string): string {
  return value.trim();
}

function parseBoolean(value: string): boolean {
  const normalized = value.trim().toLowerCase();
  return normalized === "1" || normalized === "true" || normalized === "yes";
}

export function parseCashflowTransactionsQuery(query: LocationQuery): CashflowTransactionsQuery {
  return {
    limit: parseLimit(firstValue(query.limit)),
    offset: parseOffset(firstValue(query.offset)),
    sortBy: parseSortBy(firstValue(query.sort_by)),
    sortOrder: parseSortOrder(firstValue(query.sort_order)),
    q: cleanString(firstValue(query.q)),
    description: cleanString(firstValue(query.description)),
    note: cleanString(firstValue(query.note)),
    source: cleanString(firstValue(query.source)),
    direction: parseDirection(firstValue(query.direction)),
    tags: cleanString(firstValue(query.tags)),
    untagged: parseBoolean(firstValue(query.untagged)),
    hideIgnored: parseBoolean(firstValue(query.hide_ignored)),
    from: cleanString(firstValue(query.from)),
    to: cleanString(firstValue(query.to)),
  };
}

export function toCashflowTransactionsRouteQuery(state: CashflowTransactionsQuery): LocationQueryRaw {
  return {
    limit: String(state.limit),
    offset: String(state.offset),
    sort_by: state.sortBy,
    sort_order: state.sortOrder,
    ...(state.q ? { q: state.q } : {}),
    ...(state.description ? { description: state.description } : {}),
    ...(state.note ? { note: state.note } : {}),
    ...(state.source ? { source: state.source } : {}),
    ...(state.direction ? { direction: state.direction } : {}),
    ...(state.tags ? { tags: state.tags } : {}),
    ...(state.untagged ? { untagged: "true" } : {}),
    ...(state.hideIgnored ? { hide_ignored: "true" } : {}),
    ...(state.from ? { from: state.from } : {}),
    ...(state.to ? { to: state.to } : {}),
  };
}

export function getFilterFingerprint(state: CashflowTransactionsQuery): string {
  return JSON.stringify({
    q: state.q,
    description: state.description,
    note: state.note,
    source: state.source,
    direction: state.direction,
    tags: state.tags,
    untagged: state.untagged,
    hideIgnored: state.hideIgnored,
    from: state.from,
    to: state.to,
  });
}

export function hasActiveFilters(state: CashflowTransactionsQuery): boolean {
  return Boolean(
    state.q ||
      state.description ||
      state.note ||
      state.source ||
      state.direction ||
      state.tags ||
      state.untagged ||
      state.hideIgnored ||
      state.from ||
      state.to,
  );
}

export function areRouteQueriesEqual(a: LocationQueryRaw, b: LocationQueryRaw): boolean {
  const toParams = (query: LocationQueryRaw): string => {
    const params = new URLSearchParams();
    for (const [key, rawValue] of Object.entries(query)) {
      if (rawValue === undefined || rawValue === null) {
        continue;
      }
      if (Array.isArray(rawValue)) {
        for (const value of rawValue) {
          params.append(key, String(value));
        }
      } else {
        params.set(key, String(rawValue));
      }
    }
    return params.toString();
  };

  return toParams(a) === toParams(b);
}

export interface PortfolioTransactionsRouteQueryState {
  from: string;
  to: string;
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

export const DEFAULT_PORTFOLIO_TRANSACTIONS_QUERY: PortfolioTransactionsRouteQueryState = {
  from: "",
  to: "",
  limit: DEFAULT_LIMIT,
  offset: DEFAULT_OFFSET,
  sortBy: DEFAULT_PORTFOLIO_SORT_BY,
  sortOrder: DEFAULT_PORTFOLIO_SORT_ORDER,
  q: "",
  type: "",
  origin: "",
  source: "",
  listing: "",
};

export function parsePortfolioTransactionsQuery(query: LocationQuery): PortfolioTransactionsRouteQueryState {
  return {
    from: cleanString(firstValue(query.from)),
    to: cleanString(firstValue(query.to)),
    limit: parseLimit(firstValue(query.limit)),
    offset: parseOffset(firstValue(query.offset)),
    sortBy: parsePortfolioSortBy(firstValue(query.sort_by)),
    sortOrder: parsePortfolioSortOrder(firstValue(query.sort_order)),
    q: cleanString(firstValue(query.q)),
    type: parsePortfolioTransactionType(firstValue(query.type)),
    origin: parsePortfolioTransactionOrigin(firstValue(query.origin)),
    source: cleanString(firstValue(query.source)),
    listing: cleanString(firstValue(query.listing)),
  };
}

export function toPortfolioTransactionsRouteQuery(
  state: PortfolioTransactionsRouteQueryState,
  activeTab: "positions" | "transactions",
): LocationQueryRaw {
  return {
    ...(state.from ? { from: state.from } : {}),
    ...(state.to ? { to: state.to } : {}),
    ...(activeTab === "transactions" ? { tab: "transactions" } : {}),
    limit: String(state.limit),
    offset: String(state.offset),
    sort_by: state.sortBy,
    sort_order: state.sortOrder,
    ...(state.q ? { q: state.q } : {}),
    ...(state.type ? { type: state.type } : {}),
    ...(state.origin ? { origin: state.origin } : {}),
    ...(state.source ? { source: state.source } : {}),
    ...(state.listing ? { listing: state.listing } : {}),
  };
}

export function toPortfolioTransactionsServiceQuery(
  state: PortfolioTransactionsRouteQueryState,
  accountId: string,
): PortfolioTransactionsQuery {
  return {
    accountId,
    from: state.from || undefined,
    to: state.to || undefined,
    limit: state.limit,
    offset: state.offset,
    sortBy: state.sortBy,
    sortOrder: state.sortOrder,
    q: state.q,
    type: state.type,
    origin: state.origin,
    source: state.source,
    listing: state.listing,
  };
}

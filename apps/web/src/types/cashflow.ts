export type SortBy = "date" | "description" | "note" | "tag" | "source" | "amount";
export type SortOrder = "asc" | "desc";
export type DirectionFilter = "" | "in" | "out";

export interface CashflowTransaction {
  id: string;
  description: string;
  note: string;
  source: string;
  amountCents: number;
  direction: string;
  date: string;
  tag: string;
  ignored: boolean;
}

export type ManualCashflowTransactionType = "in" | "out";

export interface CreateManualCashflowTransactionInput {
  date: string;
  amount: string;
  type: ManualCashflowTransactionType;
  description: string;
  note: string;
  tag: string;
  vendor?: string;
}

export interface CreateManualCashflowTransactionsPayload {
  account_id: string;
  transactions: CreateManualCashflowTransactionInput[];
}

export interface CreateManualCashflowTransactionsResponse {
  created_count: number;
  data: CashflowTransaction[];
}

export interface Pagination {
  limit: number;
  offset: number;
  count: number;
  total: number;
}

export interface CashflowTransactionsResponse {
  pagination: Pagination;
  data: CashflowTransaction[];
}

export interface TagTransactionsResponse {
  updated_count: number;
  status: string;
}

export interface CashflowTagFilters {
  q?: string;
  description?: string;
  note?: string;
  source?: string;
  direction?: Exclude<DirectionFilter, "">;
  tags?: string;
  untagged?: boolean;
  hide_ignored?: boolean;
  from?: string;
  to?: string;
}

export interface CashflowTransactionsQuery {
  limit: number;
  offset: number;
  sortBy: SortBy;
  sortOrder: SortOrder;
  q: string;
  description: string;
  note: string;
  source: string;
  direction: DirectionFilter;
  tags: string;
  untagged: boolean;
  hideIgnored: boolean;
  from: string;
  to: string;
}

export interface CashflowAnalyticsQuery {
  from?: string;
  to?: string;
  includeIgnored?: boolean;
}

export interface CashflowMonthlyAnalyticsPoint {
  month: string;
  incomingCents: number;
  outgoingCents: number;
  netCents: number;
}

export interface CashflowMonthlyAnalyticsResponse {
  data: CashflowMonthlyAnalyticsPoint[];
}

export interface CashflowTagDistributionEntry {
  tag: string;
  totalCents: number;
}

export interface CashflowTagDistributionResponse {
  combined: CashflowTagDistributionEntry[];
  incoming: CashflowTagDistributionEntry[];
  outgoing: CashflowTagDistributionEntry[];
}

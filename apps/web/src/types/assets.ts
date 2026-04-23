export type AssetClassSource = "MANUAL" | "PORTFOLIO";
export type AssetChangeType = "SET" | "ADJUST";
export type AssetChangeDirection = "INCREASE" | "DECREASE";

export interface AssetClassDto {
  id: string;
  name: string;
  source: AssetClassSource;
  archived: boolean;
  current_worth: string;
  last_change_at?: string;
  growth_pct?: number;
  updated_at: string;
}

export interface AssetItemDto {
  id: string;
  name: string;
  current_worth: string;
  archived: boolean;
  updated_at: string;
}

export interface AssetGrowthPointDto {
  date: string;
  total_worth: string;
}

export interface AssetHistoryDto {
  id: string;
  item_id: string;
  change_type: AssetChangeType;
  direction?: AssetChangeDirection;
  amount: string;
  previous_worth: string;
  new_worth: string;
  class_total_worth: string;
  effective_date: string;
  note: string;
  created_at: string;
}

export interface AssetClassDetailsDto {
  class: AssetClassDto;
  items: AssetItemDto[];
  growth: AssetGrowthPointDto[];
  history: AssetHistoryDto[];
}

export interface AssetClass {
  id: string;
  name: string;
  source: AssetClassSource;
  archived: boolean;
  currentWorth: string;
  lastChangeAt?: string;
  growthPct?: number;
  updatedAt: string;
}

export interface AssetItem {
  id: string;
  name: string;
  currentWorth: string;
  archived: boolean;
  updatedAt: string;
}

export interface AssetGrowthPoint {
  date: string;
  totalWorth: string;
}

export interface AssetHistoryEntry {
  id: string;
  itemId: string;
  changeType: AssetChangeType;
  direction?: AssetChangeDirection;
  amount: string;
  previousWorth: string;
  newWorth: string;
  classTotalWorth: string;
  effectiveDate: string;
  note: string;
  createdAt: string;
}

export interface AssetClassDetails {
  classRow: AssetClass;
  items: AssetItem[];
  growth: AssetGrowthPoint[];
  history: AssetHistoryEntry[];
}

export interface CreateAssetItemInput {
  accountId: string;
  classId: string;
  name: string;
  initialWorth: string;
  effectiveDate: string;
  note?: string;
}

export interface SetAssetWorthInput {
  accountId: string;
  classId: string;
  itemId: string;
  worth: string;
  effectiveDate: string;
  note?: string;
}

export interface AdjustAssetWorthInput {
  accountId: string;
  classId: string;
  itemId: string;
  direction: "increase" | "decrease";
  amount: string;
  effectiveDate: string;
  note?: string;
}

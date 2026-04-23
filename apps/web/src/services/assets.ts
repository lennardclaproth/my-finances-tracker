import { requestJson } from "./http";
import type {
  AdjustAssetWorthInput,
  AssetClass,
  AssetClassDetails,
  AssetClassDetailsDto,
  AssetClassDto,
  AssetGrowthPoint,
  AssetGrowthPointDto,
  AssetHistoryDto,
  AssetHistoryEntry,
  AssetItem,
  AssetItemDto,
  CreateAssetItemInput,
  SetAssetWorthInput,
} from "../types/assets";

function toAssetClass(dto: AssetClassDto): AssetClass {
  return {
    id: dto.id,
    name: dto.name,
    source: dto.source,
    archived: dto.archived,
    currentWorth: dto.current_worth,
    lastChangeAt: dto.last_change_at,
    growthPct: dto.growth_pct,
    updatedAt: dto.updated_at,
  };
}

function toAssetItem(dto: AssetItemDto): AssetItem {
  return {
    id: dto.id,
    name: dto.name,
    currentWorth: dto.current_worth,
    archived: dto.archived,
    updatedAt: dto.updated_at,
  };
}

function toAssetGrowthPoint(dto: AssetGrowthPointDto): AssetGrowthPoint {
  return {
    date: dto.date,
    totalWorth: dto.total_worth,
  };
}

function toAssetHistoryEntry(dto: AssetHistoryDto): AssetHistoryEntry {
  return {
    id: dto.id,
    itemId: dto.item_id,
    changeType: dto.change_type,
    direction: dto.direction,
    amount: dto.amount,
    previousWorth: dto.previous_worth,
    newWorth: dto.new_worth,
    classTotalWorth: dto.class_total_worth,
    effectiveDate: dto.effective_date,
    note: dto.note,
    createdAt: dto.created_at,
  };
}

function toAssetClassDetails(dto: AssetClassDetailsDto): AssetClassDetails {
  return {
    classRow: toAssetClass(dto.class),
    items: dto.items.map(toAssetItem),
    growth: dto.growth.map(toAssetGrowthPoint),
    history: dto.history.map(toAssetHistoryEntry),
  };
}

export async function fetchAssetClasses(accountId: string, includeArchived = false): Promise<AssetClass[]> {
  const payload = await requestJson<AssetClassDto[]>("/assets/classes", {
    method: "GET",
    query: {
      account_id: accountId,
      include_archived: includeArchived ? true : undefined,
    },
  });
  return payload.map(toAssetClass);
}

export async function fetchAssetClassDetails(accountId: string, classId: string): Promise<AssetClassDetails> {
  const payload = await requestJson<AssetClassDetailsDto>(`/assets/classes/${classId}`, {
    method: "GET",
    query: {
      account_id: accountId,
    },
  });
  return toAssetClassDetails(payload);
}

export async function fetchAssetSnapshots(accountId: string, from?: string, to?: string): Promise<AssetGrowthPoint[]> {
  const payload = await requestJson<AssetGrowthPointDto[]>("/assets/snapshots", {
    method: "GET",
    query: {
      account_id: accountId,
      from,
      to,
    },
  });
  return payload.map(toAssetGrowthPoint);
}

export async function createAssetClass(accountId: string, name: string): Promise<AssetClass> {
  const payload = await requestJson<AssetClassDto>("/assets/classes", {
    method: "POST",
    body: JSON.stringify({
      account_id: accountId,
      name,
    }),
  });
  return toAssetClass(payload);
}

export async function updateAssetClass(input: {
  accountId: string;
  classId: string;
  name?: string;
  archived?: boolean;
}): Promise<void> {
  await requestJson<{ status: string }>("/assets/classes", {
    method: "PATCH",
    body: JSON.stringify({
      account_id: input.accountId,
      id: input.classId,
      ...(input.name !== undefined ? { name: input.name } : {}),
      ...(input.archived !== undefined ? { archived: input.archived } : {}),
    }),
  });
}

export async function deleteAssetClass(accountId: string, classId: string): Promise<void> {
  await requestJson<{ status: string }>(`/assets/classes/${classId}`, {
    method: "DELETE",
    query: {
      account_id: accountId,
    },
  });
}

export async function createAssetItem(input: CreateAssetItemInput): Promise<AssetItem> {
  const payload = await requestJson<AssetItemDto>("/assets/items", {
    method: "POST",
    body: JSON.stringify({
      account_id: input.accountId,
      class_id: input.classId,
      name: input.name,
      initial_worth: input.initialWorth,
      effective_date: input.effectiveDate,
      ...(input.note ? { note: input.note } : {}),
    }),
  });
  return toAssetItem(payload);
}

export async function setAssetItemWorth(input: SetAssetWorthInput): Promise<void> {
  await requestJson<{ status: string }>("/assets/items/worth/set", {
    method: "POST",
    body: JSON.stringify({
      account_id: input.accountId,
      class_id: input.classId,
      item_id: input.itemId,
      worth: input.worth,
      effective_date: input.effectiveDate,
      ...(input.note ? { note: input.note } : {}),
    }),
  });
}

export async function adjustAssetItemWorth(input: AdjustAssetWorthInput): Promise<void> {
  await requestJson<{ status: string }>("/assets/items/worth/adjust", {
    method: "POST",
    body: JSON.stringify({
      account_id: input.accountId,
      class_id: input.classId,
      item_id: input.itemId,
      direction: input.direction,
      amount: input.amount,
      effective_date: input.effectiveDate,
      ...(input.note ? { note: input.note } : {}),
    }),
  });
}

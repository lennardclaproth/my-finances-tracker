<script setup lang="ts">
import { PlusIcon } from "@heroicons/vue/24/outline";
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import IconButton from "../components/atoms/IconButton.vue";
import CreateAssetClassModal from "../components/molecules/CreateAssetClassModal.vue";
import EditAssetClassModal from "../components/molecules/EditAssetClassModal.vue";
import AssetDistributionDonutChart from "../components/molecules/charts/AssetDistributionDonutChart.vue";
import AssetGrowthLineChart from "../components/molecules/charts/AssetGrowthLineChart.vue";
import ToastMessage from "../components/molecules/ToastMessage.vue";
import AssetClassDrawer from "../components/organisms/AssetClassDrawer.vue";
import AssetClassesTable from "../components/organisms/AssetClassesTable.vue";
import TopNavbar from "../components/organisms/TopNavbar.vue";
import AppShellTemplate from "../components/templates/AppShellTemplate.vue";
import { useAppSession } from "../composables/useAppSession";
import { ApiError } from "../services/http";
import { getRealtimeClient, type DataChangedMessage } from "../services/realtime";
import {
  adjustAssetItemWorth,
  createAssetClass,
  createAssetItem,
  deleteAssetClass,
  fetchAssetClassDetails,
  fetchAssetClasses,
  fetchAssetSnapshots,
  setAssetItemWorth,
  updateAssetClass,
} from "../services/assets";
import type { AssetClass, AssetClassDetails, AssetGrowthPoint } from "../types/assets";
import { areRouteQueriesEqual } from "../utils/routeQuery";

type AlertTone = "success" | "danger" | "info";

const session = useAppSession();
const realtimeClient = getRealtimeClient();
const route = useRoute();
const router = useRouter();

const rows = ref<AssetClass[]>([]);
const rowsLoading = ref(false);
const rowsError = ref("");

const selectedClassId = ref<string | null>(null);
const drawerOpen = ref(false);
const details = ref<AssetClassDetails | null>(null);
const detailsLoading = ref(false);
const detailsError = ref("");
const mutationBusy = ref(false);

const createClassOpen = ref(false);
const createClassBusy = ref(false);
const editClassOpen = ref(false);
const editClassRow = ref<AssetClass | null>(null);

const totalGrowthPoints = ref<AssetGrowthPoint[]>([]);
const distributionData = ref<{ label: string; value: number }[]>([]);
const analyticsLoading = ref(false);
const analyticsError = ref("");
let activeAnalyticsRequestID = 0;
let realtimeRefreshTimer: ReturnType<typeof setTimeout> | null = null;
let unsubscribeRealtime: (() => void) | null = null;

const toast = ref<{ tone: AlertTone; message: string } | null>(null);
let toastTimer: ReturnType<typeof setTimeout> | null = null;
const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 0,
});

const from = computed(() => firstQueryValue(route.query.from as string | string[] | undefined).trim());
const to = computed(() => firstQueryValue(route.query.to as string | string[] | undefined).trim());

const currentTotalWorthLabel = computed(() => {
  const latest = latestWorthFromSeries(totalGrowthPoints.value);
  return currencyFormatter.format(latest ?? 0);
});

const filteredDetails = computed<AssetClassDetails | null>(() => {
  if (!details.value) {
    return null;
  }
  return {
    ...details.value,
    growth: applyDateRange(details.value.growth, from.value, to.value),
    history: filterHistoryByRange(details.value.history, from.value, to.value),
  };
});

function toErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "An unexpected error occurred.";
}

function showToast(tone: AlertTone, message: string): void {
  toast.value = { tone, message };
  if (toastTimer) {
    clearTimeout(toastTimer);
  }
  toastTimer = setTimeout(() => {
    toast.value = null;
    toastTimer = null;
  }, 5000);
}

async function loadClasses(): Promise<void> {
  rowsLoading.value = true;
  rowsError.value = "";
  try {
    rows.value = await fetchAssetClasses(session.activeAccountId.value, false);
    distributionData.value = toDistributionData(rows.value);
    void loadAnalytics();
    if (selectedClassId.value && !rows.value.some((row) => row.id === selectedClassId.value)) {
      selectedClassId.value = null;
      details.value = null;
      drawerOpen.value = false;
    }
  } catch (error: unknown) {
    rowsError.value = toErrorMessage(error);
    distributionData.value = [];
  } finally {
    rowsLoading.value = false;
  }
}

function parseWorth(value: string): number {
  const parsed = Number.parseFloat(value);
  if (Number.isNaN(parsed)) {
    return 0;
  }
  return parsed;
}

function firstQueryValue(value: string | string[] | null | undefined): string {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }
  return value ?? "";
}

function isISODate(value: string): boolean {
  return /^\d{4}-\d{2}-\d{2}$/.test(value);
}

function latestWorthFromSeries(points: AssetGrowthPoint[]): number | null {
  if (points.length === 0) {
    return null;
  }
  const sorted = [...points].sort((a, b) => a.date.localeCompare(b.date));
  const latest = sorted[sorted.length - 1];
  const parsed = Number.parseFloat(latest.totalWorth);
  if (Number.isNaN(parsed)) {
    return null;
  }
  return parsed;
}

function applyDateRange(points: AssetGrowthPoint[], fromValue: string, toValue: string): AssetGrowthPoint[] {
  const sorted = [...points].sort((a, b) => a.date.localeCompare(b.date));
  const hasFrom = isISODate(fromValue);
  const hasTo = isISODate(toValue);
  if (!hasFrom && !hasTo) {
    return sorted;
  }
  const lowerBound = hasFrom ? fromValue : "0000-01-01";
  const upperBound = hasTo ? toValue : "9999-12-31";
  if (lowerBound > upperBound) {
    return [];
  }

  let latestOnOrBeforeFrom: AssetGrowthPoint | null = null;
  let latestOnOrBeforeTo: AssetGrowthPoint | null = null;
  const inRange: AssetGrowthPoint[] = [];
  for (const point of sorted) {
    if (point.date <= lowerBound) {
      latestOnOrBeforeFrom = point;
    }
    if (point.date <= upperBound) {
      latestOnOrBeforeTo = point;
    }
    if (point.date >= lowerBound && point.date <= upperBound) {
      inRange.push(point);
    }
  }

  if (hasFrom && latestOnOrBeforeFrom && inRange[0]?.date !== lowerBound) {
    inRange.unshift({
      date: lowerBound,
      totalWorth: latestOnOrBeforeFrom.totalWorth,
    });
  }

  if (hasTo && latestOnOrBeforeTo) {
    const last = inRange[inRange.length - 1];
    if (!last || last.date !== upperBound) {
      inRange.push({
        date: upperBound,
        totalWorth: latestOnOrBeforeTo.totalWorth,
      });
    }
  }

  return inRange;
}

function filterHistoryByRange(history: AssetClassDetails["history"], fromValue: string, toValue: string): AssetClassDetails["history"] {
  const hasFrom = isISODate(fromValue);
  const hasTo = isISODate(toValue);
  if (!hasFrom && !hasTo) {
    return history;
  }
  const lowerBound = hasFrom ? fromValue : "0000-01-01";
  const upperBound = hasTo ? toValue : "9999-12-31";
  if (lowerBound > upperBound) {
    return [];
  }
  return history.filter((entry) => entry.effectiveDate >= lowerBound && entry.effectiveDate <= upperBound);
}

function toDistributionData(classRows: AssetClass[]): { label: string; value: number }[] {
  return classRows
    .map((row) => ({
      label: row.name,
      value: parseWorth(row.currentWorth),
    }))
    .filter((entry) => entry.value > 0);
}

async function updateRouteDateQuery(nextFrom: string, nextTo: string): Promise<void> {
  const nextQuery = { ...route.query } as Record<string, string | string[] | undefined>;
  const trimmedFrom = nextFrom.trim();
  const trimmedTo = nextTo.trim();
  if (trimmedFrom !== "") {
    nextQuery.from = trimmedFrom;
  } else {
    delete nextQuery.from;
  }
  if (trimmedTo !== "") {
    nextQuery.to = trimmedTo;
  } else {
    delete nextQuery.to;
  }

  if (areRouteQueriesEqual(route.query, nextQuery)) {
    return;
  }
  await router.push({
    path: route.path,
    query: nextQuery,
  });
}

async function onDateApply(nextFrom: string, nextTo: string): Promise<void> {
  await updateRouteDateQuery(nextFrom, nextTo);
}

async function onDateClear(): Promise<void> {
  await updateRouteDateQuery("", "");
}

function onGrowthRangeSelected(range: { from: string; to: string }): void {
  void onDateApply(range.from, range.to);
}

async function loadAnalytics(): Promise<void> {
  const requestID = ++activeAnalyticsRequestID;
  analyticsLoading.value = true;
  analyticsError.value = "";
  try {
    totalGrowthPoints.value = await fetchAssetSnapshots(
      session.activeAccountId.value,
      from.value || undefined,
      to.value || undefined,
    );
  } catch (error: unknown) {
    if (requestID !== activeAnalyticsRequestID) {
      return;
    }
    analyticsError.value = toErrorMessage(error);
    totalGrowthPoints.value = [];
  } finally {
    if (requestID === activeAnalyticsRequestID) {
      analyticsLoading.value = false;
    }
  }
}

function onRealtimeMessage(message: DataChangedMessage): void {
  if (message.account_id !== session.activeAccountId.value) {
    return;
  }
  if (message.event !== "assets.rebuilt") {
    return;
  }
  scheduleRealtimeRefresh();
}

function scheduleRealtimeRefresh(): void {
  if (realtimeRefreshTimer) {
    return;
  }
  realtimeRefreshTimer = setTimeout(() => {
    realtimeRefreshTimer = null;
    void refreshAfterRealtimeEvent();
  }, 250);
}

async function refreshAfterRealtimeEvent(): Promise<void> {
  await loadClasses();
  if (drawerOpen.value && selectedClassId.value) {
    await loadDetails(selectedClassId.value);
  }
}

async function loadDetails(classId: string): Promise<void> {
  detailsLoading.value = true;
  detailsError.value = "";
  try {
    details.value = await fetchAssetClassDetails(session.activeAccountId.value, classId);
  } catch (error: unknown) {
    detailsError.value = toErrorMessage(error);
  } finally {
    detailsLoading.value = false;
  }
}

async function onSelectRow(row: AssetClass): Promise<void> {
  selectedClassId.value = row.id;
  drawerOpen.value = true;
  await loadDetails(row.id);
}

function closeDrawer(): void {
  drawerOpen.value = false;
}

function onEditClass(row: AssetClass): void {
  editClassRow.value = row;
  editClassOpen.value = true;
}

async function onCreateClass(name: string): Promise<void> {
  if (createClassBusy.value) {
    return;
  }
  createClassBusy.value = true;
  try {
    const created = await createAssetClass(session.activeAccountId.value, name);
    createClassOpen.value = false;
    showToast("success", `Created asset class "${created.name}".`);
    await loadClasses();
    selectedClassId.value = created.id;
    drawerOpen.value = true;
    await loadDetails(created.id);
  } catch (error: unknown) {
    showToast("danger", toErrorMessage(error));
  } finally {
    createClassBusy.value = false;
  }
}

async function onSaveClassSettings(payload: { name: string; archived: boolean }): Promise<void> {
  if (!editClassRow.value) {
    return;
  }
  await mutateWithReload(
    () =>
      updateAssetClass({
        accountId: session.activeAccountId.value,
        classId: editClassRow.value!.id,
        name: payload.name,
        archived: payload.archived,
      }),
    "Asset class updated.",
  );
  editClassOpen.value = false;
}

async function onDeleteClassFromModal(): Promise<void> {
  if (!editClassRow.value) {
    return;
  }
  const classID = editClassRow.value.id;
  await mutateWithReload(
    () => deleteAssetClass(session.activeAccountId.value, classID),
    "Asset class deleted.",
    selectedClassId.value === classID,
  );
  editClassOpen.value = false;
  editClassRow.value = null;
}

async function mutateWithReload(
  run: () => Promise<void>,
  successMessage: string,
  closeDrawerAfterSuccess = false,
): Promise<void> {
  if (mutationBusy.value) {
    return;
  }
  mutationBusy.value = true;
  try {
    await run();
    showToast("success", successMessage);
    await loadClasses();
    if (closeDrawerAfterSuccess) {
      selectedClassId.value = null;
      details.value = null;
      drawerOpen.value = false;
      return;
    }
    if (selectedClassId.value) {
      await loadDetails(selectedClassId.value);
    }
  } catch (error: unknown) {
    showToast("danger", toErrorMessage(error));
  } finally {
    mutationBusy.value = false;
  }
}

async function onCreateItem(payload: { name: string; initialWorth: string; effectiveDate: string; note?: string }): Promise<void> {
  if (!selectedClassId.value) {
    return;
  }
  await mutateWithReload(
    () =>
      createAssetItem({
        accountId: session.activeAccountId.value,
        classId: selectedClassId.value as string,
        name: payload.name,
        initialWorth: payload.initialWorth,
        effectiveDate: payload.effectiveDate,
        note: payload.note,
      }).then(() => undefined),
    "Asset created.",
  );
}

async function onSetWorth(payload: { itemId: string; worth: string; effectiveDate: string; note?: string }): Promise<void> {
  if (!selectedClassId.value) {
    return;
  }
  await mutateWithReload(
    () =>
      setAssetItemWorth({
        accountId: session.activeAccountId.value,
        classId: selectedClassId.value as string,
        itemId: payload.itemId,
        worth: payload.worth,
        effectiveDate: payload.effectiveDate,
        note: payload.note,
      }),
    "Worth updated.",
  );
}

async function onAdjustWorth(payload: {
  itemId: string;
  direction: "increase" | "decrease";
  amount: string;
  effectiveDate: string;
  note?: string;
}): Promise<void> {
  if (!selectedClassId.value) {
    return;
  }
  await mutateWithReload(
    () =>
      adjustAssetItemWorth({
        accountId: session.activeAccountId.value,
        classId: selectedClassId.value as string,
        itemId: payload.itemId,
        direction: payload.direction,
        amount: payload.amount,
        effectiveDate: payload.effectiveDate,
        note: payload.note,
      }),
    "Worth adjustment applied.",
  );
}

watch(
  () => session.activeAccountId.value,
  (accountID, previousAccountID) => {
    if (accountID === previousAccountID) {
      return;
    }
    realtimeClient.setAccountId(accountID);
    selectedClassId.value = null;
    details.value = null;
    drawerOpen.value = false;
    void loadClasses();
  },
);

watch(
  () => [from.value, to.value] as const,
  () => {
    void loadAnalytics();
  },
);

onMounted(() => {
  realtimeClient.setAccountId(session.activeAccountId.value);
  unsubscribeRealtime = realtimeClient.subscribe(onRealtimeMessage);
  void loadClasses();
});

onBeforeUnmount(() => {
  if (toastTimer) {
    clearTimeout(toastTimer);
    toastTimer = null;
  }
  if (realtimeRefreshTimer) {
    clearTimeout(realtimeRefreshTimer);
    realtimeRefreshTimer = null;
  }
  if (unsubscribeRealtime) {
    unsubscribeRealtime();
    unsubscribeRealtime = null;
  }
});
</script>

<template>
  <AppShellTemplate>
    <template #top>
      <TopNavbar
        :from="from"
        :to="to"
        :show-filter-controls="true"
        :show-search-control="false"
        :show-date-control="true"
        @date-apply="onDateApply"
        @date-clear="onDateClear"
      />
    </template>

    <div class="flex h-full min-h-0 flex-col gap-3 px-4 pb-4">
      <section class="shrink-0 rounded-3xl border border-slate-300 bg-white p-4 shadow-sm">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 class="font-secondary text-xl font-semibold text-slate-800 md:text-2xl">Asset Management</h1>
          </div>
        </div>
        <div class="mt-4 grid grid-cols-1 gap-3 lg:grid-cols-3">
          <article class="rounded-2xl border border-slate-200 bg-slate-50 p-3 lg:col-span-2">
            <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
              <h2 class="text-sm font-semibold uppercase tracking-wide text-slate-600">Assets Growth</h2>
              <div class="flex flex-wrap items-center gap-2">
                <span class="inline-flex items-center rounded-full border border-slate-200 bg-white px-2 py-0.5 text-[11px] font-medium text-slate-700">
                  Current Worth: {{ currentTotalWorthLabel }}
                </span>
              </div>
            </div>
            <div class="h-52">
              <AssetGrowthLineChart
                :loading="analyticsLoading"
                :data="totalGrowthPoints"
                series-label="Total Assets Worth"
                @range-selected="onGrowthRangeSelected"
              />
            </div>
          </article>
          <article class="rounded-2xl border border-slate-200 bg-slate-50 p-3">
            <h2 class="mb-2 text-sm font-semibold uppercase tracking-wide text-slate-600">Class Distribution</h2>
            <div class="h-52">
              <AssetDistributionDonutChart :loading="analyticsLoading" :data="distributionData" />
            </div>
          </article>
        </div>
        <p v-if="analyticsError" class="mt-2 text-sm text-rose-700">{{ analyticsError }}</p>
      </section>

      <div class="relative min-h-0 flex-1">
        <AssetClassesTable
          class="h-full"
          :rows="rows"
          :loading="rowsLoading"
          :selected-class-id="selectedClassId"
          :error-message="rowsError"
          @select="void onSelectRow($event)"
          @edit="onEditClass"
          @retry="void loadClasses()"
        />
        <div class="absolute bottom-6 right-6 z-40">
          <IconButton
            tone="primary"
            size="fab"
            title="Create asset class"
            :disabled="createClassBusy || rowsLoading"
            @click="createClassOpen = true"
          >
            <PlusIcon class="h-6 w-6" />
          </IconButton>
        </div>
      </div>
    </div>

    <AssetClassDrawer
      :open="drawerOpen"
      :details="filteredDetails"
      :loading="detailsLoading"
      :busy="mutationBusy"
      :error-message="detailsError"
      @close="closeDrawer"
      @create-item="void onCreateItem($event)"
      @set-worth="void onSetWorth($event)"
      @adjust-worth="void onAdjustWorth($event)"
    />

    <CreateAssetClassModal
      :open="createClassOpen"
      :loading="createClassBusy"
      @close="createClassOpen = false"
      @submit="void onCreateClass($event)"
    />
    <EditAssetClassModal
      :open="editClassOpen"
      :row="editClassRow"
      :loading="mutationBusy"
      @close="editClassOpen = false"
      @save="void onSaveClassSettings($event)"
      @delete="void onDeleteClassFromModal()"
    />

    <div class="pointer-events-none fixed right-4 top-4 z-50">
      <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="translate-y-2 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-y-2 opacity-0"
      >
        <ToastMessage
          v-if="toast"
          :tone="toast.tone"
          :message="toast.message"
          @close="toast = null"
        />
      </Transition>
    </div>
  </AppShellTemplate>
</template>


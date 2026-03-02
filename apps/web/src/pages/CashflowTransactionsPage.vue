<script setup lang="ts">
import { ArrowsPointingOutIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import ToastMessage from "../components/molecules/ToastMessage.vue";
import TagDonutChart from "../components/molecules/charts/TagDonutChart.vue";
import TrendLineChart from "../components/molecules/charts/TrendLineChart.vue";
import TagModal from "../components/organisms/TagModal.vue";
import TopNavbar from "../components/organisms/TopNavbar.vue";
import TransactionsFooterBar from "../components/organisms/TransactionsFooterBar.vue";
import TransactionsTable from "../components/organisms/TransactionsTable.vue";
import AppShellTemplate from "../components/templates/AppShellTemplate.vue";
import {
  fetchCashflowMonthlyAnalytics,
  fetchCashflowTagDistribution,
  fetchCashflowTransactions,
  filtersFromQuery,
  ignoreTransactionsByFilter,
  ignoreTransactionsBySelection,
  tagTransactionsByFilter,
  tagTransactionsBySelection,
} from "../services/cashflowTransactions";
import { ApiError } from "../services/http";
import type {
  CashflowTagDistributionEntry,
  CashflowTransaction,
  CashflowTransactionsQuery,
  CashflowMonthlyAnalyticsPoint,
  SortBy,
} from "../types/cashflow";
import {
  areRouteQueriesEqual,
  getFilterFingerprint,
  hasActiveFilters,
  parseCashflowTransactionsQuery,
  toCashflowTransactionsRouteQuery,
} from "../utils/routeQuery";

type AlertTone = "success" | "danger" | "info";

interface AlertState {
  id: number;
  tone: AlertTone;
  message: string;
}

interface ColumnFilterDraft {
  description: string;
  note: string;
  source: string;
  direction: "" | "in" | "out";
  tags: string;
  untagged: boolean;
}

type ExpandedChart = "trend" | "incoming" | "outgoing" | null;

const route = useRoute();
const router = useRouter();

const rows = ref<CashflowTransaction[]>([]);
const monthlyTrend = ref<CashflowMonthlyAnalyticsPoint[]>([]);
const incomingTagDistribution = ref<CashflowTagDistributionEntry[]>([]);
const outgoingTagDistribution = ref<CashflowTagDistributionEntry[]>([]);
const pagination = ref({
  limit: 25,
  offset: 0,
  count: 0,
  total: 0,
});

const loading = ref(false);
const analyticsLoading = ref(false);
const mutating = ref(false);
const alert = ref<AlertState | null>(null);
const selectedIds = ref<string[]>([]);
const allMatchingSelected = ref(false);
const searchDraft = ref("");
const columnFilterDraft = ref<ColumnFilterDraft>({
  description: "",
  note: "",
  source: "",
  direction: "",
  tags: "",
  untagged: false,
});
const tagModalOpen = ref(false);
const expandedChart = ref<ExpandedChart>(null);

let activeRequestId = 0;
let activeAnalyticsRequestId = 0;
let columnFilterDebounceHandle: ReturnType<typeof setTimeout> | null = null;
let toastTimeoutHandle: ReturnType<typeof setTimeout> | null = null;
let toastIdCounter = 0;

const queryState = computed(() => parseCashflowTransactionsQuery(route.query));
const selectedCount = computed(() =>
  allMatchingSelected.value ? pagination.value.total : selectedIds.value.length,
);
const actionsEnabled = computed(() => selectedIds.value.length > 0 || allMatchingSelected.value);
const isBusy = computed(() => loading.value || mutating.value);
const hasFilters = computed(() => hasActiveFilters(queryState.value));
const expandedChartTitle = computed(() => {
  if (expandedChart.value === "trend") {
    return "Cashflow";
  }
  if (expandedChart.value === "incoming") {
    return "Incoming Tags";
  }
  if (expandedChart.value === "outgoing") {
    return "Outgoing Tags";
  }
  return "";
});
const tagTargetLabel = computed(() => {
  if (selectedIds.value.length > 0) {
    return `Selected transactions (${selectedCount.value})`;
  }
  if (allMatchingSelected.value) {
    return `All matching transactions (${selectedCount.value})`;
  }
  if (hasFilters.value) {
    return "Filtered transactions";
  }
  return "All transactions";
});

onMounted(() => {
  const normalizedQuery = toCashflowTransactionsRouteQuery(queryState.value);
  if (!areRouteQueriesEqual(route.query, normalizedQuery)) {
      void router.replace({
      path: "/cashflow",
      query: normalizedQuery,
    });
  }
});

onBeforeUnmount(() => {
  if (columnFilterDebounceHandle) {
    clearTimeout(columnFilterDebounceHandle);
  }
  if (toastTimeoutHandle) {
    clearTimeout(toastTimeoutHandle);
  }
});

watch(
  () => queryState.value,
  (next, previous) => {
    if (previous && getFilterFingerprint(previous) !== getFilterFingerprint(next)) {
      selectedIds.value = [];
      allMatchingSelected.value = false;
    }

    searchDraft.value = next.q;
    columnFilterDraft.value = {
      description: next.description,
      note: next.note,
      source: next.source,
      direction: next.direction,
      tags: next.tags,
      untagged: next.untagged,
    };

    void loadTransactions(next);
    void loadAnalytics(next);
  },
  { immediate: true, deep: true },
);

async function loadTransactions(query: CashflowTransactionsQuery): Promise<void> {
  const requestId = ++activeRequestId;
  loading.value = true;

  try {
    const response = await fetchCashflowTransactions(query);
    if (requestId !== activeRequestId) {
      return;
    }

    rows.value = response.data;
    pagination.value = response.pagination;
  } catch (error: unknown) {
    if (requestId !== activeRequestId) {
      return;
    }
    showAlert("danger", toErrorMessage(error));
  } finally {
    if (requestId === activeRequestId) {
      loading.value = false;
    }
  }
}

async function loadAnalytics(query: CashflowTransactionsQuery): Promise<void> {
  const requestId = ++activeAnalyticsRequestId;
  analyticsLoading.value = true;

  try {
    const [trendResponse, tagDistributionResponse] = await Promise.all([
      fetchCashflowMonthlyAnalytics({
        from: query.from || undefined,
        to: query.to || undefined,
        includeIgnored: false,
      }),
      fetchCashflowTagDistribution({
        from: query.from || undefined,
        to: query.to || undefined,
        includeIgnored: false,
      }),
    ]);

    if (requestId !== activeAnalyticsRequestId) {
      return;
    }

    monthlyTrend.value = trendResponse.data;
    incomingTagDistribution.value = tagDistributionResponse.incoming;
    outgoingTagDistribution.value = tagDistributionResponse.outgoing;
  } catch (error: unknown) {
    if (requestId !== activeAnalyticsRequestId) {
      return;
    }
    showAlert("danger", toErrorMessage(error));
  } finally {
    if (requestId === activeAnalyticsRequestId) {
      analyticsLoading.value = false;
    }
  }
}

function toErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "An unexpected error occurred.";
}

function showAlert(tone: AlertTone, message: string): void {
  const id = ++toastIdCounter;
  alert.value = { id, tone, message };

  if (toastTimeoutHandle) {
    clearTimeout(toastTimeoutHandle);
  }

  toastTimeoutHandle = window.setTimeout(() => {
    if (alert.value?.id === id) {
      alert.value = null;
    }
    toastTimeoutHandle = null;
  }, 5000);
}

function closeAlert(): void {
  alert.value = null;
  if (toastTimeoutHandle) {
    clearTimeout(toastTimeoutHandle);
    toastTimeoutHandle = null;
  }
}

function openExpandedChart(chart: Exclude<ExpandedChart, null>): void {
  expandedChart.value = chart;
}

function closeExpandedChart(): void {
  expandedChart.value = null;
}

function showMutationStatus(status: string | undefined, fallbackMessage: string): void {
  const message = (status || fallbackMessage).trim();
  const lower = message.toLowerCase();
  const tone: AlertTone =
    lower.includes("scheduled") || lower.includes("background") || lower.includes("queued")
      ? "info"
      : "success";
  showAlert(tone, message);
}

async function updateRouteQuery(
  nextState: CashflowTransactionsQuery,
  mode: "push" | "replace" = "push",
): Promise<void> {
  const currentQuery = toCashflowTransactionsRouteQuery(queryState.value);
  const nextQuery = toCashflowTransactionsRouteQuery(nextState);

  if (areRouteQueriesEqual(currentQuery, nextQuery)) {
    return;
  }

  if (mode === "replace") {
    await router.replace({
      path: "/cashflow",
      query: nextQuery,
    });
    return;
  }

  await router.push({
    path: "/cashflow",
    query: nextQuery,
  });
}

function normalizeDraftValue(value: string): string {
  return value.trim();
}

function actionFiltersFromDrafts(): ReturnType<typeof filtersFromQuery> {
  return filtersFromQuery({
    ...queryState.value,
    q: normalizeDraftValue(searchDraft.value),
    description: normalizeDraftValue(columnFilterDraft.value.description),
    note: normalizeDraftValue(columnFilterDraft.value.note),
    source: normalizeDraftValue(columnFilterDraft.value.source),
    direction: columnFilterDraft.value.direction,
    tags: columnFilterDraft.value.untagged ? "" : normalizeDraftValue(columnFilterDraft.value.tags),
    untagged: columnFilterDraft.value.untagged,
  });
}

async function onSearchDebounced(value: string): Promise<void> {
  await updateRouteQuery({
    ...queryState.value,
    q: normalizeDraftValue(value),
    offset: 0,
  }, "replace");
}

async function onDateApply(from: string, to: string): Promise<void> {
  await updateRouteQuery({
    ...queryState.value,
    from,
    to,
    offset: 0,
  });
}

async function onDateClear(): Promise<void> {
  await updateRouteQuery({
    ...queryState.value,
    from: "",
    to: "",
    offset: 0,
  });
}

async function onShowHiddenChange(value: boolean): Promise<void> {
  await updateRouteQuery({
    ...queryState.value,
    hideIgnored: !value,
    offset: 0,
  });
}

async function onTrendRangeSelected(range: { from: string; to: string }): Promise<void> {
  await updateRouteQuery({
    ...queryState.value,
    from: range.from,
    to: range.to,
    offset: 0,
  });
}

async function onTagDistributionTagSelected(payload: {
  tag: string;
  variant: "incoming" | "outgoing";
}): Promise<void> {
  const normalizedTag = payload.tag.trim().toLowerCase();
  const direction = payload.variant === "incoming" ? "in" : "out";
  const isUntagged = normalizedTag === "untagged";

  await updateRouteQuery({
    ...queryState.value,
    direction,
    tags: isUntagged ? "" : payload.tag.trim(),
    untagged: isUntagged,
    offset: 0,
  });
}

function onColumnFilterChange(field: Exclude<keyof ColumnFilterDraft, "untagged">, value: string): void {
  const shouldDisableUntagged = field === "tags" && value.trim() !== "";
  columnFilterDraft.value = {
    ...columnFilterDraft.value,
    [field]: value,
    ...(shouldDisableUntagged ? { untagged: false } : {}),
  };

  if (columnFilterDebounceHandle) {
    clearTimeout(columnFilterDebounceHandle);
  }

  columnFilterDebounceHandle = setTimeout(() => {
    void updateRouteQuery({
      ...queryState.value,
      description: normalizeDraftValue(columnFilterDraft.value.description),
      note: normalizeDraftValue(columnFilterDraft.value.note),
      source: normalizeDraftValue(columnFilterDraft.value.source),
      direction: columnFilterDraft.value.direction,
      tags: columnFilterDraft.value.untagged ? "" : normalizeDraftValue(columnFilterDraft.value.tags),
      untagged: columnFilterDraft.value.untagged,
      offset: 0,
    }, "replace");
  }, 320);
}

function onUntaggedFilterChange(untagged: boolean): void {
  columnFilterDraft.value = {
    ...columnFilterDraft.value,
    untagged,
    ...(untagged ? { tags: "" } : {}),
  };

  if (columnFilterDebounceHandle) {
    clearTimeout(columnFilterDebounceHandle);
    columnFilterDebounceHandle = null;
  }

  void updateRouteQuery({
    ...queryState.value,
    tags: untagged ? "" : normalizeDraftValue(columnFilterDraft.value.tags),
    untagged,
    offset: 0,
  }, "replace");
}

async function onSort(field: SortBy): Promise<void> {
  const current = queryState.value;
  const nextOrder =
    current.sortBy === field ? (current.sortOrder === "asc" ? "desc" : "asc") : "asc";

  await updateRouteQuery({
    ...current,
    sortBy: field,
    sortOrder: nextOrder,
  });
}

async function onLimitChange(limit: number): Promise<void> {
  await updateRouteQuery({
    ...queryState.value,
    limit,
    offset: 0,
  });
}

async function goToFirstPage(): Promise<void> {
  await updateRouteQuery({
    ...queryState.value,
    offset: 0,
  });
}

async function goToPreviousPage(): Promise<void> {
  await updateRouteQuery({
    ...queryState.value,
    offset: Math.max(0, queryState.value.offset - queryState.value.limit),
  });
}

async function goToNextPage(): Promise<void> {
  await updateRouteQuery({
    ...queryState.value,
    offset: queryState.value.offset + queryState.value.limit,
  });
}

async function goToLastPage(): Promise<void> {
  const total = pagination.value.total;
  const limit = queryState.value.limit;
  const lastOffset = total > 0 ? Math.floor((total - 1) / limit) * limit : 0;

  await updateRouteQuery({
    ...queryState.value,
    offset: lastOffset,
  });
}

function onToggleRow(id: string, checked: boolean): void {
  if (allMatchingSelected.value) {
    allMatchingSelected.value = false;
    selectedIds.value = rows.value.map((row) => row.id);
  }

  const next = new Set(selectedIds.value);
  if (checked) {
    next.add(id);
  } else {
    next.delete(id);
  }
  selectedIds.value = Array.from(next);
}

function onToggleVisible(checked: boolean): void {
  if (allMatchingSelected.value) {
    allMatchingSelected.value = false;
    selectedIds.value = rows.value.map((row) => row.id);
  }

  const next = new Set(selectedIds.value);
  for (const row of rows.value) {
    if (checked) {
      next.add(row.id);
    } else {
      next.delete(row.id);
    }
  }
  selectedIds.value = Array.from(next);
}

function clearSelection(): void {
  allMatchingSelected.value = false;
  selectedIds.value = [];
}

function toggleAllMatchingSelection(): void {
  if (allMatchingSelected.value) {
    allMatchingSelected.value = false;
    selectedIds.value = [];
    return;
  }

  allMatchingSelected.value = true;
  selectedIds.value = [];
}

function openTagModal(): void {
  tagModalOpen.value = true;
}

function closeTagModal(): void {
  if (!mutating.value) {
    tagModalOpen.value = false;
  }
}

async function submitTag(tag: string): Promise<void> {
  if (mutating.value) {
    return;
  }

  if (!actionsEnabled.value) {
    return;
  }

  if (tag.trim() === "") {
    showAlert("danger", "Tag is required.");
    return;
  }

  mutating.value = true;
  try {
    const response =
      selectedIds.value.length > 0
        ? await tagTransactionsBySelection(selectedIds.value, tag)
        : await tagTransactionsByFilter(actionFiltersFromDrafts(), tag);

    showMutationStatus(response.status, "Tag update completed.");
    clearSelection();
    tagModalOpen.value = false;
    await Promise.all([loadTransactions(queryState.value), loadAnalytics(queryState.value)]);
  } catch (error: unknown) {
    showAlert("danger", toErrorMessage(error));
  } finally {
    mutating.value = false;
  }
}

async function applyIgnore(ignored: boolean): Promise<void> {
  if (mutating.value) {
    return;
  }

  if (!actionsEnabled.value) {
    return;
  }

  mutating.value = true;
  try {
    const response =
      selectedIds.value.length > 0
        ? await ignoreTransactionsBySelection(selectedIds.value, ignored)
        : await ignoreTransactionsByFilter(actionFiltersFromDrafts(), ignored);

    showMutationStatus(response.status, "Ignore update completed.");
    clearSelection();
    await Promise.all([loadTransactions(queryState.value), loadAnalytics(queryState.value)]);
  } catch (error: unknown) {
    showAlert("danger", toErrorMessage(error));
  } finally {
    mutating.value = false;
  }
}
</script>

<template>
  <AppShellTemplate>
    <template #top>
      <TopNavbar
        :search-value="searchDraft"
        :from="queryState.from"
        :to="queryState.to"
        :loading="mutating"
        @update:search-value="searchDraft = $event"
        @search-debounced="onSearchDebounced"
        @date-apply="onDateApply"
        @date-clear="onDateClear"
      />
    </template>

    <div class="flex h-full min-h-0 flex-col gap-3 px-4 pb-4">
      <div class="grid shrink-0 grid-cols-1 gap-3 lg:grid-cols-4">
        <section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm lg:col-span-2">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h2 class="text-sm font-semibold text-slate-700">Cashflow</h2>
            <button
              type="button"
              class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
              title="Expand cashflow graph"
              @click="openExpandedChart('trend')"
            >
              <ArrowsPointingOutIcon class="h-4 w-4" />
            </button>
          </div>
          <div class="h-[22vh] min-h-44 max-h-64">
            <TrendLineChart
              :loading="analyticsLoading"
              :data="monthlyTrend"
              @range-selected="onTrendRangeSelected"
            />
          </div>
        </section>

        <section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm lg:col-span-1">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h2 class="text-sm font-semibold text-slate-700">Incoming Tags</h2>
            <button
              type="button"
              class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
              title="Expand incoming tags graph"
              @click="openExpandedChart('incoming')"
            >
              <ArrowsPointingOutIcon class="h-4 w-4" />
            </button>
          </div>
          <div class="h-[22vh] min-h-44 max-h-64">
            <TagDonutChart
              :loading="analyticsLoading"
              :data="incomingTagDistribution"
              variant="incoming"
              @tag-selected="onTagDistributionTagSelected"
            />
          </div>
        </section>

        <section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm lg:col-span-1">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h2 class="text-sm font-semibold text-slate-700">Outgoing Tags</h2>
            <button
              type="button"
              class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
              title="Expand outgoing tags graph"
              @click="openExpandedChart('outgoing')"
            >
              <ArrowsPointingOutIcon class="h-4 w-4" />
            </button>
          </div>
          <div class="h-[22vh] min-h-44 max-h-64">
            <TagDonutChart
              :loading="analyticsLoading"
              :data="outgoingTagDistribution"
              variant="outgoing"
              @tag-selected="onTagDistributionTagSelected"
            />
          </div>
        </section>
      </div>

      <div class="min-h-0 flex-1">
        <TransactionsTable
          class="h-full"
          :rows="rows"
          :selected-ids="selectedIds"
          :all-matching-selected="allMatchingSelected"
          :sort-by="queryState.sortBy"
          :sort-order="queryState.sortOrder"
          :column-filters="columnFilterDraft"
          :show-hidden="!queryState.hideIgnored"
          :loading="loading"
          @sort="onSort"
          @toggle-row="onToggleRow"
          @toggle-visible="onToggleVisible"
          @update-filter="onColumnFilterChange"
          @update-untagged-filter="onUntaggedFilterChange"
          @update-show-hidden="onShowHiddenChange"
        >
          <template #footer>
            <TransactionsFooterBar
              :limit="pagination.limit"
              :offset="pagination.offset"
              :count="pagination.count"
              :total="pagination.total"
              :selected-count="selectedCount"
              :all-matching-selected="allMatchingSelected"
              :actions-enabled="actionsEnabled"
              :loading="isBusy"
              @change-limit="onLimitChange"
              @go-first="goToFirstPage"
              @go-prev="goToPreviousPage"
              @go-next="goToNextPage"
              @go-last="goToLastPage"
              @open-tag="openTagModal"
              @ignore="applyIgnore(true)"
              @unignore="applyIgnore(false)"
              @clear-selection="clearSelection"
              @toggle-all-matching="toggleAllMatchingSelection"
            />
          </template>
        </TransactionsTable>
      </div>
    </div>

    <div
      v-if="expandedChart"
      class="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/50 p-4"
      @click.self="closeExpandedChart"
    >
      <section class="w-full max-w-6xl rounded-3xl border border-slate-200 bg-white shadow-2xl">
        <header class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
          <h3 class="text-base font-semibold text-slate-900">{{ expandedChartTitle }}</h3>
          <button
            type="button"
            class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
            title="Close expanded chart"
            @click="closeExpandedChart"
          >
            <XMarkIcon class="h-4 w-4" />
          </button>
        </header>

        <div class="px-5 pb-5 pt-4">
          <div class="h-[72vh] min-h-[420px]">
            <TrendLineChart
              v-if="expandedChart === 'trend'"
              :loading="analyticsLoading"
              :data="monthlyTrend"
              @range-selected="onTrendRangeSelected"
            />
            <TagDonutChart
              v-else-if="expandedChart === 'incoming'"
              :loading="analyticsLoading"
              :data="incomingTagDistribution"
              variant="incoming"
              @tag-selected="onTagDistributionTagSelected"
            />
            <TagDonutChart
              v-else
              :loading="analyticsLoading"
              :data="outgoingTagDistribution"
              variant="outgoing"
              @tag-selected="onTagDistributionTagSelected"
            />
          </div>
        </div>
      </section>
    </div>

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
          v-if="alert"
          :tone="alert.tone"
          :message="alert.message"
          @close="closeAlert"
        />
      </Transition>
    </div>

    <TagModal
      :open="tagModalOpen"
      :loading="mutating"
      :target-label="tagTargetLabel"
      @close="closeTagModal"
      @confirm="submitTag"
    />
  </AppShellTemplate>
</template>

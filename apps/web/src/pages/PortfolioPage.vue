<script setup lang="ts">
import { ArrowPathIcon, ArrowsPointingOutIcon, PlusIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import BaseButton from "../components/atoms/BaseButton.vue";
import BaseToggle from "../components/atoms/BaseToggle.vue";
import IconButton from "../components/atoms/IconButton.vue";
import CreatePortfolioTransactionModal from "../components/molecules/CreatePortfolioTransactionModal.vue";
import PortfolioTabs from "../components/molecules/PortfolioTabs.vue";
import PortfolioGrowthComboChart from "../components/molecules/charts/PortfolioGrowthComboChart.vue";
import ToastMessage from "../components/molecules/ToastMessage.vue";
import PortfolioGrowthKpis from "../components/organisms/PortfolioGrowthKpis.vue";
import PortfolioTransactionsFooterBar from "../components/organisms/PortfolioTransactionsFooterBar.vue";
import TopNavbar from "../components/organisms/TopNavbar.vue";
import PortfolioPositionsTable from "../components/organisms/PortfolioPositionsTable.vue";
import PortfolioTransactionsTable from "../components/organisms/PortfolioTransactionsTable.vue";
import AppShellTemplate from "../components/templates/AppShellTemplate.vue";
import { useAppSession } from "../composables/useAppSession";
import {
  fetchPortfolioPositions,
  fetchPortfolioSnapshots,
  fetchPortfolioTransactions,
  requestPortfolioRebuild,
} from "../services/portfolio";
import { ApiError } from "../services/http";
import { getRealtimeClient, type DataChangedMessage } from "../services/realtime";
import type {
  PortfolioGrowthPoint,
  PortfolioPosition,
  PortfolioSnapshotPoint,
  PortfolioTransaction,
  PortfolioTransactionOriginFilter,
  PortfolioTransactionsPagination,
  PortfolioTransactionsQuery,
  PortfolioTransactionSortOrder,
  PortfolioTransactionTypeFilter,
} from "../types/portfolio";
import {
  areRouteQueriesEqual,
  parsePortfolioTransactionsQuery,
  toPortfolioTransactionsRouteQuery,
} from "../utils/routeQuery";

type AlertTone = "success" | "danger" | "info";
type PortfolioTab = "positions" | "transactions";

interface TransactionFilterDraft {
  type: PortfolioTransactionTypeFilter;
  origin: PortfolioTransactionOriginFilter;
  source: string;
  listing: string;
}

const session = useAppSession();
const realtimeClient = getRealtimeClient();
const route = useRoute();
const router = useRouter();

const snapshots = ref<PortfolioSnapshotPoint[]>([]);
const positions = ref<PortfolioPosition[]>([]);
const transactions = ref<PortfolioTransaction[]>([]);
const transactionsPagination = ref<PortfolioTransactionsPagination>({
  limit: 25,
  offset: 0,
  count: 0,
  total: 0,
});

const includeClosed = ref(false);
const createModalOpen = ref(false);
const activeTab = ref<PortfolioTab>("positions");
const expandedCardOpen = ref(false);

const searchDraft = ref("");
const transactionFilterDraft = ref<TransactionFilterDraft>({
  type: "",
  origin: "",
  source: "",
  listing: "",
});

const snapshotsLoading = ref(false);
const positionsLoading = ref(false);
const transactionsLoading = ref(false);
const rebuildLoading = ref(false);
const snapshotsError = ref("");
const positionsError = ref("");
const transactionsError = ref("");
const toast = ref<{ tone: AlertTone; message: string } | null>(null);

let toastTimer: ReturnType<typeof setTimeout> | null = null;
let filterDebounceTimer: ReturnType<typeof setTimeout> | null = null;
let realtimeRefreshTimer: ReturnType<typeof setTimeout> | null = null;
let unsubscribeRealtime: (() => void) | null = null;
let activeTransactionsRequestID = 0;

function firstQueryValue(value: string | string[] | null | undefined): string {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }
  return value ?? "";
}

function parseTab(value: string): PortfolioTab {
  return value === "transactions" ? "transactions" : "positions";
}

const queryState = computed(() => parsePortfolioTransactionsQuery(route.query));
const from = computed(() => queryState.value.from);
const to = computed(() => queryState.value.to);

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
  }, 4500);
}

const growthPoints = computed<PortfolioGrowthPoint[]>(() => {
  const sorted = [...snapshots.value].sort((a, b) => {
    return new Date(a.occurredAt).getTime() - new Date(b.occurredAt).getTime();
  });
  return sorted.map((point) => ({
    occurredAt: point.occurredAt,
    timeWeightedReturnPct: point.timeWeightedReturnPct,
    returnVsCostBasisPct: point.returnVsCostBasisPct,
  }));
});

const latestSnapshot = computed<PortfolioSnapshotPoint | null>(() => {
  if (snapshots.value.length === 0) {
    return null;
  }
  const sorted = [...snapshots.value].sort((a, b) => {
    return new Date(a.occurredAt).getTime() - new Date(b.occurredAt).getTime();
  });
  return sorted[sorted.length - 1] ?? null;
});

function toServiceTransactionsQuery(): PortfolioTransactionsQuery {
  return {
    accountId: session.activeAccountId.value,
    from: queryState.value.from || undefined,
    to: queryState.value.to || undefined,
    limit: queryState.value.limit,
    offset: queryState.value.offset,
    sortBy: queryState.value.sortBy,
    sortOrder: queryState.value.sortOrder,
    q: queryState.value.q,
    type: queryState.value.type,
    origin: queryState.value.origin,
    source: queryState.value.source,
    listing: queryState.value.listing,
  };
}

async function loadSnapshots(): Promise<void> {
  snapshotsLoading.value = true;
  snapshotsError.value = "";
  try {
    snapshots.value = await fetchPortfolioSnapshots(
      session.activeAccountId.value,
      from.value || undefined,
      to.value || undefined,
    );
  } catch (error: unknown) {
    snapshotsError.value = toErrorMessage(error);
  } finally {
    snapshotsLoading.value = false;
  }
}

async function loadPositions(): Promise<void> {
  positionsLoading.value = true;
  positionsError.value = "";
  try {
    const response = await fetchPortfolioPositions(session.activeAccountId.value, includeClosed.value);
    positions.value = response.data;
  } catch (error: unknown) {
    positionsError.value = toErrorMessage(error);
  } finally {
    positionsLoading.value = false;
  }
}

async function loadTransactions(): Promise<void> {
  const requestID = ++activeTransactionsRequestID;
  transactionsLoading.value = true;
  transactionsError.value = "";
  try {
    const response = await fetchPortfolioTransactions(toServiceTransactionsQuery());
    if (requestID !== activeTransactionsRequestID) {
      return;
    }
    transactions.value = response.data;
    transactionsPagination.value = response.pagination;
  } catch (error: unknown) {
    if (requestID !== activeTransactionsRequestID) {
      return;
    }
    transactionsError.value = toErrorMessage(error);
  } finally {
    if (requestID === activeTransactionsRequestID) {
      transactionsLoading.value = false;
    }
  }
}

async function onRebuildPortfolio(): Promise<void> {
  if (rebuildLoading.value) {
    return;
  }
  rebuildLoading.value = true;
  try {
    await requestPortfolioRebuild(session.activeAccountId.value);
    showToast("info", "Portfolio rebuild queued in the background.");
  } catch (error: unknown) {
    showToast("danger", toErrorMessage(error));
  } finally {
    rebuildLoading.value = false;
  }
}

async function updateRouteQuery(
  nextState: ReturnType<typeof parsePortfolioTransactionsQuery>,
  nextTab: PortfolioTab = activeTab.value,
  mode: "push" | "replace" = "push",
): Promise<void> {
  const nextQuery = toPortfolioTransactionsRouteQuery(nextState, nextTab);
  if (areRouteQueriesEqual(route.query, nextQuery)) {
    return;
  }
  if (mode === "replace") {
    await router.replace({
      path: route.path,
      query: nextQuery,
    });
    return;
  }
  await router.push({
    path: route.path,
    query: nextQuery,
  });
}

async function onDateApply(nextFrom: string, nextTo: string): Promise<void> {
  await updateRouteQuery(
    {
      ...queryState.value,
      from: nextFrom,
      to: nextTo,
      offset: 0,
    },
    activeTab.value,
  );
}

async function onDateClear(): Promise<void> {
  await updateRouteQuery(
    {
      ...queryState.value,
      from: "",
      to: "",
      offset: 0,
    },
    activeTab.value,
  );
}

async function onChartRangeSelected(range: { from: string; to: string }): Promise<void> {
  await onDateApply(range.from, range.to);
}

async function onTabChange(nextTab: PortfolioTab): Promise<void> {
  if (nextTab === activeTab.value) {
    return;
  }
  activeTab.value = nextTab;
  await updateRouteQuery(queryState.value, nextTab);
}

async function onSearchDebounced(value: string): Promise<void> {
  await updateRouteQuery(
    {
      ...queryState.value,
      q: value.trim(),
      offset: 0,
    },
    activeTab.value,
    "replace",
  );
}

function onTransactionFilterChange(field: keyof TransactionFilterDraft, value: string): void {
  transactionFilterDraft.value = {
    ...transactionFilterDraft.value,
    [field]: value,
  };
  if (field === "source" || field === "listing") {
    if (filterDebounceTimer) {
      clearTimeout(filterDebounceTimer);
    }
    filterDebounceTimer = setTimeout(() => {
      void updateRouteQuery(
        {
          ...queryState.value,
          source: transactionFilterDraft.value.source.trim(),
          listing: transactionFilterDraft.value.listing.trim(),
          offset: 0,
        },
        activeTab.value,
        "replace",
      );
    }, 320);
    return;
  }

  void updateRouteQuery(
    {
      ...queryState.value,
      type: transactionFilterDraft.value.type,
      origin: transactionFilterDraft.value.origin,
      offset: 0,
    },
    activeTab.value,
  );
}

async function onToggleTransactionSort(): Promise<void> {
  const nextSortOrder: PortfolioTransactionSortOrder = queryState.value.sortOrder === "asc" ? "desc" : "asc";
  await updateRouteQuery(
    {
      ...queryState.value,
      sortOrder: nextSortOrder,
    },
    activeTab.value,
  );
}

async function onLimitChange(limit: number): Promise<void> {
  await updateRouteQuery(
    {
      ...queryState.value,
      limit,
      offset: 0,
    },
    activeTab.value,
  );
}

async function goToFirstPage(): Promise<void> {
  await updateRouteQuery(
    {
      ...queryState.value,
      offset: 0,
    },
    activeTab.value,
  );
}

async function goToPreviousPage(): Promise<void> {
  await updateRouteQuery(
    {
      ...queryState.value,
      offset: Math.max(0, queryState.value.offset - queryState.value.limit),
    },
    activeTab.value,
  );
}

async function goToNextPage(): Promise<void> {
  await updateRouteQuery(
    {
      ...queryState.value,
      offset: queryState.value.offset + queryState.value.limit,
    },
    activeTab.value,
  );
}

async function goToLastPage(): Promise<void> {
  const total = transactionsPagination.value.total;
  const limit = queryState.value.limit;
  const lastOffset = total > 0 ? Math.floor((total - 1) / limit) * limit : 0;
  await updateRouteQuery(
    {
      ...queryState.value,
      offset: lastOffset,
    },
    activeTab.value,
  );
}

async function onTransactionCreated(): Promise<void> {
  showToast("success", "Transaction created successfully.");
  await loadTransactions();
}

function closeExpandedCard(): void {
  expandedCardOpen.value = false;
}

function onEscKeyDown(event: KeyboardEvent): void {
  if (event.key === "Escape" && expandedCardOpen.value) {
    closeExpandedCard();
  }
}

function onRealtimeMessage(message: DataChangedMessage): void {
  if (message.account_id !== session.activeAccountId.value) {
    return;
  }
  if (message.event !== "import.completed" && message.event !== "portfolio.rebuilt") {
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
    void loadSnapshots();
    void loadPositions();
    void loadTransactions();
  }, 250);
}

watch(includeClosed, () => {
  void loadPositions();
});

watch(
  () => session.activeAccountId.value,
  (accountID, prevAccountID) => {
    if (accountID === prevAccountID) {
      return;
    }
    realtimeClient.setAccountId(accountID);
    void loadSnapshots();
    void loadPositions();
    void loadTransactions();
  },
);

watch(
  () => [from.value, to.value],
  () => {
    void loadSnapshots();
  },
  { immediate: true },
);

watch(
  queryState,
  (next) => {
    searchDraft.value = next.q;
    transactionFilterDraft.value = {
      type: next.type,
      origin: next.origin,
      source: next.source,
      listing: next.listing,
    };
    void loadTransactions();
  },
  { immediate: true, deep: true },
);

watch(
  () => firstQueryValue(route.query.tab as string | string[] | undefined).trim(),
  (tabValue) => {
    activeTab.value = parseTab(tabValue);
  },
  { immediate: true },
);

watch(
  () => route.query,
  async () => {
    const normalized = toPortfolioTransactionsRouteQuery(queryState.value, activeTab.value);
    if (!areRouteQueriesEqual(route.query, normalized)) {
      await router.replace({ path: route.path, query: normalized });
    }
  },
  { immediate: true },
);

onMounted(() => {
  realtimeClient.setAccountId(session.activeAccountId.value);
  unsubscribeRealtime = realtimeClient.subscribe(onRealtimeMessage);
  void loadPositions();
  window.addEventListener("keydown", onEscKeyDown);
});

onBeforeUnmount(() => {
  if (toastTimer) {
    clearTimeout(toastTimer);
    toastTimer = null;
  }
  if (filterDebounceTimer) {
    clearTimeout(filterDebounceTimer);
    filterDebounceTimer = null;
  }
  if (realtimeRefreshTimer) {
    clearTimeout(realtimeRefreshTimer);
    realtimeRefreshTimer = null;
  }
  if (unsubscribeRealtime) {
    unsubscribeRealtime();
    unsubscribeRealtime = null;
  }
  window.removeEventListener("keydown", onEscKeyDown);
});
</script>

<template>
  <AppShellTemplate>
    <template #top>
      <TopNavbar
        :search-value="searchDraft"
        :from="from"
        :to="to"
        :loading="snapshotsLoading || positionsLoading || transactionsLoading || rebuildLoading"
        :show-filter-controls="true"
        :show-search-control="activeTab === 'transactions'"
        :show-date-control="true"
        @update:search-value="searchDraft = $event"
        @search-debounced="onSearchDebounced"
        @date-apply="onDateApply"
        @date-clear="onDateClear"
      />
    </template>

    <div class="flex h-full min-h-0 flex-col gap-3 px-4 pb-4">
      <section class="rounded-3xl border border-slate-300 bg-white p-4 shadow-sm">
        <header class="mb-3 flex items-center justify-between gap-2">
          <h2 class="font-secondary text-xl font-semibold text-slate-700 md:text-2xl">Portfolio Performance</h2>
          <div class="flex items-center gap-2">
            <IconButton
              tone="neutral"
              size="sm"
              :disabled="rebuildLoading || snapshotsLoading || positionsLoading"
              title="Rebuild portfolio"
              @click="void onRebuildPortfolio()"
            >
              <ArrowPathIcon class="h-4 w-4" :class="{ 'animate-spin': rebuildLoading }" />
            </IconButton>
            <BaseButton v-if="snapshotsError" size="sm" variant="secondary" @click="void loadSnapshots()">Retry</BaseButton>
          </div>
        </header>

        <div class="mb-3">
          <PortfolioGrowthKpis
            :market-value="latestSnapshot?.marketValue"
            :total-pnl="latestSnapshot?.totalPnL"
            :total-pnl-pct="latestSnapshot?.totalPnLPct"
            :loading="snapshotsLoading"
          />
        </div>

        <div class="h-[30vh] min-h-52">
          <PortfolioGrowthComboChart
            :loading="snapshotsLoading"
            :data="growthPoints"
            @range-selected="onChartRangeSelected"
          />
        </div>
        <p v-if="snapshotsError" class="mt-2 text-sm text-rose-700">{{ snapshotsError }}</p>
      </section>

      <div class="min-h-0 flex-1" v-if="!expandedCardOpen">
        <section class="relative flex h-full min-h-0 flex-col overflow-hidden rounded-3xl border border-slate-300 bg-white/95 p-4 shadow-sm">
          <header class="mb-3 flex items-center justify-between gap-3">
            <PortfolioTabs :model-value="activeTab" @update:model-value="void onTabChange($event)" />
            <div class="flex items-center gap-2">
              <div v-if="activeTab === 'positions'" class="flex items-center gap-2">
                <span class="text-xs font-medium uppercase tracking-wide text-slate-500">Include closed</span>
                <BaseToggle
                  :checked="includeClosed"
                  :disabled="positionsLoading"
                  @update:checked="includeClosed = $event"
                />
              </div>
              <IconButton
                tone="neutral"
                size="sm"
                title="Expand portfolio card"
                @click="expandedCardOpen = true"
              >
                <ArrowsPointingOutIcon class="h-4 w-4" />
              </IconButton>
            </div>
          </header>

          <div class="min-h-0 flex-1">
            <PortfolioPositionsTable
              v-if="activeTab === 'positions'"
              class="h-full"
              :rows="positions"
              :loading="positionsLoading"
              :include-closed="includeClosed"
              :error-message="positionsError"
              :framed="false"
              :show-include-closed-control="false"
              @retry="void loadPositions()"
              @update:include-closed="includeClosed = $event"
            />
            <PortfolioTransactionsTable
              v-else
              class="h-full"
              :rows="transactions"
              :sort-order="queryState.sortOrder"
              :filters="transactionFilterDraft"
              :loading="transactionsLoading"
              :error-message="transactionsError"
              :framed="false"
              @retry="void loadTransactions()"
              @sort-date="void onToggleTransactionSort()"
              @update-filter="onTransactionFilterChange"
            >
              <template #footer>
                <PortfolioTransactionsFooterBar
                  :limit="queryState.limit"
                  :offset="queryState.offset"
                  :count="transactionsPagination.count"
                  :total="transactionsPagination.total"
                  :loading="transactionsLoading"
                  @change-limit="onLimitChange"
                  @go-first="goToFirstPage"
                  @go-prev="goToPreviousPage"
                  @go-next="goToNextPage"
                  @go-last="goToLastPage"
                />
              </template>
            </PortfolioTransactionsTable>
          </div>

          <div class="absolute bottom-6 right-6 z-40">
            <IconButton
              tone="primary"
              size="fab"
              title="Add transaction"
              @click="createModalOpen = true"
            >
              <PlusIcon class="h-6 w-6" />
            </IconButton>
          </div>
        </section>
      </div>
    </div>

    <div
      v-if="expandedCardOpen"
      class="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/50 p-4"
      @click.self="closeExpandedCard"
    >
      <section class="flex h-[92vh] w-full max-w-7xl min-h-0 flex-col rounded-3xl border border-slate-300 bg-white shadow-2xl">
        <header class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
          <h3 class="font-secondary text-xl font-semibold text-slate-900 md:text-2xl">Portfolio</h3>
          <button
            type="button"
            class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
            title="Close expanded portfolio card"
            @click="closeExpandedCard"
          >
            <XMarkIcon class="h-4 w-4" />
          </button>
        </header>

        <div class="min-h-0 flex-1 p-4">
          <div class="flex h-full min-h-0 flex-col">
            <header class="mb-3 flex items-center justify-between gap-3">
              <PortfolioTabs :model-value="activeTab" @update:model-value="void onTabChange($event)" />
              <div v-if="activeTab === 'positions'" class="flex items-center gap-2">
                <span class="text-xs font-medium uppercase tracking-wide text-slate-500">Include closed</span>
                <BaseToggle
                  :checked="includeClosed"
                  :disabled="positionsLoading"
                  @update:checked="includeClosed = $event"
                />
              </div>
            </header>

            <div class="min-h-0 flex-1">
              <PortfolioPositionsTable
                v-if="activeTab === 'positions'"
                class="h-full"
                :rows="positions"
                :loading="positionsLoading"
                :include-closed="includeClosed"
                :error-message="positionsError"
                :framed="false"
                :show-include-closed-control="false"
                @retry="void loadPositions()"
                @update:include-closed="includeClosed = $event"
              />
              <PortfolioTransactionsTable
                v-else
                class="h-full"
                :rows="transactions"
                :sort-order="queryState.sortOrder"
                :filters="transactionFilterDraft"
                :loading="transactionsLoading"
                :error-message="transactionsError"
                :framed="false"
                @retry="void loadTransactions()"
                @sort-date="void onToggleTransactionSort()"
                @update-filter="onTransactionFilterChange"
              >
                <template #footer>
                  <PortfolioTransactionsFooterBar
                    :limit="queryState.limit"
                    :offset="queryState.offset"
                    :count="transactionsPagination.count"
                    :total="transactionsPagination.total"
                    :loading="transactionsLoading"
                    @change-limit="onLimitChange"
                    @go-first="goToFirstPage"
                    @go-prev="goToPreviousPage"
                    @go-next="goToNextPage"
                    @go-last="goToLastPage"
                  />
                </template>
              </PortfolioTransactionsTable>
            </div>
          </div>
        </div>
      </section>
    </div>

    <CreatePortfolioTransactionModal
      :open="createModalOpen"
      :account-id="session.activeAccountId.value"
      @close="createModalOpen = false"
      @created="void onTransactionCreated()"
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


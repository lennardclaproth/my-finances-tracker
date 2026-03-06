<script setup lang="ts">
import { PlusIcon } from "@heroicons/vue/24/outline";
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import BaseButton from "../components/atoms/BaseButton.vue";
import IconButton from "../components/atoms/IconButton.vue";
import ListingSearchSelect from "../components/molecules/ListingSearchSelect.vue";
import ToastMessage from "../components/molecules/ToastMessage.vue";
import UploadDailyDataModal from "../components/molecules/UploadDailyDataModal.vue";
import TopNavbar from "../components/organisms/TopNavbar.vue";
import ListingDailiesTable from "../components/organisms/ListingDailiesTable.vue";
import AppShellTemplate from "../components/templates/AppShellTemplate.vue";
import { ApiError } from "../services/http";
import { fetchListingDailies } from "../services/marketdata";
import { searchListings } from "../services/listings";
import type { Listing } from "../types/listings";
import type { ListingDaily } from "../types/marketdata";
import { areRouteQueriesEqual } from "../utils/routeQuery";

type AlertTone = "success" | "danger" | "info";

interface DailiesQueryState {
  listingId: string;
  listingSymbol: string;
  from: string;
  to: string;
  limit: number;
  offset: number;
}

const DEFAULT_LIMIT = 25;
const LIMIT_OPTIONS = [10, 25, 50, 100];

const route = useRoute();
const router = useRouter();

const selectedListing = ref<Listing | null>(null);
const selectedListingId = ref("");
const dailies = ref<ListingDaily[]>([]);
const loading = ref(false);
const errorMessage = ref("");
const totalCount = ref(0);
const uploadModalOpen = ref(false);
const toast = ref<{ tone: AlertTone; message: string } | null>(null);
let toastTimer: ReturnType<typeof setTimeout> | null = null;
let activeLoadId = 0;
let activeSyncId = 0;

function firstQueryValue(value: string | string[] | null | undefined): string {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }
  return value ?? "";
}

function parseLimit(raw: string): number {
  const parsed = Number.parseInt(raw, 10);
  if (LIMIT_OPTIONS.includes(parsed)) {
    return parsed;
  }
  return DEFAULT_LIMIT;
}

function parseOffset(raw: string): number {
  const parsed = Number.parseInt(raw, 10);
  if (Number.isNaN(parsed) || parsed < 0) {
    return 0;
  }
  return parsed;
}

function parseQueryState(): DailiesQueryState {
  return {
    listingId: firstQueryValue(route.query.listing_id as string | string[] | undefined).trim(),
    listingSymbol: firstQueryValue(route.query.listing_symbol as string | string[] | undefined).trim(),
    from: firstQueryValue(route.query.from as string | string[] | undefined).trim(),
    to: firstQueryValue(route.query.to as string | string[] | undefined).trim(),
    limit: parseLimit(firstQueryValue(route.query.limit as string | string[] | undefined).trim()),
    offset: parseOffset(firstQueryValue(route.query.offset as string | string[] | undefined).trim()),
  };
}

const queryState = computed(parseQueryState);

function toRouteQuery(state: DailiesQueryState): Record<string, string> {
  const query: Record<string, string> = {
    limit: String(state.limit),
    offset: String(state.offset),
  };
  if (state.listingId) {
    query.listing_id = state.listingId;
  }
  if (state.listingSymbol) {
    query.listing_symbol = state.listingSymbol;
  }
  if (state.from) {
    query.from = state.from;
  }
  if (state.to) {
    query.to = state.to;
  }
  return query;
}

async function updateRouteQuery(next: DailiesQueryState, mode: "push" | "replace" = "push"): Promise<void> {
  const nextQuery = toRouteQuery(next);
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

function toErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "Failed to load listing dailies.";
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

function toFallbackListing(id: string, symbol: string): Listing {
  return {
    id,
    symbol,
    name: symbol,
    source: "",
    createdAt: "",
    updatedAt: "",
  };
}

async function resolveListing(listingId: string, listingSymbol: string): Promise<Listing | null> {
  if (!listingId || !listingSymbol) {
    return null;
  }
  try {
    const response = await searchListings(listingSymbol, 25, 0);
    const byID = response.data.find((item) => item.id === listingId);
    if (byID) {
      return byID;
    }
    const bySymbol = response.data.find((item) => item.symbol.toLowerCase() === listingSymbol.toLowerCase());
    if (bySymbol) {
      return bySymbol;
    }
  } catch {
    // Fall back to query values when search lookup fails.
  }
  return toFallbackListing(listingId, listingSymbol);
}

async function loadDailies(state: DailiesQueryState): Promise<void> {
  if (!state.listingSymbol) {
    dailies.value = [];
    totalCount.value = 0;
    errorMessage.value = "";
    loading.value = false;
    return;
  }

  const requestId = ++activeLoadId;
  loading.value = true;
  errorMessage.value = "";
  try {
    const response = await fetchListingDailies({
      listingId: state.listingId || undefined,
      symbol: state.listingSymbol,
      from: state.from || undefined,
      to: state.to || undefined,
      sortOrder: "desc",
      limit: state.limit,
      offset: state.offset,
    });
    if (requestId !== activeLoadId) {
      return;
    }
    dailies.value = response.data;
    totalCount.value = response.metadata.totalCount;
  } catch (error: unknown) {
    if (requestId !== activeLoadId) {
      return;
    }
    errorMessage.value = toErrorMessage(error);
    dailies.value = [];
    totalCount.value = 0;
  } finally {
    if (requestId === activeLoadId) {
      loading.value = false;
    }
  }
}

async function syncFromQuery(state: DailiesQueryState, syncID: number): Promise<void> {
  if (!state.listingId || !state.listingSymbol) {
    selectedListingId.value = "";
    selectedListing.value = null;
    if (syncID === activeSyncId) {
      await loadDailies(state);
    }
    return;
  }

  if (
    selectedListingId.value !== state.listingId ||
    selectedListing.value?.symbol.toLowerCase() !== state.listingSymbol.toLowerCase()
  ) {
    const resolvedListing = await resolveListing(state.listingId, state.listingSymbol);
    if (syncID !== activeSyncId) {
      return;
    }
    if (resolvedListing) {
      selectedListingId.value = resolvedListing.id;
      selectedListing.value = resolvedListing;
    } else {
      selectedListingId.value = state.listingId;
      selectedListing.value = toFallbackListing(state.listingId, state.listingSymbol);
    }
  }

  if (syncID === activeSyncId) {
    await loadDailies(state);
  }
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

async function onListingSelected(listing: Listing | null): Promise<void> {
  if (!listing) {
    selectedListing.value = null;
    selectedListingId.value = "";
    await updateRouteQuery({
      ...queryState.value,
      listingId: "",
      listingSymbol: "",
      offset: 0,
    });
    return;
  }
  selectedListingId.value = listing.id;
  selectedListing.value = listing;
  await updateRouteQuery({
    ...queryState.value,
    listingId: listing.id,
    listingSymbol: listing.symbol,
    offset: 0,
  });
}

async function onLimitChange(limit: number): Promise<void> {
  const normalizedLimit = parseLimit(String(limit));
  await updateRouteQuery({
    ...queryState.value,
    limit: normalizedLimit,
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
  const total = totalCount.value;
  const limit = queryState.value.limit;
  const lastOffset = total > 0 ? Math.floor((total - 1) / limit) * limit : 0;
  await updateRouteQuery({
    ...queryState.value,
    offset: lastOffset,
  });
}

function onUploadAccepted(): void {
  showToast("success", "Daily upload accepted and queued for processing.");
}

watch(
  queryState,
  (nextState) => {
    const syncID = ++activeSyncId;
    void syncFromQuery(nextState, syncID);
  },
  { immediate: true },
);

watch(
  () => route.query,
  async () => {
    const normalized = toRouteQuery(queryState.value);
    if (!areRouteQueriesEqual(route.query, normalized)) {
      await router.replace({ path: route.path, query: normalized });
    }
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  if (toastTimer) {
    clearTimeout(toastTimer);
    toastTimer = null;
  }
});
</script>

<template>
  <AppShellTemplate>
    <template #top>
      <TopNavbar
        :from="queryState.from"
        :to="queryState.to"
        :loading="loading"
        :show-filter-controls="true"
        :show-search-control="false"
        :show-date-control="true"
        @date-apply="onDateApply"
        @date-clear="onDateClear"
      />
    </template>

    <div class="flex h-full min-h-0 flex-col gap-3 px-4 pb-4">
      <section class="relative flex h-full min-h-0 flex-col gap-3 rounded-3xl border border-slate-300 bg-white/95 p-4 shadow-sm">
        <header class="flex flex-wrap items-end justify-between gap-3">
          <div class="w-full max-w-2xl space-y-1">
            <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Listing</p>
            <ListingSearchSelect
              :model-value="selectedListingId"
              :selected-listing="selectedListing"
              :disabled="loading"
              @update:model-value="selectedListingId = $event"
              @select="void onListingSelected($event)"
            />
            <p v-if="selectedListing" class="text-xs text-slate-500">
              {{ selectedListing.symbol }} | ISIN: {{ selectedListing.isin || "-" }}
            </p>
          </div>

          <div class="flex items-center gap-2">
            <BaseButton size="sm" variant="secondary" :disabled="loading" @click="void loadDailies(queryState)">
              Retry
            </BaseButton>
          </div>
        </header>

        <ListingDailiesTable
          class="min-h-0 flex-1"
          :rows="dailies"
          :limit="queryState.limit"
          :offset="queryState.offset"
          :total="totalCount"
          :loading="loading"
          :error-message="errorMessage"
          :empty-message="selectedListing ? 'No daily data found for this listing and date range.' : 'Select a listing to view daily data.'"
          :framed="false"
          :page-size-options="LIMIT_OPTIONS"
          @retry="void loadDailies(queryState)"
          @change-limit="void onLimitChange($event)"
          @go-first="void goToFirstPage()"
          @go-prev="void goToPreviousPage()"
          @go-next="void goToNextPage()"
          @go-last="void goToLastPage()"
        />

        <div class="absolute bottom-6 right-6 z-40">
          <IconButton
            tone="primary"
            size="fab"
            title="Upload dailies"
            @click="uploadModalOpen = true"
          >
            <PlusIcon class="h-6 w-6" />
          </IconButton>
        </div>
      </section>
    </div>

    <UploadDailyDataModal
      :open="uploadModalOpen"
      :initial-listing="selectedListing"
      @close="uploadModalOpen = false"
      @uploaded="onUploadAccepted"
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

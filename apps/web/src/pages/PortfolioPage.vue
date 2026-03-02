<script setup lang="ts">
import { ArrowPathIcon } from "@heroicons/vue/24/outline";
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import BaseButton from "../components/atoms/BaseButton.vue";
import IconButton from "../components/atoms/IconButton.vue";
import PortfolioGrowthComboChart from "../components/molecules/charts/PortfolioGrowthComboChart.vue";
import ToastMessage from "../components/molecules/ToastMessage.vue";
import PortfolioGrowthKpis from "../components/organisms/PortfolioGrowthKpis.vue";
import TopNavbar from "../components/organisms/TopNavbar.vue";
import PortfolioPositionsTable from "../components/organisms/PortfolioPositionsTable.vue";
import AppShellTemplate from "../components/templates/AppShellTemplate.vue";
import { useAppSession } from "../composables/useAppSession";
import { fetchPortfolioPositions, fetchPortfolioSnapshots, requestPortfolioRebuild } from "../services/portfolio";
import { ApiError } from "../services/http";
import type { PortfolioGrowthPoint, PortfolioPosition, PortfolioSnapshotPoint } from "../types/portfolio";

type AlertTone = "success" | "danger" | "info";

const session = useAppSession();
const route = useRoute();
const router = useRouter();

const snapshots = ref<PortfolioSnapshotPoint[]>([]);
const positions = ref<PortfolioPosition[]>([]);
const includeClosed = ref(false);

const snapshotsLoading = ref(false);
const positionsLoading = ref(false);
const rebuildLoading = ref(false);
const snapshotsError = ref("");
const positionsError = ref("");
const toast = ref<{ tone: AlertTone; message: string } | null>(null);
let toastTimer: ReturnType<typeof setTimeout> | null = null;

function firstQueryValue(value: string | string[] | null | undefined): string {
  if (Array.isArray(value)) {
    return value[0] ?? "";
  }
  return value ?? "";
}

const from = computed(() => firstQueryValue(route.query.from as string | string[] | undefined).trim());
const to = computed(() => firstQueryValue(route.query.to as string | string[] | undefined).trim());

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
  if (sorted.length === 0) {
    return [];
  }
  return sorted.map((point) => {
    return {
      occurredAt: point.occurredAt,
      timeWeightedReturnPct: point.timeWeightedReturnPct,
      returnVsCostBasisPct: point.returnVsCostBasisPct,
    };
  });
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

async function updateDateRange(nextFrom: string, nextTo: string, mode: "push" | "replace" = "push"): Promise<void> {
  const nextQuery: Record<string, string> = {};
  if (nextFrom) {
    nextQuery.from = nextFrom;
  }
  if (nextTo) {
    nextQuery.to = nextTo;
  }

  const currentFrom = from.value;
  const currentTo = to.value;
  if (currentFrom === (nextFrom || "") && currentTo === (nextTo || "")) {
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
  await updateDateRange(nextFrom, nextTo);
}

async function onDateClear(): Promise<void> {
  await updateDateRange("", "");
}

async function onChartRangeSelected(range: { from: string; to: string }): Promise<void> {
  await updateDateRange(range.from, range.to);
}

watch(includeClosed, () => {
  void loadPositions();
});

watch(
  () => [from.value, to.value],
  () => {
    void loadSnapshots();
  },
  { immediate: true },
);

onMounted(() => {
  void loadPositions();
});

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
        :from="from"
        :to="to"
        :loading="snapshotsLoading || positionsLoading || rebuildLoading"
        :show-filter-controls="true"
        :show-search-control="false"
        :show-date-control="true"
        @date-apply="onDateApply"
        @date-clear="onDateClear"
      />
    </template>

    <div class="flex h-full min-h-0 flex-col gap-3 px-4 pb-4">
      <section class="rounded-3xl border border-slate-200 bg-white p-4 shadow-sm">
        <header class="mb-3 flex items-center justify-between gap-2">
          <h2 class="text-sm font-semibold text-slate-700">Portfolio Growth</h2>
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

      <div class="min-h-0 flex-1">
        <PortfolioPositionsTable
          class="h-full"
          :rows="positions"
          :loading="positionsLoading"
          :include-closed="includeClosed"
          :error-message="positionsError"
          @retry="void loadPositions()"
          @update:include-closed="includeClosed = $event"
        />
      </div>
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
          v-if="toast"
          :tone="toast.tone"
          :message="toast.message"
          @close="toast = null"
        />
      </Transition>
    </div>
  </AppShellTemplate>
</template>

<script setup lang="ts">
import { InformationCircleIcon } from "@heroicons/vue/24/outline";
import { Chart } from "chart.js/auto";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import BasePopover from "../../atoms/BasePopover.vue";

interface DonutSlice {
  label: string;
  value: number;
}

interface Props {
  loading?: boolean;
  data?: DonutSlice[];
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  data: () => [],
});

const canvasRef = ref<HTMLCanvasElement | null>(null);
let chart: Chart<"doughnut"> | null = null;

const hasData = computed(() => props.data.some((item) => item.value > 0));
const MAX_VISIBLE_LABELS = 4;

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 0,
});

const palette = [
  "#047857",
  "#059669",
  "#10b981",
  "#34d399",
  "#4f46e5",
  "#6366f1",
  "#4338ca",
  "#0f766e",
  "#f59e0b",
  "#f97316",
];

const sortedEntries = computed(() =>
  [...props.data].sort((a, b) => b.value - a.value),
);

const totalValue = computed(() =>
  sortedEntries.value.reduce((sum, entry) => sum + entry.value, 0),
);

function formatPercentage(value: number): string {
  const rounded = Math.round(value * 10) / 10;
  return `${Number.isInteger(rounded) ? rounded.toFixed(0) : rounded.toFixed(1)}%`;
}

const topLabelEntries = computed(() => {
  const total = totalValue.value;
  return sortedEntries.value.slice(0, MAX_VISIBLE_LABELS).map((entry, index) => ({
    ...entry,
    color: palette[index % palette.length],
    percentageLabel: total > 0 ? formatPercentage((entry.value / total) * 100) : "0%",
  }));
});

const allLabelEntries = computed(() => {
  const total = totalValue.value;
  return sortedEntries.value.map((entry, index) => ({
    ...entry,
    color: palette[index % palette.length],
    percentageLabel: total > 0 ? formatPercentage((entry.value / total) * 100) : "0%",
  }));
});

const hiddenLabelCount = computed(() => Math.max(0, sortedEntries.value.length - MAX_VISIBLE_LABELS));

function destroyChart(): void {
  chart?.destroy();
  chart = null;
}

function renderChart(): void {
  if (props.loading || !canvasRef.value || !hasData.value) {
    destroyChart();
    return;
  }
  destroyChart();

  chart = new Chart(canvasRef.value, {
    type: "doughnut",
    data: {
      labels: sortedEntries.value.map((item) => item.label),
      datasets: [
        {
          data: sortedEntries.value.map((item) => item.value),
          backgroundColor: sortedEntries.value.map((_, index) => palette[index % palette.length]),
          borderColor: "#ffffff",
          borderWidth: 1,
          hoverOffset: 3,
        },
      ],
    },
    options: {
      maintainAspectRatio: false,
      responsive: true,
      cutout: "64%",
      plugins: {
        legend: {
          display: false,
        },
        tooltip: {
          backgroundColor: "#0f172a",
          titleColor: "#f8fafc",
          bodyColor: "#f8fafc",
          borderColor: "rgba(148, 163, 184, 0.35)",
          borderWidth: 1,
          callbacks: {
            label: (ctx) => {
              const raw = Number(ctx.parsed);
              const total = totalValue.value;
              const ratio = total === 0 ? 0 : (raw / total) * 100;
              return `${ctx.label}: ${currencyFormatter.format(raw)} (${formatPercentage(ratio)})`;
            },
          },
        },
      },
    },
  });
}

onMounted(async () => {
  await nextTick();
  renderChart();
});

watch(
  () => [props.loading, props.data],
  async () => {
    await nextTick();
    renderChart();
  },
  { deep: true },
);

onBeforeUnmount(() => {
  destroyChart();
});
</script>

<template>
  <div class="flex h-full w-full flex-col">
    <div class="relative min-h-0 flex-1">
      <canvas ref="canvasRef" class="h-full w-full" aria-label="Asset class distribution chart" />
      <div
        v-if="loading"
        class="absolute inset-0 bg-white/90 p-3"
      >
        <div class="flex h-full w-full flex-col items-center justify-center gap-3">
          <div class="h-28 w-28 animate-pulse rounded-full border-[14px] border-slate-200 bg-slate-100" />
          <div class="flex flex-wrap justify-center gap-2">
            <div class="h-5 w-16 animate-pulse rounded-full bg-slate-200" />
            <div class="h-5 w-20 animate-pulse rounded-full bg-slate-200" />
            <div class="h-5 w-14 animate-pulse rounded-full bg-slate-200" />
          </div>
        </div>
      </div>
      <div
        v-else-if="!hasData"
        class="absolute inset-0 flex items-center justify-center bg-white/75 text-sm text-slate-500"
      >
        No class worth distribution yet
      </div>
    </div>

    <div
      v-if="!loading && hasData"
      class="mt-2 flex flex-wrap items-center gap-2"
    >
      <span
        v-for="entry in topLabelEntries"
        :key="entry.label"
        class="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-[11px] font-medium text-slate-700"
      >
        <span
          class="h-1.5 w-1.5 rounded-full"
          :style="{ backgroundColor: entry.color }"
        />
        <span>{{ entry.label }} {{ entry.percentageLabel }}</span>
      </span>
      <BasePopover
        v-if="hiddenLabelCount > 0"
        side="top"
        align="right"
        offset-class="mb-2"
        panel-class="w-56 rounded-lg border border-slate-200 bg-white p-2 shadow-lg"
      >
        <template #trigger="{ toggle }">
          <button
            type="button"
            class="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-white px-2 py-0.5 text-[11px] font-medium text-slate-500 transition hover:border-slate-300 hover:text-slate-700"
            title="Show all asset classes"
            @click="toggle"
          >
            <span>+{{ hiddenLabelCount }} more classes</span>
            <InformationCircleIcon class="h-3.5 w-3.5" />
          </button>
        </template>
        <template #default="{ close }">
          <p class="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-500">All asset classes</p>
          <div class="max-h-56 space-y-1 overflow-auto pr-1">
            <button
              v-for="entry in allLabelEntries"
              :key="`all-${entry.label}`"
              type="button"
              class="flex w-full items-center justify-between gap-2 rounded-md bg-slate-50 px-2 py-1.5 text-left text-xs text-slate-700 transition hover:bg-slate-100"
              @click="close()"
            >
              <span class="inline-flex min-w-0 items-center gap-1.5">
                <span class="h-1.5 w-1.5 shrink-0 rounded-full" :style="{ backgroundColor: entry.color }" />
                <span class="truncate">{{ entry.label }}</span>
              </span>
              <span class="shrink-0 font-semibold text-slate-600">{{ entry.percentageLabel }}</span>
            </button>
          </div>
        </template>
      </BasePopover>
    </div>
  </div>
</template>

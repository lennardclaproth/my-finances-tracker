<script setup lang="ts">
import { InformationCircleIcon } from "@heroicons/vue/24/outline";
import { Chart } from "chart.js/auto";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { CashflowTagDistributionEntry } from "../../../types/cashflow";
import BasePopover from "../../atoms/BasePopover.vue";

interface Props {
  loading?: boolean;
  data?: CashflowTagDistributionEntry[];
  variant?: "incoming" | "outgoing";
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  data: () => [],
  variant: "incoming",
});

const emit = defineEmits<{
  (event: "tag-selected", value: { tag: string; variant: "incoming" | "outgoing" }): void;
}>();

const canvasRef = ref<HTMLCanvasElement | null>(null);
const hasData = computed(() => props.data.length > 0);

let chart: Chart<"doughnut"> | null = null;

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 0,
});
const MONEY_SCALE = 1_000_000;
const chartTheme = {
  text: "#334155",
  muted: "#64748b",
  tooltipBg: "#0f172a",
};
const MAX_VISIBLE_LABELS = 4;

function toCurrency(cents: number): string {
  return currencyFormatter.format(cents / MONEY_SCALE);
}

function chartColors(size: number): string[] {
  const incomingPalette = [
    "#047857", // emerald-700
    "#059669", // emerald-600
    "#10b981", // emerald-500
    "#34d399", // emerald-400
    "#6366f1", // indigo-500
    "#4f46e5", // indigo-600
    "#4338ca", // indigo-700
    "#4f46e5", // indigo-600
    "#6366f1", // indigo-500
    "#34d399", // emerald-400
    "#10b981", // emerald-500
    "#059669", // emerald-600
  ];
  const outgoingPalette = [
    "#be123c", // rose-700
    "#e11d48", // rose-600
    "#f43f5e", // rose-500
    "#fb7185", // rose-400
    "#f97316", // orange-500
    "#f59e0b", // amber-500
    "#fbbf24", // amber-400
    "#f59e0b", // amber-500
    "#f97316", // orange-500
    "#fb7185", // rose-400
    "#f43f5e", // rose-500
    "#e11d48", // rose-600
  ];
  const palette = props.variant === "outgoing" ? outgoingPalette : incomingPalette;
  return Array.from({ length: size }, (_, idx) => palette[idx % palette.length]);
}

const sortedEntries = computed(() =>
  [...props.data].sort((a, b) => b.totalCents - a.totalCents),
);

const totalCents = computed(() =>
  sortedEntries.value.reduce((sum, entry) => sum + entry.totalCents, 0),
);

function formatPercentage(value: number): string {
  const rounded = Math.round(value * 10) / 10;
  return `${Number.isInteger(rounded) ? rounded.toFixed(0) : rounded.toFixed(1)}%`;
}

const topLabelEntries = computed(() => {
  const palette = chartColors(sortedEntries.value.length);
  const total = totalCents.value;
  return sortedEntries.value.slice(0, MAX_VISIBLE_LABELS).map((entry, index) => ({
    ...entry,
    color: palette[index],
    percentageLabel: total > 0 ? formatPercentage((entry.totalCents / total) * 100) : "0%",
  }));
});

const allLabelEntries = computed(() => {
  const palette = chartColors(sortedEntries.value.length);
  const total = totalCents.value;
  return sortedEntries.value.map((entry, index) => ({
    ...entry,
    color: palette[index],
    percentageLabel: total > 0 ? formatPercentage((entry.totalCents / total) * 100) : "0%",
  }));
});

const hiddenLabelCount = computed(() => Math.max(0, sortedEntries.value.length - MAX_VISIBLE_LABELS));

function selectTag(tag: string): void {
  emit("tag-selected", { tag, variant: props.variant });
}

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
      labels: sortedEntries.value.map((entry) => entry.tag),
      datasets: [
        {
          data: sortedEntries.value.map((entry) => entry.totalCents),
          backgroundColor: chartColors(sortedEntries.value.length),
          borderColor: "#ffffff",
          borderWidth: 1,
          hoverOffset: 3,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      onClick: (_event, elements) => {
        if (!elements || elements.length === 0) {
          return;
        }
        const selected = sortedEntries.value[elements[0].index];
        if (!selected) {
          return;
        }
        selectTag(selected.tag);
      },
      plugins: {
        legend: {
          display: false,
        },
        tooltip: {
          backgroundColor: chartTheme.tooltipBg,
          titleColor: "#f8fafc",
          bodyColor: "#f8fafc",
          borderColor: "rgba(148, 163, 184, 0.35)",
          borderWidth: 1,
          callbacks: {
            label: (ctx) => {
              const value = Number(ctx.parsed);
              const percent = totalCents.value > 0 ? (value / totalCents.value) * 100 : 0;
              return `${ctx.label}: ${toCurrency(value)} (${formatPercentage(percent)})`;
            },
          },
        },
      },
      cutout: "58%",
    },
  });
}

onMounted(async () => {
  await nextTick();
  renderChart();
});

watch(
  () => [props.data, props.loading],
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
      <canvas
        ref="canvasRef"
        class="h-full w-full"
        aria-label="Cashflow tag distribution chart"
      />
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
        No data yet
      </div>
    </div>

    <div
      v-if="!loading && hasData"
      class="mt-2 flex flex-wrap items-center gap-2"
    >
      <button
        v-for="entry in topLabelEntries"
        :key="entry.tag"
        type="button"
        class="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-[11px] font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-100"
        :title="`Filter on ${entry.tag}`"
        @click="selectTag(entry.tag)"
      >
        <span
          class="h-1.5 w-1.5 rounded-full"
          :style="{ backgroundColor: entry.color }"
        />
        <span>{{ entry.tag }} {{ entry.percentageLabel }}</span>
      </button>
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
            title="Show all tags"
            @click="toggle"
          >
            <span>+{{ hiddenLabelCount }} more tags</span>
            <InformationCircleIcon class="h-3.5 w-3.5" />
          </button>
        </template>
        <template #default="{ close }">
          <p class="mb-2 text-[11px] font-semibold uppercase tracking-wide text-slate-500">All tags</p>
          <div class="max-h-56 space-y-1 overflow-auto pr-1">
            <button
              v-for="entry in allLabelEntries"
              :key="`all-${entry.tag}`"
              type="button"
              class="flex w-full items-center justify-between gap-2 rounded-md bg-slate-50 px-2 py-1.5 text-left text-xs text-slate-700 transition hover:bg-slate-100"
              :title="`Filter on ${entry.tag}`"
              @click="
                selectTag(entry.tag);
                close();
              "
            >
              <span class="inline-flex min-w-0 items-center gap-1.5">
                <span class="h-1.5 w-1.5 shrink-0 rounded-full" :style="{ backgroundColor: entry.color }" />
                <span class="truncate">{{ entry.tag }}</span>
              </span>
              <span class="shrink-0 font-semibold text-slate-600">{{ entry.percentageLabel }}</span>
            </button>
          </div>
        </template>
      </BasePopover>
    </div>
  </div>
</template>

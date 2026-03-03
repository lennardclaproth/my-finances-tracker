<script setup lang="ts">
import { Chart } from "chart.js/auto";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { PortfolioGrowthPoint } from "../../../types/portfolio";
import {
  chartTheme,
  gradientFillWithAlpha,
  horizontalGridConfig,
  horizontalGuidePlugin,
  HOVER_MARKER_BORDER_COLOR,
  HOVER_MARKER_BORDER_WIDTH,
  HOVER_MARKER_FILL_COLOR,
  HOVER_MARKER_SIZE,
  hoverGuidePlugin,
  selectionIndicatorStyles,
  selectionLineStyle,
  toCanvasX,
} from "./chartHelpers";

interface Props {
  loading?: boolean;
  data?: PortfolioGrowthPoint[];
}

interface RangeSelection {
  from: string;
  to: string;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  data: () => [],
});

const emit = defineEmits<{
  (event: "range-selected", value: RangeSelection): void;
}>();

const hasData = computed(() => props.data.length > 0);
const canvasRef = ref<HTMLCanvasElement | null>(null);
const isSelecting = ref(false);
const pointerDownX = ref(0);
const pointerCurrentX = ref(0);
const pointerDownIndex = ref<number | null>(null);
const anchorIndex = ref<number | null>(null);
const hoverIndex = ref<number | null>(null);
let chart: Chart<"bar" | "line"> | null = null;

const trendColors = {
  dailyPositive: "#059669", // emerald-600
  dailyNegative: "#dc2626", // red-600
  cumulative: "#334155", // slate-700
};
const legendEntries = [
  { key: "daily", label: "Daily TWR", color: trendColors.dailyPositive },
  { key: "cumulative", label: "Return vs Cost Basis", color: trendColors.cumulative },
];

function formatPercent(value: number): string {
  return `${value.toFixed(2)}%`;
}

function formatDate(input: string): string {
  const date = new Date(input);
  if (Number.isNaN(date.getTime())) {
    return input;
  }
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function destroyChart(): void {
  chart?.destroy();
  chart = null;
}

function toIsoDateUtc(date: Date): string {
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  const day = String(date.getUTCDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function clampIndex(index: number): number {
  if (props.data.length === 0) {
    return 0;
  }
  return Math.max(0, Math.min(props.data.length - 1, index));
}

function pixelToIndex(canvasX: number): number | null {
  if (!chart || props.data.length === 0) {
    return null;
  }
  const xScale = chart.scales.x;
  const maybeIndex = xScale.getValueForPixel(canvasX);
  if (typeof maybeIndex !== "number" || Number.isNaN(maybeIndex)) {
    return null;
  }
  return clampIndex(Math.round(maybeIndex));
}

function dragStartPointY(index: number): number | null {
  if (!chart) {
    return null;
  }
  const point = props.data[clampIndex(index)];
  if (!point) {
    return null;
  }
  return chart.scales.y.getPixelForValue(point.returnVsCostBasisPct);
}

function pointX(index: number): number | null {
  if (!chart) {
    return null;
  }
  return chart.scales.x.getPixelForValue(index);
}

function pointDate(index: number): Date | null {
  const point = props.data[clampIndex(index)];
  if (!point) {
    return null;
  }
  const date = new Date(point.occurredAt);
  if (Number.isNaN(date.getTime())) {
    return null;
  }
  return date;
}

function emitSingleSelection(index: number): void {
  const date = pointDate(index);
  if (!date) {
    return;
  }
  const iso = toIsoDateUtc(date);
  emit("range-selected", { from: iso, to: iso });
}

function emitRangeSelection(startIndex: number, endIndex: number): void {
  const leftDate = pointDate(Math.min(startIndex, endIndex));
  const rightDate = pointDate(Math.max(startIndex, endIndex));
  if (!leftDate || !rightDate) {
    return;
  }
  emit("range-selected", {
    from: toIsoDateUtc(leftDate),
    to: toIsoDateUtc(rightDate),
  });
}

function dragOverlayStyle(): Record<string, string> {
  if (!isSelecting.value || anchorIndex.value === null || hoverIndex.value === null || !chart) {
    return { display: "none" };
  }
  const anchorX = pointX(anchorIndex.value);
  const targetX = pointX(hoverIndex.value);
  const anchorY = dragStartPointY(anchorIndex.value);
  if (anchorX === null || targetX === null || anchorY === null) {
    return { display: "none" };
  }
  return selectionIndicatorStyles(anchorX, targetX, anchorY, chart.chartArea.top, chart.chartArea.bottom - chart.chartArea.top).band;
}

function dragStartMarkerStyle(): Record<string, string> {
  if (!isSelecting.value || anchorIndex.value === null || !chart) {
    return { display: "none" };
  }
  const anchorX = pointX(anchorIndex.value);
  const anchorY = dragStartPointY(anchorIndex.value);
  if (anchorX === null || anchorY === null) {
    return { display: "none" };
  }
  return selectionIndicatorStyles(anchorX, anchorX, anchorY, 0, 0).marker;
}

function dragHoverMarkerStyle(): Record<string, string> {
  if (!isSelecting.value || hoverIndex.value === null || !chart) {
    return { display: "none" };
  }
  const hoverX = pointX(hoverIndex.value);
  const hoverY = dragStartPointY(hoverIndex.value);
  if (hoverX === null || hoverY === null) {
    return { display: "none" };
  }
  return selectionIndicatorStyles(hoverX, hoverX, hoverY, 0, 0).marker;
}

function dragStartLineStyle(): Record<string, string> {
  if (!isSelecting.value || anchorIndex.value === null || !chart) {
    return { display: "none" };
  }
  const x = pointX(anchorIndex.value);
  if (x === null) {
    return { display: "none" };
  }
  return selectionLineStyle(x, chart.chartArea.top, chart.chartArea.bottom - chart.chartArea.top);
}

function dragHoverLineStyle(): Record<string, string> {
  if (!isSelecting.value || hoverIndex.value === null || !chart) {
    return { display: "none" };
  }
  const x = pointX(hoverIndex.value);
  if (x === null) {
    return { display: "none" };
  }
  return selectionLineStyle(x, chart.chartArea.top, chart.chartArea.bottom - chart.chartArea.top);
}

function renderChart(): void {
  if (!canvasRef.value || props.loading || !hasData.value) {
    destroyChart();
    return;
  }

  destroyChart();
  const cumulativeValues = props.data.map((p) => p.returnVsCostBasisPct);
  const positiveFillValues = cumulativeValues.map((value) => (value >= 0 ? value : null));
  const negativeFillValues = cumulativeValues.map((value) => (value < 0 ? value : null));

  chart = new Chart(canvasRef.value, {
    plugins: [horizontalGuidePlugin, hoverGuidePlugin],
    type: "bar",
    data: {
      labels: props.data.map((p) => formatDate(p.occurredAt)),
      datasets: [
        {
          type: "bar",
          label: "Time-Weighted Return (%)",
          data: props.data.map((p) => p.timeWeightedReturnPct),
          backgroundColor: (ctx) => {
            const value = Number(ctx.raw ?? 0);
            return value >= 0 ? "rgba(16, 185, 129, 0.38)" : "rgba(239, 68, 68, 0.38)";
          },
          borderColor: (ctx) => {
            const value = Number(ctx.raw ?? 0);
            return value >= 0 ? "rgba(5, 150, 105, 0.95)" : "rgba(220, 38, 38, 0.95)";
          },
          borderWidth: 1,
          barPercentage: 0.74,
          categoryPercentage: 0.86,
          borderRadius: (ctx) => {
            const value = Number(ctx.raw ?? 0);
            if (value >= 0) {
              return { topLeft: 4, topRight: 4, bottomLeft: 0, bottomRight: 0 };
            }
            return { topLeft: 0, topRight: 0, bottomLeft: 4, bottomRight: 4 };
          },
          borderSkipped: false,
        },
        {
          type: "line",
          label: "Return vs Cost Basis %",
          data: cumulativeValues,
          borderColor: trendColors.cumulative,
          borderWidth: 2,
          tension: 0.48,
          cubicInterpolationMode: "monotone",
          pointRadius: 0,
          pointHoverRadius: HOVER_MARKER_SIZE / 2,
          pointHoverBackgroundColor: HOVER_MARKER_FILL_COLOR,
          pointHoverBorderColor: HOVER_MARKER_BORDER_COLOR,
          pointHoverBorderWidth: HOVER_MARKER_BORDER_WIDTH,
          pointHitRadius: 10,
          fill: false,
          segment: {
            borderColor: (ctx) => {
              const y0 = Number(ctx.p0.parsed.y ?? 0);
              const y1 = Number(ctx.p1.parsed.y ?? 0);
              return y0 < 0 || y1 < 0 ? trendColors.dailyNegative : trendColors.cumulative;
            },
          },
          order: 3,
        },
        {
          type: "line",
          label: "Positive Area",
          data: positiveFillValues,
          borderColor: "rgba(0, 0, 0, 0)",
          backgroundColor: (ctx) => gradientFillWithAlpha(ctx.chart, trendColors.cumulative, 0.2, 0.02),
          borderWidth: 0,
          pointRadius: 0,
          pointHoverRadius: 0,
          pointHitRadius: 0,
          tension: 0.48,
          cubicInterpolationMode: "monotone",
          fill: "origin",
          order: 2,
        },
        {
          type: "line",
          label: "Negative Area",
          data: negativeFillValues,
          borderColor: "rgba(0, 0, 0, 0)",
          backgroundColor: (ctx) => gradientFillWithAlpha(ctx.chart, trendColors.dailyNegative, 0.02, 0.2),
          borderWidth: 0,
          pointRadius: 0,
          pointHoverRadius: 0,
          pointHitRadius: 0,
          tension: 0.48,
          cubicInterpolationMode: "monotone",
          fill: "origin",
          order: 2,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      interaction: {
        mode: "index",
        intersect: false,
      },
      scales: {
        x: {
          ticks: {
            color: chartTheme.muted,
            autoSkip: true,
            maxTicksLimit: 8,
            maxRotation: 0,
          },
          grid: {
            display: false,
          },
          border: {
            display: false,
          },
        },
        y: {
          ticks: {
            color: chartTheme.muted,
            callback: (value) => formatPercent(Number(value)),
          },
          grid: {
            ...horizontalGridConfig(),
          },
          border: {
            display: false,
          },
        },
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
          filter: (ctx) => ctx.datasetIndex <= 1,
          callbacks: {
            label: (ctx) => `${ctx.dataset.label}: ${formatPercent(Number(ctx.parsed.y))}`,
          },
        },
      },
    },
  });
}

function onPointerDown(event: PointerEvent): void {
  if (!hasData.value || props.loading || !chart || !canvasRef.value) {
    return;
  }
  if (event.button !== 0) {
    return;
  }
  const startX = toCanvasX(chart, canvasRef.value, event.clientX);
  const startIndex = pixelToIndex(startX);
  if (startIndex === null) {
    return;
  }
  const startY = dragStartPointY(startIndex);
  if (startY === null) {
    return;
  }
  pointerDownIndex.value = startIndex;
  anchorIndex.value = startIndex;
  hoverIndex.value = startIndex;
  isSelecting.value = true;
  (chart as Chart & { __suspendHoverGuide?: boolean }).__suspendHoverGuide = true;
  pointerDownX.value = startX;
  pointerCurrentX.value = startX;
  (event.currentTarget as HTMLElement | null)?.setPointerCapture?.(event.pointerId);
}

function onPointerMove(event: PointerEvent): void {
  if (!isSelecting.value || !chart || !canvasRef.value) {
    return;
  }
  pointerCurrentX.value = toCanvasX(chart, canvasRef.value, event.clientX);
  const nextIndex = pixelToIndex(pointerCurrentX.value);
  if (nextIndex !== null) {
    hoverIndex.value = nextIndex;
  }
}

function onPointerUp(event: PointerEvent): void {
  if (!isSelecting.value || !chart || !canvasRef.value) {
    return;
  }

  pointerCurrentX.value = toCanvasX(chart, canvasRef.value, event.clientX);
  const distance = Math.abs(pointerCurrentX.value - pointerDownX.value);
  const startIndex = anchorIndex.value ?? pointerDownIndex.value;
  const endIndex = hoverIndex.value ?? pointerDownIndex.value ?? startIndex;
  isSelecting.value = false;
  (chart as Chart & { __suspendHoverGuide?: boolean }).__suspendHoverGuide = false;

  if (startIndex === null || endIndex === null) {
    pointerDownIndex.value = null;
    return;
  }

  if (distance < 6) {
    const nextAnchor = pointerDownIndex.value ?? startIndex;
    emitSingleSelection(nextAnchor);
    anchorIndex.value = null;
    hoverIndex.value = null;
    pointerDownIndex.value = null;
    return;
  }

  if (endIndex === startIndex) {
    anchorIndex.value = null;
    hoverIndex.value = null;
    pointerDownIndex.value = null;
    return;
  }

  emitRangeSelection(startIndex, endIndex);
  anchorIndex.value = null;
  hoverIndex.value = null;
  pointerDownIndex.value = null;
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
      <canvas
        ref="canvasRef"
        class="h-full w-full"
        aria-label="Portfolio growth chart"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
        @pointerleave="onPointerUp"
      />

      <div
        v-if="isSelecting"
        class="pointer-events-none absolute z-10 bg-indigo-300/25"
        :style="dragOverlayStyle()"
      />
      <div
        v-if="isSelecting"
        class="pointer-events-none absolute z-10 border-l border-dashed border-slate-400/80"
        :style="dragStartLineStyle()"
      />
      <div
        v-if="isSelecting"
        class="pointer-events-none absolute z-10 border-l border-dashed border-slate-400/80"
        :style="dragHoverLineStyle()"
      />
      <div
        v-if="isSelecting"
        class="pointer-events-none absolute z-10 rounded-full border-2 border-indigo-500 bg-white shadow-sm"
        :style="dragStartMarkerStyle()"
      />
      <div
        v-if="isSelecting"
        class="pointer-events-none absolute z-10 rounded-full border-2 border-indigo-500 bg-white shadow-sm"
        :style="dragHoverMarkerStyle()"
      />

      <div v-if="loading" class="absolute inset-0 bg-white/90 p-3">
        <div class="flex h-full w-full flex-col gap-2 animate-pulse">
          <div class="relative min-h-0 flex-1">
            <div class="absolute inset-0 flex items-end gap-2 px-2 pb-2">
              <div class="h-[30%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[50%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[42%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[64%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[58%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[75%] w-full rounded-md bg-slate-200/90" />
            </div>
            <svg
              viewBox="0 0 100 100"
              preserveAspectRatio="none"
              class="absolute inset-0 h-full w-full px-1 py-1 text-slate-300/95"
              aria-hidden="true"
            >
              <polyline
                points="4,62 20,56 36,58 52,49 68,52 84,41 96,38"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </div>
          <div class="flex flex-wrap items-center gap-2 px-1 pb-1">
            <div class="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-2 py-1">
              <span class="h-1.5 w-1.5 rounded-full bg-slate-300" />
              <span class="h-3 w-16 rounded bg-slate-200" />
            </div>
            <div class="inline-flex items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-2 py-1">
              <span class="h-1.5 w-1.5 rounded-full bg-slate-300" />
              <span class="h-3 w-24 rounded bg-slate-200" />
            </div>
          </div>
        </div>
      </div>

      <div
        v-else-if="!hasData"
        class="absolute inset-0 flex items-center justify-center bg-white/75 text-sm text-slate-500"
      >
        No portfolio snapshots yet.
      </div>
    </div>

    <div
      v-if="!loading && hasData"
      class="mt-2 flex flex-wrap items-center gap-2"
    >
      <span
        v-for="entry in legendEntries"
        :key="entry.key"
        class="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-[11px] font-medium text-slate-700"
      >
        <span
          class="h-1.5 w-1.5 rounded-full"
          :style="{ backgroundColor: entry.color }"
        />
        <span>{{ entry.label }}</span>
      </span>
    </div>
  </div>
</template>

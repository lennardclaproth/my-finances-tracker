<script setup lang="ts">
import { Chart } from "chart.js/auto";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { AssetGrowthPoint } from "../../../types/assets";
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
  data?: AssetGrowthPoint[];
  secondaryData?: AssetGrowthPoint[];
  seriesLabel?: string;
  secondarySeriesLabel?: string;
}

interface RangeSelection {
  from: string;
  to: string;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  data: () => [],
  secondaryData: () => [],
  seriesLabel: "Total Assets Worth",
  secondarySeriesLabel: "Portfolio Worth",
});

const emit = defineEmits<{
  (event: "range-selected", value: RangeSelection): void;
}>();

const canvasRef = ref<HTMLCanvasElement | null>(null);
const hasData = computed(() => props.data.length > 0 || props.secondaryData.length > 0);
const seriesLabel = computed(() => props.seriesLabel);
const secondarySeriesLabel = computed(() => props.secondarySeriesLabel);
const hasSecondarySeries = computed(() => props.secondaryData.length > 0);
const isSelecting = ref(false);
const pointerDownX = ref(0);
const pointerCurrentX = ref(0);
const pointerDownIndex = ref<number | null>(null);
const anchorIndex = ref<number | null>(null);
const hoverIndex = ref<number | null>(null);
const chartDates = ref<string[]>([]);
const chartPrimaryValues = ref<number[]>([]);
let chart: Chart<"line"> | null = null;
const lineColor = "#334155"; // slate-700
const negativeLineColor = "#dc2626"; // red-600
const secondaryLineColor = "#334155"; // slate-700

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 0,
});

function formatDate(value: string): string {
  const date = new Date(`${value}T00:00:00Z`);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "2-digit",
    year: "numeric",
    timeZone: "UTC",
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
  if (chartDates.value.length === 0) {
    return 0;
  }
  return Math.max(0, Math.min(chartDates.value.length - 1, index));
}

function pixelToIndex(canvasX: number): number | null {
  if (!chart || chartDates.value.length === 0) {
    return null;
  }
  const xScale = chart.scales.x;
  const maybeIndex = xScale.getValueForPixel(canvasX);
  if (typeof maybeIndex !== "number" || Number.isNaN(maybeIndex)) {
    return null;
  }
  return clampIndex(Math.round(maybeIndex));
}

function pointWorth(index: number): number | null {
  const point = chartPrimaryValues.value[clampIndex(index)];
  if (point === undefined) {
    return null;
  }
  return point;
}

function dragStartPointY(index: number): number | null {
  if (!chart) {
    return null;
  }
  const worth = pointWorth(index);
  if (worth === null) {
    return null;
  }
  return chart.scales.y.getPixelForValue(worth);
}

function pointX(index: number): number | null {
  if (!chart) {
    return null;
  }
  return chart.scales.x.getPixelForValue(index);
}

function pointDate(index: number): Date | null {
  const dateValue = chartDates.value[clampIndex(index)];
  if (!dateValue) {
    return null;
  }
  const date = new Date(`${dateValue}T00:00:00Z`);
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

function alignedSeries(series: AssetGrowthPoint[], dates: string[]): number[] {
  const sorted = [...series].sort((a, b) => a.date.localeCompare(b.date));
  const values: number[] = [];
  let index = 0;
  let lastValue = 0;

  for (const date of dates) {
    while (index < sorted.length && sorted[index].date <= date) {
      const parsed = Number.parseFloat(sorted[index].totalWorth);
      lastValue = Number.isNaN(parsed) ? lastValue : parsed;
      index += 1;
    }
    values.push(lastValue);
  }
  return values;
}

function renderChart(): void {
  if (props.loading || !canvasRef.value || !hasData.value) {
    chartDates.value = [];
    chartPrimaryValues.value = [];
    isSelecting.value = false;
    pointerDownIndex.value = null;
    anchorIndex.value = null;
    hoverIndex.value = null;
    destroyChart();
    return;
  }

  const dateSet = new Set<string>();
  for (const point of props.data) {
    dateSet.add(point.date);
  }
  for (const point of props.secondaryData) {
    dateSet.add(point.date);
  }
  const dates = Array.from(dateSet).sort((a, b) => a.localeCompare(b));

  const primaryValues = alignedSeries(props.data, dates);
  const secondaryValues = alignedSeries(props.secondaryData, dates);
  chartDates.value = dates;
  chartPrimaryValues.value = primaryValues;
  const positiveFillValues = primaryValues.map((value) => (value >= 0 ? value : null));
  const negativeFillValues = primaryValues.map((value) => (value < 0 ? value : null));

  destroyChart();
  chart = new Chart(canvasRef.value, {
    plugins: [horizontalGuidePlugin, hoverGuidePlugin],
    type: "line",
    data: {
      labels: dates.map((date) => formatDate(date)),
      datasets: [
        {
          label: props.seriesLabel,
          data: primaryValues,
          borderColor: lineColor,
          fill: false,
          borderWidth: 2,
          tension: 0.48,
          cubicInterpolationMode: "monotone",
          pointRadius: 0,
          pointHoverRadius: HOVER_MARKER_SIZE / 2,
          pointHoverBackgroundColor: HOVER_MARKER_FILL_COLOR,
          pointHoverBorderColor: HOVER_MARKER_BORDER_COLOR,
          pointHoverBorderWidth: HOVER_MARKER_BORDER_WIDTH,
          pointHitRadius: 12,
          segment: {
            borderColor: (ctx) => {
              const y0 = Number(ctx.p0.parsed.y ?? 0);
              const y1 = Number(ctx.p1.parsed.y ?? 0);
              return y0 < 0 || y1 < 0 ? negativeLineColor : lineColor;
            },
          },
          order: 3,
        },
        {
          label: "Positive Area",
          data: positiveFillValues,
          borderColor: "rgba(0, 0, 0, 0)",
          backgroundColor: (ctx) => gradientFillWithAlpha(ctx.chart, lineColor, 0.2, 0.02),
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
          label: "Negative Area",
          data: negativeFillValues,
          borderColor: "rgba(0, 0, 0, 0)",
          backgroundColor: (ctx) => gradientFillWithAlpha(ctx.chart, negativeLineColor, 0.02, 0.2),
          borderWidth: 0,
          pointRadius: 0,
          pointHoverRadius: 0,
          pointHitRadius: 0,
          tension: 0.48,
          cubicInterpolationMode: "monotone",
          fill: "origin",
          order: 2,
        },
        ...(hasSecondarySeries.value
          ? [
              {
                label: props.secondarySeriesLabel,
                data: secondaryValues,
                borderColor: secondaryLineColor,
                backgroundColor: "rgba(51, 65, 85, 0.05)",
                fill: false,
                borderWidth: 2,
                tension: 0.48,
                cubicInterpolationMode: "monotone" as const,
                pointRadius: 0,
                pointHoverRadius: HOVER_MARKER_SIZE / 2,
                pointHoverBackgroundColor: HOVER_MARKER_FILL_COLOR,
                pointHoverBorderColor: HOVER_MARKER_BORDER_COLOR,
                pointHoverBorderWidth: HOVER_MARKER_BORDER_WIDTH,
                pointHitRadius: 12,
                order: 4,
              },
            ]
          : []),
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
            maxRotation: 0,
            autoSkip: true,
            maxTicksLimit: 6,
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
            callback: (value) => currencyFormatter.format(Number(value)),
            maxTicksLimit: 4,
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
          callbacks: {
            label: (ctx) => `${ctx.dataset.label}: ${currencyFormatter.format(Number(ctx.parsed.y))}`,
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
  () => [props.data, props.secondaryData, props.loading, props.seriesLabel, props.secondarySeriesLabel],
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
        aria-label="Asset growth chart"
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
      <div
        v-if="loading"
        class="absolute inset-0 bg-white/90 p-3"
      >
        <div class="flex h-full w-full flex-col gap-2 animate-pulse">
          <div class="relative min-h-0 flex-1">
            <div class="absolute inset-0 flex items-end gap-2 px-2 pb-2">
              <div class="h-[28%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[46%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[40%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[62%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[54%] w-full rounded-md bg-slate-200/90" />
              <div class="h-[72%] w-full rounded-md bg-slate-200/90" />
            </div>
            <svg
              viewBox="0 0 100 100"
              preserveAspectRatio="none"
              class="absolute inset-0 h-full w-full px-1 py-1 text-slate-300/95"
              aria-hidden="true"
            >
              <polyline
                points="4,66 20,60 36,58 52,50 68,49 84,40 96,36"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </div>
          <div class="inline-flex w-fit items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5">
            <span class="h-1.5 w-1.5 rounded-full bg-slate-300" />
            <span class="h-3 w-24 rounded bg-slate-200" />
          </div>
        </div>
      </div>
      <div
        v-else-if="!hasData"
        class="absolute inset-0 flex items-center justify-center bg-white/75 text-sm text-slate-500"
      >
        No growth history yet
      </div>
    </div>

    <div
      v-if="!loading && hasData"
      class="mt-2 flex flex-wrap items-center gap-2"
    >
      <span class="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-[11px] font-medium text-slate-700">
        <span class="h-1.5 w-1.5 rounded-full bg-slate-700" />
        <span>{{ seriesLabel }}</span>
      </span>
      <span
        v-if="hasSecondarySeries"
        class="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-[11px] font-medium text-slate-700"
      >
        <span class="h-1.5 w-1.5 rounded-full bg-slate-700" />
        <span>{{ secondarySeriesLabel }}</span>
      </span>
    </div>
  </div>
</template>

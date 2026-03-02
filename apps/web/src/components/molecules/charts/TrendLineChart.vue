<script setup lang="ts">
import { Chart } from "chart.js/auto";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { CashflowMonthlyAnalyticsPoint } from "../../../types/cashflow";
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
  data?: CashflowMonthlyAnalyticsPoint[];
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

const canvasRef = ref<HTMLCanvasElement | null>(null);
const hasData = computed(() => props.data.length > 0);

const isSelecting = ref(false);
const pointerDownX = ref(0);
const pointerCurrentX = ref(0);
const pointerDownIndex = ref<number | null>(null);
const anchorIndex = ref<number | null>(null);
const hoverIndex = ref<number | null>(null);

let chart: Chart<"line"> | null = null;

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 0,
});
const MONEY_SCALE = 1_000_000;
const trendColors = {
  incoming: "#059669", // emerald-600
  outgoing: "#e11d48", // rose-600
  net: "#4f46e5", // indigo-600
};
const trendTagEntries = [
  { key: "incoming", label: "Incoming", color: trendColors.incoming },
  { key: "outgoing", label: "Outgoing", color: trendColors.outgoing },
  { key: "net", label: "Net", color: trendColors.net },
];

function toCurrency(amountCents: number): string {
  return currencyFormatter.format(amountCents / MONEY_SCALE);
}

function toIsoDateUtc(date: Date): string {
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  const day = String(date.getUTCDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function monthStartAndEnd(monthStartIso: string): RangeSelection | null {
  const monthStart = new Date(`${monthStartIso}T00:00:00Z`);
  if (Number.isNaN(monthStart.getTime())) {
    return null;
  }
  const monthEnd = new Date(Date.UTC(monthStart.getUTCFullYear(), monthStart.getUTCMonth() + 1, 0));
  return {
    from: toIsoDateUtc(monthStart),
    to: toIsoDateUtc(monthEnd),
  };
}

function monthLabel(value: string): string {
  const asDate = new Date(`${value}T00:00:00Z`);
  if (Number.isNaN(asDate.getTime())) {
    return value;
  }
  return asDate.toLocaleDateString("en-US", {
    month: "short",
    year: "numeric",
    timeZone: "UTC",
  });
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
  return chart.scales.y.getPixelForValue(point.netCents);
}

function pointX(index: number): number | null {
  if (!chart) {
    return null;
  }
  return chart.scales.x.getPixelForValue(index);
}

function emitSingleMonthSelection(index: number): void {
  const safeIndex = clampIndex(index);
  const month = props.data[safeIndex]?.month;
  if (!month) {
    return;
  }
  const range = monthStartAndEnd(month);
  if (!range) {
    return;
  }
  emit("range-selected", range);
}

function emitRangeSelection(startIndex: number, endIndex: number): void {
  const left = Math.min(startIndex, endIndex);
  const right = Math.max(startIndex, endIndex);

  const startMonth = props.data[left]?.month;
  const endMonth = props.data[right]?.month;
  if (!startMonth || !endMonth) {
    return;
  }

  const startRange = monthStartAndEnd(startMonth);
  const endRange = monthStartAndEnd(endMonth);
  if (!startRange || !endRange) {
    return;
  }

  emit("range-selected", {
    from: startRange.from,
    to: endRange.to,
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
    plugins: [horizontalGuidePlugin, hoverGuidePlugin],
    type: "line",
    data: {
      labels: props.data.map((point) => monthLabel(point.month)),
      datasets: [
        {
          label: "Incoming",
          data: props.data.map((point) => point.incomingCents),
          borderColor: trendColors.incoming,
          backgroundColor: (ctx) => gradientFillWithAlpha(ctx.chart, trendColors.incoming, 0.08, 0.01),
          tension: 0.3,
          fill: "origin",
          borderWidth: 2,
          pointRadius: 0,
          pointHoverRadius: HOVER_MARKER_SIZE / 2,
          pointHoverBackgroundColor: HOVER_MARKER_FILL_COLOR,
          pointHoverBorderColor: HOVER_MARKER_BORDER_COLOR,
          pointHoverBorderWidth: HOVER_MARKER_BORDER_WIDTH,
          pointHitRadius: 14,
          order: 1,
        },
        {
          label: "Outgoing",
          data: props.data.map((point) => point.outgoingCents),
          borderColor: trendColors.outgoing,
          backgroundColor: (ctx) => gradientFillWithAlpha(ctx.chart, trendColors.outgoing, 0.1, 0.01),
          tension: 0.3,
          fill: "origin",
          borderWidth: 2,
          pointRadius: 0,
          pointHoverRadius: HOVER_MARKER_SIZE / 2,
          pointHoverBackgroundColor: HOVER_MARKER_FILL_COLOR,
          pointHoverBorderColor: HOVER_MARKER_BORDER_COLOR,
          pointHoverBorderWidth: HOVER_MARKER_BORDER_WIDTH,
          pointHitRadius: 14,
          order: 2,
        },
        {
          label: "Net",
          data: props.data.map((point) => point.netCents),
          borderColor: trendColors.net,
          backgroundColor: (ctx) => gradientFillWithAlpha(ctx.chart, trendColors.net, 0.2, 0.02),
          tension: 0.3,
          fill: "origin",
          borderWidth: 2,
          pointRadius: 0,
          pointHoverRadius: HOVER_MARKER_SIZE / 2,
          pointHoverBackgroundColor: HOVER_MARKER_FILL_COLOR,
          pointHoverBorderColor: HOVER_MARKER_BORDER_COLOR,
          pointHoverBorderWidth: HOVER_MARKER_BORDER_WIDTH,
          pointHitRadius: 14,
          order: 3,
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
            callback: (value) => toCurrency(Number(value)),
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
          displayColors: true,
          callbacks: {
            label: (ctx) => `${ctx.dataset.label}: ${toCurrency(Number(ctx.parsed.y))}`,
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
    emitSingleMonthSelection(nextAnchor);
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
        aria-label="Cashflow chart"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
        @pointerleave="onPointerUp"
      />

      <div
        v-if="isSelecting"
        class="pointer-events-none absolute z-10 bg-blue-300/25"
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
        class="pointer-events-none absolute z-10 rounded-full border-2 border-blue-500 bg-white shadow-sm"
        :style="dragStartMarkerStyle()"
      />
      <div
        v-if="isSelecting"
        class="pointer-events-none absolute z-10 rounded-full border-2 border-blue-500 bg-white shadow-sm"
        :style="dragHoverMarkerStyle()"
      />

      <div
        v-if="loading"
        class="absolute inset-0 bg-white/90 p-3"
      >
        <div class="flex h-full w-full flex-col gap-3">
          <div class="h-4 w-28 animate-pulse rounded-full bg-slate-200" />
          <div class="flex-1 rounded-2xl border border-slate-200 bg-slate-50 p-3">
            <div class="flex h-full items-end gap-2">
              <div class="h-[35%] w-full animate-pulse rounded-md bg-slate-200/90" />
              <div class="h-[55%] w-full animate-pulse rounded-md bg-slate-200/90" />
              <div class="h-[48%] w-full animate-pulse rounded-md bg-slate-200/90" />
              <div class="h-[68%] w-full animate-pulse rounded-md bg-slate-200/90" />
              <div class="h-[52%] w-full animate-pulse rounded-md bg-slate-200/90" />
              <div class="h-[78%] w-full animate-pulse rounded-md bg-slate-200/90" />
            </div>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <div class="h-5 w-20 animate-pulse rounded-full bg-slate-200" />
            <div class="h-5 w-20 animate-pulse rounded-full bg-slate-200" />
            <div class="h-5 w-16 animate-pulse rounded-full bg-slate-200" />
          </div>
        </div>
      </div>

      <div
        v-else-if="!hasData"
        class="absolute inset-0 flex items-center justify-center bg-white/75 text-sm text-slate-500"
      >
        No cashflow data yet
      </div>
    </div>

    <div
      v-if="!loading && hasData"
      class="mt-2 flex flex-wrap items-center gap-2"
    >
      <span
        v-for="entry in trendTagEntries"
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

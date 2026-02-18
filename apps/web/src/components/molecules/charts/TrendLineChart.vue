<script setup lang="ts">
import { Chart } from "chart.js/auto";
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import type { CashflowMonthlyAnalyticsPoint } from "../../../types/cashflow";

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

const isDragging = ref(false);
const dragStartX = ref(0);
const dragCurrentX = ref(0);
let suppressClickUntil = 0;

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
const chartTheme = {
  text: "#3f3f46",
  muted: "#71717a",
  tooltipBg: "#0f172a",
};
const trendTagEntries = [
  { key: "incoming", label: "Incoming", color: trendColors.incoming },
  { key: "outgoing", label: "Outgoing", color: trendColors.outgoing },
  { key: "net", label: "Net", color: trendColors.net },
];

function toCurrency(amountCents: number): string {
  return currencyFormatter.format(amountCents / MONEY_SCALE);
}

function hexToRgba(hex: string, alpha: number): string {
  const normalized = hex.replace("#", "");
  const bigint = Number.parseInt(normalized, 16);
  const r = (bigint >> 16) & 255;
  const g = (bigint >> 8) & 255;
  const b = bigint & 255;
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

function gradientFill(chartInstance: Chart, color: string): CanvasGradient | string {
  return gradientFillWithAlpha(chartInstance, color, 0.24, 0.02);
}

function gradientFillWithAlpha(
  chartInstance: Chart,
  color: string,
  alphaTop: number,
  alphaBottom: number,
): CanvasGradient | string {
  const chartArea = chartInstance.chartArea;
  if (!chartArea) {
    return hexToRgba(color, alphaTop);
  }
  const gradient = chartInstance.ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
  gradient.addColorStop(0, hexToRgba(color, alphaTop));
  gradient.addColorStop(1, hexToRgba(color, alphaBottom));
  return gradient;
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

function pixelToIndex(clientX: number): number | null {
  if (!chart || !canvasRef.value || props.data.length === 0) {
    return null;
  }
  const rect = canvasRef.value.getBoundingClientRect();
  const xFromCanvas = clientX - rect.left;
  const xScale = chart.scales.x;
  const area = chart.chartArea;

  const clampedX = Math.max(area.left, Math.min(area.right, xFromCanvas));
  const maybeIndex = xScale.getValueForPixel(clampedX);
  if (typeof maybeIndex !== "number" || Number.isNaN(maybeIndex)) {
    return null;
  }
  return clampIndex(Math.round(maybeIndex));
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
  if (!isDragging.value || !canvasRef.value) {
    return { display: "none" };
  }
  const rect = canvasRef.value.getBoundingClientRect();
  const left = Math.min(dragStartX.value, dragCurrentX.value) - rect.left;
  const right = Math.max(dragStartX.value, dragCurrentX.value) - rect.left;
  return {
    left: `${Math.max(0, left)}px`,
    width: `${Math.max(0, right - left)}px`,
  };
}

function destroyChart(): void {
  chart?.destroy();
  chart = null;
}

function onChartClick(index: number): void {
  if (Date.now() < suppressClickUntil || isDragging.value) {
    return;
  }
  emitSingleMonthSelection(index);
}

function renderChart(): void {
  if (props.loading || !canvasRef.value || !hasData.value) {
    destroyChart();
    return;
  }

  destroyChart();

  chart = new Chart(canvasRef.value, {
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
          pointBackgroundColor: "#ffffff",
          pointBorderColor: trendColors.incoming,
          pointBorderWidth: 0,
          pointHoverRadius: 0,
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
          pointBackgroundColor: "#ffffff",
          pointBorderColor: trendColors.outgoing,
          pointBorderWidth: 0,
          pointHoverRadius: 0,
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
          pointBackgroundColor: "#ffffff",
          pointBorderColor: trendColors.net,
          pointBorderWidth: 0,
          pointHoverRadius: 0,
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
      onClick: (_event, elements) => {
        if (!elements || elements.length === 0) {
          return;
        }
        onChartClick(elements[0].index);
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
            display: false,
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
  if (!hasData.value || props.loading) {
    return;
  }
  if (event.button !== 0) {
    return;
  }
  isDragging.value = true;
  dragStartX.value = event.clientX;
  dragCurrentX.value = event.clientX;
  (event.currentTarget as HTMLElement | null)?.setPointerCapture?.(event.pointerId);
}

function onPointerMove(event: PointerEvent): void {
  if (!isDragging.value) {
    return;
  }
  dragCurrentX.value = event.clientX;
}

function onPointerUp(event: PointerEvent): void {
  if (!isDragging.value) {
    return;
  }

  dragCurrentX.value = event.clientX;
  const distance = Math.abs(dragCurrentX.value - dragStartX.value);
  isDragging.value = false;

  if (distance < 6) {
    return;
  }

  const startIndex = pixelToIndex(dragStartX.value);
  const endIndex = pixelToIndex(dragCurrentX.value);
  if (startIndex === null || endIndex === null) {
    return;
  }

  suppressClickUntil = Date.now() + 220;
  emitRangeSelection(startIndex, endIndex);
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
        aria-label="Cashflow trend chart"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
        @pointerleave="onPointerUp"
      />

      <div
        v-if="isDragging"
        class="pointer-events-none absolute bottom-0 top-0 z-10 border border-indigo-400 bg-indigo-300/20"
        :style="dragOverlayStyle()"
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
        No trend data yet
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

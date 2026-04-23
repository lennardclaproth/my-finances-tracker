<script setup lang="ts">
import { computed } from "vue";
import type { PortfolioPosition } from "../../types/portfolio";
import StatusBadge from "../atoms/StatusBadge.vue";
import BaseButton from "../atoms/BaseButton.vue";
import BaseToggle from "../atoms/BaseToggle.vue";
import UnrealizedPnLBadge from "../molecules/UnrealizedPnLBadge.vue";

interface Props {
  rows: PortfolioPosition[];
  loading?: boolean;
  errorMessage?: string;
  includeClosed?: boolean;
  framed?: boolean;
  showIncludeClosedControl?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  errorMessage: "",
  includeClosed: false,
  framed: true,
  showIncludeClosedControl: true,
});

const emit = defineEmits<{
  retry: [];
  "update:includeClosed": [value: boolean];
}>();

const skeletonRows = computed(() => Array.from({ length: 8 }, (_, index) => index));

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 2,
});
const MONEY_SCALE = 1_000_000;

function formatMoney(value?: number): string {
  if (value === undefined || value === null) {
    return "-";
  }
  return currencyFormatter.format(value / MONEY_SCALE);
}

function formatQuantity(value: number): string {
  return value.toLocaleString("en-US", {
    maximumFractionDigits: 6,
  });
}

function formatDate(value?: string): string {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return date.toLocaleDateString("en-US");
}

function averageCost(position: PortfolioPosition): number | undefined {
  if (position.quantity <= 0) {
    return undefined;
  }
  return position.costBasis / position.quantity;
}

const rootClasses = computed(() => {
  if (!props.framed) {
    return "h-full";
  }
  return "h-full overflow-hidden rounded-3xl border border-slate-300 bg-white/95 p-4 shadow-sm";
});
</script>

<template>
  <section :class="rootClasses">
    <div class="flex h-full min-h-0 flex-col">
      <header v-if="showIncludeClosedControl" class="mb-3 flex items-center justify-end gap-3">
        <div class="flex items-center gap-2">
          <span class="text-xs font-medium uppercase tracking-wide text-slate-500">Include closed</span>
          <BaseToggle
            :checked="includeClosed"
            :disabled="loading"
            @update:checked="emit('update:includeClosed', $event)"
          />
        </div>
      </header>

      <div class="min-h-0 flex-1 overflow-auto bg-slate-100">
        <table class="w-full min-w-[1020px] border-separate border-spacing-0 bg-white">
          <thead class="sticky top-0 z-20 bg-white/95 backdrop-blur">
            <tr>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Name</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Quantity</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Avg Cost</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Market Value</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Realized</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Unrealized %</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Last Snapshot</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Status</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loading">
              <tr v-for="index in skeletonRows" :key="`skeleton-${index}`">
                <td v-for="cell in 8" :key="`skeleton-${index}-${cell}`" class="border-b border-slate-100 px-3 py-3">
                  <div class="h-4 w-full animate-pulse rounded bg-slate-200" />
                </td>
              </tr>
            </template>

            <tr v-else-if="errorMessage">
              <td colspan="8" class="px-3 py-10 text-center">
                <p class="mb-3 text-sm text-rose-700">{{ errorMessage }}</p>
                <BaseButton size="sm" variant="secondary" @click="emit('retry')">Retry</BaseButton>
              </td>
            </tr>

            <tr v-else-if="rows.length === 0">
              <td colspan="8" class="px-3 py-10 text-center text-sm text-slate-500">
                No positions found for this account.
              </td>
            </tr>

            <template v-else>
              <tr v-for="row in rows" :key="row.id" class="hover:bg-slate-50">
                <td class="border-b border-slate-100 px-3 py-2 text-sm font-semibold text-slate-900">{{ row.name || "-" }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatQuantity(row.quantity) }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatMoney(averageCost(row)) }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatMoney(row.marketValue) }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">
                  <UnrealizedPnLBadge :value="row.realizedPnL" mode="currency" />
                </td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">
                  <UnrealizedPnLBadge :value="row.unrealizedPnLPct" />
                </td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatDate(row.lastSnapshotAt) }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">
                  <StatusBadge :tone="row.isClosed ? 'neutral' : 'success'">{{ row.isClosed ? "Closed" : "Open" }}</StatusBadge>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>

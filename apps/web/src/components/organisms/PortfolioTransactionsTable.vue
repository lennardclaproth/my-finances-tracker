<script setup lang="ts">
import { computed } from "vue";
import type { PortfolioTransaction } from "../../types/portfolio";
import BaseButton from "../atoms/BaseButton.vue";
import StatusBadge from "../atoms/StatusBadge.vue";

interface Props {
  rows: PortfolioTransaction[];
  loading?: boolean;
  errorMessage?: string;
  framed?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  errorMessage: "",
  framed: true,
});

const emit = defineEmits<{
  retry: [];
}>();

const skeletonRows = computed(() => Array.from({ length: 8 }, (_, index) => index));

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 2,
});

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return date.toLocaleDateString("en-US");
}

function formatMoney(value: string): string {
  const parsed = Number.parseFloat(value);
  if (Number.isNaN(parsed)) {
    return "-";
  }
  return currencyFormatter.format(parsed);
}

function formatDecimal(value: string): string {
  const parsed = Number.parseFloat(value);
  if (Number.isNaN(parsed)) {
    return "-";
  }
  return parsed.toLocaleString("en-US", {
    maximumFractionDigits: 6,
  });
}

function listingLabel(row: PortfolioTransaction): string {
  if (row.symbol && row.symbol.trim() !== "") {
    return row.symbol;
  }
  if (row.isin && row.isin.trim() !== "") {
    return row.isin;
  }
  return "-";
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
    <div class="h-full overflow-auto bg-slate-100">
      <table class="w-full min-w-[1100px] border-separate border-spacing-0 bg-white">
        <thead class="sticky top-0 z-20 bg-white/95 backdrop-blur">
          <tr>
            <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Occurred At</th>
            <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Type</th>
            <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Origin</th>
            <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Source</th>
            <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Listing</th>
            <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Amount</th>
            <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Quantity</th>
            <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Unit Price</th>
            <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Description</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="loading" v-for="index in skeletonRows" :key="`skeleton-${index}`">
            <td v-for="cell in 9" :key="`skeleton-${index}-${cell}`" class="border-b border-slate-100 px-3 py-3">
              <div class="h-4 w-full animate-pulse rounded bg-slate-200" />
            </td>
          </tr>

          <tr v-else-if="errorMessage">
            <td colspan="9" class="px-3 py-10 text-center">
              <p class="mb-3 text-sm text-rose-700">{{ errorMessage }}</p>
              <BaseButton size="sm" variant="secondary" @click="emit('retry')">Retry</BaseButton>
            </td>
          </tr>

          <tr v-else-if="rows.length === 0">
            <td colspan="9" class="px-3 py-10 text-center text-sm text-slate-500">
              No portfolio transactions found for this account.
            </td>
          </tr>

          <tr v-for="row in rows" v-else :key="row.id" class="hover:bg-slate-50">
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatDate(row.occurredAt) }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ row.type }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">
              <StatusBadge :tone="row.origin === 'MANUAL' ? 'info' : 'neutral'">{{ row.origin }}</StatusBadge>
            </td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ row.source }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ listingLabel(row) }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm font-medium text-slate-900">{{ formatMoney(row.amount) }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatDecimal(row.quantity) }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatMoney(row.unitPrice) }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ row.description || "-" }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

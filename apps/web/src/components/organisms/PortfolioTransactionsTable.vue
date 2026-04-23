<script setup lang="ts">
import { ArrowDownIcon, ArrowUpIcon } from "@heroicons/vue/24/solid";
import { computed } from "vue";
import type {
  PortfolioTransaction,
  PortfolioTransactionOriginFilter,
  PortfolioTransactionSortOrder,
  PortfolioTransactionTypeFilter,
} from "../../types/portfolio";
import BaseButton from "../atoms/BaseButton.vue";
import StatusBadge from "../atoms/StatusBadge.vue";
import HeaderFilterPopover from "../molecules/HeaderFilterPopover.vue";
import SelectFilterPopover from "../molecules/SelectFilterPopover.vue";

interface TransactionFilters {
  type: PortfolioTransactionTypeFilter;
  origin: PortfolioTransactionOriginFilter;
  source: string;
  listing: string;
}

interface Props {
  rows: PortfolioTransaction[];
  sortOrder: PortfolioTransactionSortOrder;
  filters: TransactionFilters;
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
  "sort-date": [];
  "update-filter": [field: keyof TransactionFilters, value: string];
}>();

const skeletonRows = computed(() => Array.from({ length: 8 }, (_, index) => index));

const typeOptions = [
  { label: "All", value: "" },
  { label: "BUY", value: "BUY" },
  { label: "SELL", value: "SELL" },
  { label: "DIVIDEND", value: "DIVIDEND" },
  { label: "TAX", value: "TAX" },
  { label: "FEE", value: "FEE" },
  { label: "CASH", value: "CASH" },
];

const originOptions = [
  { label: "All", value: "" },
  { label: "IMPORT", value: "IMPORT" },
  { label: "MANUAL", value: "MANUAL" },
];

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
    <div class="relative h-full">
      <div class="h-full overflow-auto bg-slate-100 pb-20">
        <table class="w-full min-w-[1160px] border-separate border-spacing-0 bg-white">
          <thead class="sticky top-0 z-20 bg-white/95 backdrop-blur">
            <tr>
              <th class="border-b border-slate-200 px-3 py-2 text-left">
                <button
                  type="button"
                  class="inline-flex items-center gap-1 text-left text-xs font-semibold uppercase tracking-wide text-slate-600 hover:text-slate-900"
                  @click="emit('sort-date')"
                >
                  <span>Occurred At</span>
                  <ArrowUpIcon v-if="sortOrder === 'asc'" class="h-3.5 w-3.5" />
                  <ArrowDownIcon v-else class="h-3.5 w-3.5" />
                </button>
              </th>
              <th class="border-b border-slate-200 px-3 py-2 text-left">
                <div class="flex items-center justify-between gap-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Type</span>
                  <SelectFilterPopover
                    label="Type"
                    :model-value="filters.type"
                    :options="typeOptions"
                    :loading="loading"
                    @update:model-value="emit('update-filter', 'type', $event)"
                  />
                </div>
              </th>
              <th class="border-b border-slate-200 px-3 py-2 text-left">
                <div class="flex items-center justify-between gap-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Origin</span>
                  <SelectFilterPopover
                    label="Origin"
                    :model-value="filters.origin"
                    :options="originOptions"
                    :loading="loading"
                    @update:model-value="emit('update-filter', 'origin', $event)"
                  />
                </div>
              </th>
              <th class="border-b border-slate-200 px-3 py-2 text-left">
                <div class="flex items-center justify-between gap-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Source</span>
                  <HeaderFilterPopover
                    label="Source"
                    :model-value="filters.source"
                    placeholder="Contains text"
                    :loading="loading"
                    @update:model-value="emit('update-filter', 'source', $event)"
                  />
                </div>
              </th>
              <th class="border-b border-slate-200 px-3 py-2 text-left">
                <div class="flex items-center justify-between gap-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Listing</span>
                  <HeaderFilterPopover
                    label="Listing"
                    :model-value="filters.listing"
                    placeholder="Contains symbol/isin"
                    :loading="loading"
                    @update:model-value="emit('update-filter', 'listing', $event)"
                  />
                </div>
              </th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Amount</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Quantity</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Unit Price</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Description</th>
            </tr>
          </thead>

          <tbody>
            <template v-if="loading">
              <tr v-for="index in skeletonRows" :key="`skeleton-${index}`">
                <td v-for="cell in 9" :key="`skeleton-${index}-${cell}`" class="border-b border-slate-100 px-3 py-3">
                  <div class="h-4 w-full animate-pulse rounded bg-slate-200" />
                </td>
              </tr>
            </template>

            <tr v-else-if="errorMessage">
              <td colspan="9" class="px-3 py-10 text-center">
                <p class="mb-3 text-sm text-rose-700">{{ errorMessage }}</p>
                <BaseButton size="sm" variant="secondary" @click="emit('retry')">Retry</BaseButton>
              </td>
            </tr>

            <tr v-else-if="rows.length === 0">
              <td colspan="9" class="px-3 py-10 text-center text-sm text-slate-500">
                No portfolio transactions found for the current filters.
              </td>
            </tr>

            <template v-else>
              <tr v-for="row in rows" :key="row.id" class="hover:bg-slate-50">
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
            </template>
          </tbody>
        </table>
      </div>
      <div class="absolute bottom-0 left-0 right-0 z-30 border-t border-slate-200 bg-white/95 backdrop-blur">
        <slot name="footer" />
      </div>
    </div>
  </section>
</template>

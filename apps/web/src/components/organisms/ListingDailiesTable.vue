<script setup lang="ts">
import {
  BackwardIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ForwardIcon,
} from "@heroicons/vue/24/outline";
import { computed } from "vue";
import BaseButton from "../atoms/BaseButton.vue";
import BaseSelect from "../atoms/BaseSelect.vue";
import IconButton from "../atoms/IconButton.vue";
import type { ListingDaily } from "../../types/marketdata";

interface Props {
  rows: ListingDaily[];
  limit: number;
  offset: number;
  total: number;
  loading?: boolean;
  errorMessage?: string;
  emptyMessage?: string;
  framed?: boolean;
  pageSizeOptions?: number[];
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  errorMessage: "",
  emptyMessage: "No daily data found for this listing and date range.",
  framed: true,
  pageSizeOptions: () => [10, 25, 50, 100],
});

const emit = defineEmits<{
  retry: [];
  "change-limit": [value: number];
  "go-first": [];
  "go-prev": [];
  "go-next": [];
  "go-last": [];
}>();

const skeletonRows = computed(() => Array.from({ length: 10 }, (_, index) => index));
const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.limit)));
const currentPage = computed(() => Math.floor(props.offset / props.limit) + 1);
const hasPrevious = computed(() => props.offset > 0);
const hasNext = computed(() => props.offset + props.limit < props.total);
const rangeStart = computed(() => (props.total === 0 ? 0 : props.offset + 1));
const rangeEnd = computed(() => (props.total === 0 ? 0 : props.offset + props.rows.length));

const pageSizeSelectOptions = computed(() =>
  props.pageSizeOptions.map((option) => ({
    label: `${option} / page`,
    value: option,
  })),
);

const rootClasses = computed(() => {
  if (!props.framed) {
    return "h-full";
  }
  return "h-full overflow-hidden rounded-3xl border border-slate-300 bg-white/95 p-4 shadow-sm";
});

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return date.toLocaleDateString("en-US");
}

function formatPrice(value: number): string {
  if (!Number.isFinite(value)) {
    return "-";
  }
  return value.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 6,
  });
}

function onLimitChange(rawValue: string): void {
  const nextLimit = Number.parseInt(rawValue, 10);
  if (!Number.isNaN(nextLimit)) {
    emit("change-limit", nextLimit);
  }
}
</script>

<template>
  <section :class="rootClasses">
    <div class="relative flex h-full min-h-0 flex-col">
      <div class="min-h-0 flex-1 overflow-auto bg-slate-100">
        <table class="w-full min-w-[960px] border-separate border-spacing-0 bg-white">
          <thead class="sticky top-0 z-20 bg-white/95 backdrop-blur">
            <tr>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Date</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Open</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">High</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Low</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Close</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Volume</th>
            </tr>
          </thead>
          <tbody>
            <template v-if="loading">
              <tr v-for="index in skeletonRows" :key="`skeleton-${index}`">
                <td v-for="cell in 6" :key="`skeleton-${index}-${cell}`" class="border-b border-slate-100 px-3 py-3">
                  <div class="h-4 w-full animate-pulse rounded bg-slate-200" />
                </td>
              </tr>
            </template>

            <tr v-else-if="errorMessage">
              <td colspan="6" class="px-3 py-10 text-center">
                <p class="mb-3 text-sm text-rose-700">{{ errorMessage }}</p>
                <BaseButton size="sm" variant="secondary" @click="emit('retry')">Retry</BaseButton>
              </td>
            </tr>

            <tr v-else-if="rows.length === 0">
              <td colspan="6" class="px-3 py-10 text-center text-sm text-slate-500">
                {{ emptyMessage }}
              </td>
            </tr>

            <template v-else>
              <tr v-for="row in rows" :key="row.id" class="hover:bg-slate-50">
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatDate(row.date) }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatPrice(row.open) }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatPrice(row.high) }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatPrice(row.low) }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm font-medium text-slate-900">{{ formatPrice(row.close) }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ row.volume.toLocaleString("en-US") }}</td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <footer class="mt-2">
        <div class="grid grid-cols-1 gap-3 lg:grid-cols-[1fr_auto] lg:items-center">
          <div class="flex flex-wrap items-center gap-3 text-sm text-slate-600">
            <BaseSelect
              :model-value="limit"
              :options="pageSizeSelectOptions"
              :disabled="loading"
              @update:model-value="onLimitChange"
            />
            <span>Showing {{ rangeStart }}-{{ rangeEnd }} of {{ total }}</span>
          </div>

          <div class="flex items-center justify-start gap-2 lg:justify-end">
            <span class="text-sm text-slate-600">Page {{ currentPage }} / {{ totalPages }}</span>
            <IconButton title="First page" :disabled="loading || !hasPrevious" @click="emit('go-first')">
              <BackwardIcon class="h-4 w-4" />
            </IconButton>
            <IconButton title="Previous page" :disabled="loading || !hasPrevious" @click="emit('go-prev')">
              <ChevronLeftIcon class="h-4 w-4" />
            </IconButton>
            <IconButton title="Next page" :disabled="loading || !hasNext" @click="emit('go-next')">
              <ChevronRightIcon class="h-4 w-4" />
            </IconButton>
            <IconButton title="Last page" :disabled="loading || !hasNext" @click="emit('go-last')">
              <ForwardIcon class="h-4 w-4" />
            </IconButton>
          </div>
        </div>
      </footer>
    </div>
  </section>
</template>

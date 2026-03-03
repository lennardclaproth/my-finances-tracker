<script setup lang="ts">
import { computed } from "vue";
import type { CashflowTransaction, DirectionFilter, SortBy, SortOrder } from "../../types/cashflow";
import { formatAmountCents, formatLocalDate, normalizeDirection } from "../../utils/formatters";
import BaseCheckbox from "../atoms/BaseCheckbox.vue";
import StatusBadge from "../atoms/StatusBadge.vue";
import VisibilityIndicator from "../atoms/VisibilityIndicator.vue";
import DirectionFilterPopover from "../molecules/DirectionFilterPopover.vue";
import HeaderFilterPopover from "../molecules/HeaderFilterPopover.vue";
import SortableHeader from "../molecules/SortableHeader.vue";
import VisibilityFilterPopover from "../molecules/VisibilityFilterPopover.vue";

export interface ColumnFilters {
  description: string;
  note: string;
  source: string;
  direction: DirectionFilter;
  tags: string;
  untagged: boolean;
}

interface Props {
  rows: CashflowTransaction[];
  selectedIds: string[];
  allMatchingSelected: boolean;
  sortBy: SortBy;
  sortOrder: SortOrder;
  columnFilters: ColumnFilters;
  showHidden: boolean;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  sort: [field: SortBy];
  "toggle-row": [id: string, checked: boolean];
  "toggle-visible": [checked: boolean];
  "update-filter": [field: Exclude<keyof ColumnFilters, "untagged">, value: string];
  "update-untagged-filter": [value: boolean];
  "update-show-hidden": [value: boolean];
}>();

const selectedIdSet = computed(() => new Set(props.selectedIds));
const selectedVisibleCount = computed(() =>
  props.allMatchingSelected
    ? props.rows.length
    : props.rows.filter((row) => selectedIdSet.value.has(row.id)).length,
);
const allVisibleSelected = computed(
  () => props.rows.length > 0 && selectedVisibleCount.value === props.rows.length,
);
const someVisibleSelected = computed(
  () => selectedVisibleCount.value > 0 && !allVisibleSelected.value,
);
const skeletonRows = computed(() => Array.from({ length: 8 }, (_, index) => index));

function directionTone(value: string): "success" | "warning" | "neutral" {
  const normalized = value.trim().toLowerCase();
  if (normalized === "in") {
    return "success";
  }
  if (normalized === "out") {
    return "warning";
  }
  return "neutral";
}

</script>

<template>
  <section class="h-full overflow-hidden rounded-3xl border border-slate-300 bg-white/95 p-4 shadow-sm">
    <div class="relative h-full">
      <div class="h-full overflow-auto bg-slate-100 pb-20">
        <table class="w-full min-w-[1240px] border-separate border-spacing-0 bg-white">
        <thead class="sticky top-0 z-20 bg-white/95 backdrop-blur">
          <tr>
            <th class="w-12 border-b border-slate-200 px-3 py-2 text-left">
              <BaseCheckbox
                :checked="allVisibleSelected"
                :indeterminate="someVisibleSelected"
                :disabled="rows.length === 0 || loading"
                @update:checked="emit('toggle-visible', $event)"
              />
            </th>
            <th class="border-b border-slate-200 px-3 py-2 text-left">
              <div class="flex items-center justify-between gap-1">
                <SortableHeader
                  label="Description"
                  field="description"
                  :active-sort-by="sortBy"
                  :active-sort-order="sortOrder"
                  @sort="emit('sort', $event)"
                />
                <HeaderFilterPopover
                  label="Description"
                  :model-value="columnFilters.description"
                  placeholder="Contains text"
                  @update:model-value="emit('update-filter', 'description', $event)"
                />
              </div>
            </th>
            <th class="border-b border-slate-200 px-3 py-2 text-left">
              <div class="flex items-center justify-between gap-1">
                <SortableHeader
                  label="Note"
                  field="note"
                  :active-sort-by="sortBy"
                  :active-sort-order="sortOrder"
                  @sort="emit('sort', $event)"
                />
                <HeaderFilterPopover
                  label="Note"
                  :model-value="columnFilters.note"
                  placeholder="Contains text"
                  @update:model-value="emit('update-filter', 'note', $event)"
                />
              </div>
            </th>
            <th class="border-b border-slate-200 px-3 py-2 text-left">
              <div class="flex items-center justify-between gap-1">
                <SortableHeader
                  label="Source"
                  field="source"
                  :active-sort-by="sortBy"
                  :active-sort-order="sortOrder"
                  @sort="emit('sort', $event)"
                />
                <HeaderFilterPopover
                  label="Source"
                  :model-value="columnFilters.source"
                  placeholder="Contains text"
                  @update:model-value="emit('update-filter', 'source', $event)"
                />
              </div>
            </th>
            <th class="border-b border-slate-200 px-3 py-2 text-left">
              <SortableHeader
                label="Amount"
                field="amount"
                :active-sort-by="sortBy"
                :active-sort-order="sortOrder"
                @sort="emit('sort', $event)"
              />
            </th>
            <th class="border-b border-slate-200 px-3 py-2 text-left">
              <div class="flex items-center justify-between gap-1">
                <SortableHeader
                  label="Direction"
                  field="date"
                  :active-sort-by="sortBy"
                  :active-sort-order="sortOrder"
                  :sortable="false"
                />
                <DirectionFilterPopover
                  :model-value="columnFilters.direction"
                  :loading="loading"
                  @update:model-value="emit('update-filter', 'direction', $event)"
                />
              </div>
            </th>
            <th class="border-b border-slate-200 px-3 py-2 text-left">
              <SortableHeader
                label="Date"
                field="date"
                :active-sort-by="sortBy"
                :active-sort-order="sortOrder"
                @sort="emit('sort', $event)"
              />
            </th>
            <th class="border-b border-slate-200 px-3 py-2 text-left">
              <div class="flex items-center justify-between gap-1">
                <SortableHeader
                  label="Tag"
                  field="tag"
                  :active-sort-by="sortBy"
                  :active-sort-order="sortOrder"
                  @sort="emit('sort', $event)"
                />
                <HeaderFilterPopover
                  label="Tag"
                  :model-value="columnFilters.tags"
                  placeholder="Comma separated"
                  :supports-untagged="true"
                  :untagged-only="columnFilters.untagged"
                  @update:model-value="emit('update-filter', 'tags', $event)"
                  @update:untagged-only="emit('update-untagged-filter', $event)"
                />
              </div>
            </th>
            <th class="border-b border-slate-200 px-3 py-2 text-left">
              <div class="flex items-center justify-between gap-1">
                <SortableHeader
                  label="Visibility"
                  field="date"
                  :active-sort-by="sortBy"
                  :active-sort-order="sortOrder"
                  :sortable="false"
                />
                <VisibilityFilterPopover
                  :show-hidden="showHidden"
                  :loading="loading"
                  @update:show-hidden="emit('update-show-hidden', $event)"
                />
              </div>
            </th>
          </tr>
        </thead>

        <tbody>
          <tr v-if="loading" v-for="index in skeletonRows" :key="`skeleton-${index}`" class="border-b border-slate-100">
            <td class="border-b border-slate-100 px-3 py-3">
              <div class="h-4 w-4 animate-pulse rounded bg-slate-200" />
            </td>
            <td v-for="cell in 8" :key="`skeleton-${index}-cell-${cell}`" class="border-b border-slate-100 px-3 py-3">
              <div class="h-4 w-full animate-pulse rounded bg-slate-200" />
            </td>
          </tr>

          <tr v-else-if="rows.length === 0">
            <td colspan="9" class="px-3 py-12 text-center text-sm text-slate-500">
              No transactions found for the current filters.
            </td>
          </tr>

          <tr v-for="row in rows" :key="row.id" class="border-b border-slate-100 hover:bg-slate-50">
            <td class="border-b border-slate-100 px-3 py-2">
              <BaseCheckbox
                :checked="allMatchingSelected || selectedIdSet.has(row.id)"
                :disabled="loading"
                @update:checked="emit('toggle-row', row.id, $event)"
              />
            </td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-800">{{ row.description }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ row.note || "-" }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ row.source }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm font-medium text-slate-900">
              {{ formatAmountCents(row.amountCents, row.direction) }}
            </td>
            <td class="border-b border-slate-100 px-3 py-2">
              <StatusBadge :tone="directionTone(row.direction)">{{ normalizeDirection(row.direction) }}</StatusBadge>
            </td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatLocalDate(row.date) }}</td>
            <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ row.tag || "-" }}</td>
            <td class="border-b border-slate-100 px-3 py-2">
              <VisibilityIndicator :visible="!row.ignored" />
            </td>
          </tr>
        </tbody>
        </table>
      </div>
      <div class="absolute bottom-0 left-0 right-0 z-30 border-t border-slate-200 bg-white/95 backdrop-blur">
        <slot name="footer" />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ArrowDownIcon, ArrowUpIcon, ArrowsUpDownIcon } from "@heroicons/vue/24/solid";
import type { SortBy, SortOrder } from "../../types/cashflow";

interface Props {
  label: string;
  field: SortBy;
  activeSortBy: SortBy;
  activeSortOrder: SortOrder;
  sortable?: boolean;
}

withDefaults(defineProps<Props>(), {
  sortable: true,
});

const emit = defineEmits<{ sort: [field: SortBy] }>();
</script>

<template>
  <span
    v-if="!sortable"
    class="inline-flex items-center text-left text-xs font-semibold uppercase tracking-wide text-slate-600"
  >
    {{ label }}
  </span>
  <button
    v-else
    type="button"
    class="inline-flex items-center gap-1 text-left text-xs font-semibold uppercase tracking-wide text-slate-600 hover:text-slate-900"
    @click="emit('sort', field)"
  >
    <span>{{ label }}</span>
    <ArrowUpIcon v-if="activeSortBy === field && activeSortOrder === 'asc'" class="h-3.5 w-3.5" />
    <ArrowDownIcon v-else-if="activeSortBy === field && activeSortOrder === 'desc'" class="h-3.5 w-3.5" />
    <ArrowsUpDownIcon v-else class="h-3.5 w-3.5 text-slate-300" />
  </button>
</template>

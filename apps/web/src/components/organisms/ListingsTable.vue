<script setup lang="ts">
import { computed } from "vue";
import BaseButton from "../atoms/BaseButton.vue";
import type { Listing } from "../../types/listings";

interface Props {
  rows: Listing[];
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

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return date.toLocaleString();
}
</script>

<template>
  <section :class="props.framed ? 'h-full overflow-hidden rounded-3xl border border-slate-300 bg-white/95 p-4 shadow-sm' : 'h-full'">
    <div class="relative h-full">
      <div class="h-full overflow-auto bg-slate-100">
        <table class="w-full min-w-[1120px] border-separate border-spacing-0 bg-white">
          <thead class="sticky top-0 z-20 bg-white/95 backdrop-blur">
            <tr>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Symbol</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Name</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Source</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">ISIN</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Exchange</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Region</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Currency</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Type</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Updated At</th>
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
                No listings found. Add your first listing.
              </td>
            </tr>

            <template v-else>
              <tr v-for="listing in rows" :key="listing.id" class="hover:bg-slate-50">
                <td class="border-b border-slate-100 px-3 py-2 text-sm font-semibold text-slate-900">{{ listing.symbol }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ listing.name }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ listing.source }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ listing.isin || "-" }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ listing.exchange || "-" }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ listing.region || "-" }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ listing.currency || "-" }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ listing.type || "-" }}</td>
                <td class="border-b border-slate-100 px-3 py-2 text-sm text-slate-700">{{ formatDate(listing.updatedAt) }}</td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>

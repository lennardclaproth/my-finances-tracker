<script setup lang="ts">
import { Cog6ToothIcon } from "@heroicons/vue/24/outline";
import { computed } from "vue";
import type { AssetClass } from "../../types/assets";
import BaseButton from "../atoms/BaseButton.vue";
import IconButton from "../atoms/IconButton.vue";
import UnrealizedPnLBadge from "../molecules/UnrealizedPnLBadge.vue";

interface Props {
  rows: AssetClass[];
  loading?: boolean;
  selectedClassId?: string | null;
  errorMessage?: string;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  selectedClassId: null,
  errorMessage: "",
});

const emit = defineEmits<{
  select: [row: AssetClass];
  edit: [row: AssetClass];
  retry: [];
}>();

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

const hasRows = computed(() => props.rows.length > 0);

function formatWorth(value: string): string {
  const amount = Number.parseFloat(value);
  if (Number.isNaN(amount)) {
    return value;
  }
  return currencyFormatter.format(amount);
}

function formatDate(value?: string): string {
  if (!value) {
    return "n/a";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString();
}

</script>

<template>
  <section class="h-full overflow-hidden rounded-3xl border border-slate-300 bg-white/95 p-4 shadow-sm">
    <div class="flex h-full min-h-0 flex-col">
      <header class="mb-3 flex items-center justify-between gap-2">
        <h2 class="font-secondary text-xl font-semibold text-slate-700 md:text-2xl">Asset classes</h2>
        <BaseButton
          v-if="errorMessage"
          variant="secondary"
          size="sm"
          @click="emit('retry')"
        >
          Retry
        </BaseButton>
      </header>

      <p v-if="errorMessage" class="mb-3 rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
        {{ errorMessage }}
      </p>

      <div class="min-h-0 flex-1 overflow-auto bg-slate-100 pb-20">
        <table class="w-full min-w-[1060px] border-separate border-spacing-0 bg-white text-sm">
          <thead class="sticky top-0 z-20 bg-white/95 backdrop-blur">
            <tr>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Name</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Source</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Current worth</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Last change</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Updated</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Growth</th>
              <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Status</th>
              <th class="border-b border-slate-200 px-3 py-2 text-right text-xs font-semibold uppercase tracking-wide text-slate-500">Edit</th>
            </tr>
          </thead>
          <tbody v-if="loading">
            <tr v-for="index in 6" :key="`skeleton-${index}`">
              <td v-for="column in 8" :key="column" class="border-b border-slate-100 px-3 py-3">
                <div class="h-4 w-full animate-pulse rounded bg-slate-200" />
              </td>
            </tr>
          </tbody>
          <tbody v-else-if="hasRows">
            <tr
              v-for="row in rows"
              :key="row.id"
              class="cursor-pointer transition hover:bg-slate-50"
              :class="selectedClassId === row.id ? 'bg-indigo-50/60' : ''"
              @click="emit('select', row)"
            >
              <td class="border-b border-slate-100 px-3 py-2 font-medium text-slate-800">{{ row.name }}</td>
              <td class="border-b border-slate-100 px-3 py-2 text-slate-700">{{ row.source === "PORTFOLIO" ? "Portfolio" : "Manual" }}</td>
              <td class="border-b border-slate-100 px-3 py-2 text-slate-800">{{ formatWorth(row.currentWorth) }}</td>
              <td class="border-b border-slate-100 px-3 py-2 text-slate-600">{{ formatDate(row.lastChangeAt) }}</td>
              <td class="border-b border-slate-100 px-3 py-2 text-slate-600">{{ formatDate(row.updatedAt) }}</td>
              <td class="border-b border-slate-100 px-3 py-2">
                <UnrealizedPnLBadge :value="row.growthPct" />
              </td>
              <td class="border-b border-slate-100 px-3 py-2">
                <span
                  class="inline-flex rounded-full border px-2 py-0.5 text-xs font-medium"
                  :class="
                    row.archived
                      ? 'border-amber-200 bg-amber-50 text-amber-700'
                      : 'border-emerald-200 bg-emerald-50 text-emerald-700'
                  "
                >
                  {{ row.archived ? "Archived" : "Active" }}
                </span>
              </td>
              <td class="border-b border-slate-100 px-3 py-2 text-right">
                <IconButton
                  v-if="row.source === 'MANUAL'"
                  tone="neutral"
                  size="sm"
                  title="Edit asset class"
                  @click.stop="emit('edit', row)"
                >
                  <Cog6ToothIcon class="h-4 w-4" />
                </IconButton>
              </td>
            </tr>
          </tbody>
          <tbody v-else>
            <tr>
              <td colspan="8" class="px-4 py-10 text-center text-sm text-slate-500">
                No asset classes yet. Use the add button to create one.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </section>
</template>


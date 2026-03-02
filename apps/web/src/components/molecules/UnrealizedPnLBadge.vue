<script setup lang="ts">
import { ArrowTrendingDownIcon, ArrowTrendingUpIcon } from "@heroicons/vue/24/outline";
import { computed } from "vue";

interface Props {
  value?: number;
  mode?: "percent" | "currency";
}

const props = withDefaults(defineProps<Props>(), {
  mode: "percent",
});
const moneyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 2,
});

const toneClass = computed(() => {
  if (props.value === undefined || props.value === null) {
    return "border-slate-200 bg-slate-50 text-slate-500";
  }
  if (props.value > 0) {
    return "border-emerald-200 bg-emerald-50 text-emerald-700";
  }
  if (props.value < 0) {
    return "border-rose-200 bg-rose-50 text-rose-700";
  }
  return "border-slate-200 bg-slate-50 text-slate-600";
});

const formatted = computed(() => {
  if (props.value === undefined || props.value === null) {
    return "-";
  }
  if (props.mode === "currency") {
    return moneyFormatter.format(props.value / 1_000_000);
  }
  return `${props.value.toFixed(2)}%`;
});
</script>

<template>
  <span
    class="inline-flex items-center gap-1 rounded-full border px-2.5 py-1 text-xs font-semibold"
    :class="toneClass"
  >
    <ArrowTrendingUpIcon
      v-if="value !== undefined && value !== null && value > 0"
      class="h-3.5 w-3.5"
    />
    <ArrowTrendingDownIcon
      v-else-if="value !== undefined && value !== null && value < 0"
      class="h-3.5 w-3.5"
    />
    <span>{{ formatted }}</span>
  </span>
</template>

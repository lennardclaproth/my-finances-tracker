<script setup lang="ts">
import { computed } from "vue";

type PortfolioTab = "positions" | "transactions";

interface TabOption {
  label: string;
  value: PortfolioTab;
}

interface Props {
  modelValue: PortfolioTab;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  "update:modelValue": [value: PortfolioTab];
}>();

const tabs: TabOption[] = [
  { label: "Positions", value: "positions" },
  { label: "Transactions", value: "transactions" },
];

const activeValue = computed(() => props.modelValue);
</script>

<template>
  <div class="inline-flex rounded-full border border-slate-200 bg-slate-100 p-1">
    <button
      v-for="tab in tabs"
      :key="tab.value"
      type="button"
      class="rounded-full px-4 py-1.5 text-sm font-medium transition"
      :class="activeValue === tab.value ? 'bg-white text-slate-900 shadow-sm' : 'text-slate-600 hover:text-slate-800'"
      @click="emit('update:modelValue', tab.value)"
    >
      {{ tab.label }}
    </button>
  </div>
</template>

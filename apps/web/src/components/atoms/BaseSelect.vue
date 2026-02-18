<script setup lang="ts">
import { ChevronDownIcon } from "@heroicons/vue/24/solid";
import { computed } from "vue";

interface Option {
  label: string;
  value: string | number;
}

interface Props {
  modelValue: string | number;
  options: Option[];
  disabled?: boolean;
  rounded?: "default" | "pill";
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  rounded: "pill",
});

const emit = defineEmits<{ "update:modelValue": [value: string] }>();

const classes = computed(() => {
  const shape = props.rounded === "pill" ? "rounded-full" : "rounded-md";
  return `w-full appearance-none ${shape} border border-slate-300 bg-white px-3 py-2 pr-9 text-sm font-normal text-slate-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 disabled:cursor-not-allowed disabled:bg-slate-100`;
});
</script>

<template>
  <div class="relative">
    <select
      :value="modelValue"
      :disabled="disabled"
      :class="classes"
      @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
    >
      <option v-for="option in options" :key="String(option.value)" :value="option.value">
        {{ option.label }}
      </option>
    </select>
    <ChevronDownIcon class="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
  </div>
</template>

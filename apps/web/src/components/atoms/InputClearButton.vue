<script setup lang="ts">
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { computed } from "vue";

type Size = "sm" | "md";

interface Props {
  disabled?: boolean;
  title?: string;
  size?: Size;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  title: "Clear",
  size: "md",
});

const classes = computed(() => {
  const base =
    "inline-flex items-center justify-center rounded-full text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 disabled:cursor-not-allowed disabled:opacity-60";
  const sizeMap: Record<Size, string> = {
    sm: "h-5 w-5",
    md: "h-6 w-6",
  };
  return `${base} ${sizeMap[props.size]}`;
});

const iconClass = computed(() => (props.size === "sm" ? "h-3.5 w-3.5" : "h-4 w-4"));
</script>

<template>
  <button type="button" :disabled="disabled" :title="title" :aria-label="title" :class="classes">
    <XMarkIcon :class="iconClass" />
  </button>
</template>

<script setup lang="ts">
import { computed } from "vue";

type Tone = "neutral" | "success" | "warning" | "info";

interface Props {
  tone?: Tone;
}

const props = withDefaults(defineProps<Props>(), {
  tone: "neutral",
});

const classes = computed(() => {
  const mapping: Record<Tone, string> = {
    neutral: "bg-slate-100 text-slate-700",
    success: "bg-emerald-100 text-emerald-800",
    warning: "bg-amber-100 text-amber-800",
    info: "bg-cyan-100 text-cyan-800",
  };

  return `inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold ${mapping[props.tone]}`;
});
</script>

<template>
  <span :class="classes">
    <slot />
  </span>
</template>
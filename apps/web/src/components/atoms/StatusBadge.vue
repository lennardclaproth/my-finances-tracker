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
    neutral: "border-slate-200 bg-slate-50 text-slate-700",
    success: "border-emerald-200 bg-emerald-50 text-emerald-700",
    warning: "border-amber-200 bg-amber-50 text-amber-700",
    info: "border-cyan-200 bg-cyan-50 text-cyan-700",
  };

  return `inline-flex items-center rounded-full border px-2.5 py-1 text-xs font-semibold ${mapping[props.tone]}`;
});
</script>

<template>
  <span :class="classes">
    <slot />
  </span>
</template>

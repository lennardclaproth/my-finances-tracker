<script setup lang="ts">
import { computed } from "vue";

type Tone = "neutral" | "positive" | "negative";

interface Props {
  label: string;
  value: string;
  tone?: Tone;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  tone: "neutral",
  loading: false,
});

const valueToneClass = computed(() => {
  if (props.tone === "positive") {
    return "text-emerald-700";
  }
  if (props.tone === "negative") {
    return "text-rose-700";
  }
  return "text-slate-900";
});
</script>

<template>
  <article class="rounded-2xl border border-slate-200 bg-slate-50/70 p-3">
    <template v-if="loading">
      <div class="h-3 w-24 animate-pulse rounded bg-slate-200" />
      <div class="mt-2 h-7 w-28 animate-pulse rounded bg-slate-200" />
    </template>
    <template v-else>
      <p class="text-[11px] font-semibold uppercase tracking-wide text-slate-500">{{ label }}</p>
      <p class="mt-1 text-xl font-semibold" :class="valueToneClass">{{ value }}</p>
    </template>
  </article>
</template>

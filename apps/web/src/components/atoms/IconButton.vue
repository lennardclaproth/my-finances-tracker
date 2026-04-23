<script setup lang="ts">
import { computed } from "vue";

type Tone = "neutral" | "primary" | "warning" | "info";
type Size = "sm" | "md" | "fab";

interface Props {
  tone?: Tone;
  size?: Size;
  disabled?: boolean;
  title: string;
  type?: "button" | "submit" | "reset";
  unstyled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  tone: "neutral",
  size: "md",
  disabled: false,
  type: "button",
  unstyled: false,
});

const classes = computed(() => {
  if (props.unstyled) {
    return "";
  }

  const base =
    "inline-flex items-center justify-center rounded-full border transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50";
  const sizeClass =
    props.size === "fab" ? "h-14 w-14 shadow-md" : props.size === "sm" ? "h-8 w-8" : "h-9 w-9";

  const toneMap: Record<Tone, string> = {
    neutral: "border-slate-300 bg-white text-slate-600 hover:bg-slate-100 focus-visible:ring-slate-300",
    primary: "border-blue-300 bg-blue-100 text-blue-800 hover:bg-blue-200 focus-visible:ring-blue-300",
    warning: "border-amber-300 bg-amber-100 text-amber-800 hover:bg-amber-200 focus-visible:ring-amber-300",
    info: "border-cyan-300 bg-cyan-100 text-cyan-800 hover:bg-cyan-200 focus-visible:ring-cyan-300",
  };

  return `${base} ${sizeClass} ${toneMap[props.tone]}`;
});
</script>

<template>
  <button :type="type" :disabled="disabled" :class="classes" :title="title" :aria-label="title">
    <slot />
  </button>
</template>

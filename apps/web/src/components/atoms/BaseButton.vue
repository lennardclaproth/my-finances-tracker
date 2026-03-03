<script setup lang="ts">
import { computed } from "vue";

type Variant = "primary" | "secondary" | "success" | "warning" | "info" | "danger" | "ghost";
type Size = "sm" | "md";

interface Props {
  variant?: Variant;
  size?: Size;
  type?: "button" | "submit" | "reset";
  disabled?: boolean;
  unstyled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  variant: "primary",
  size: "md",
  type: "button",
  disabled: false,
  unstyled: false,
});

const classes = computed(() => {
  if (props.unstyled) {
    return "";
  }

  const base =
    "inline-flex items-center justify-center gap-2 rounded-md font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60";
  const size = props.size === "sm" ? "px-3 py-1.5 text-xs" : "px-4 py-2 text-sm";

  const variantMap: Record<Variant, string> = {
    primary: "bg-indigo-600 text-white hover:bg-indigo-700 focus-visible:ring-indigo-500",
    secondary: "bg-slate-200 text-slate-900 hover:bg-slate-300 focus-visible:ring-slate-400",
    success: "bg-emerald-600 text-white hover:bg-emerald-700 focus-visible:ring-emerald-500",
    warning: "bg-amber-500 text-slate-950 hover:bg-amber-600 focus-visible:ring-amber-400",
    info: "bg-cyan-600 text-white hover:bg-cyan-700 focus-visible:ring-cyan-500",
    danger: "bg-rose-600 text-white hover:bg-rose-700 focus-visible:ring-rose-500",
    ghost: "bg-transparent text-slate-700 hover:bg-slate-100 focus-visible:ring-slate-400",
  };

  return [base, size, variantMap[props.variant]].join(" ");
});
</script>

<template>
  <button :type="type" :disabled="disabled" :class="classes">
    <slot />
  </button>
</template>

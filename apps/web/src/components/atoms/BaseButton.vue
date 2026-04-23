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
    "inline-flex items-center justify-center gap-2 rounded-md border font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60";
  const size = props.size === "sm" ? "px-3 py-1.5 text-xs" : "px-4 py-2 text-sm";

  const variantMap: Record<Variant, string> = {
    primary: "border-blue-300 bg-blue-100 text-blue-800 hover:bg-blue-200 focus-visible:ring-blue-300",
    secondary: "border-slate-300 bg-slate-100 text-slate-800 hover:bg-slate-200 focus-visible:ring-slate-300",
    success: "border-emerald-300 bg-emerald-100 text-emerald-800 hover:bg-emerald-200 focus-visible:ring-emerald-300",
    warning: "border-amber-300 bg-amber-100 text-amber-800 hover:bg-amber-200 focus-visible:ring-amber-300",
    info: "border-cyan-300 bg-cyan-100 text-cyan-800 hover:bg-cyan-200 focus-visible:ring-cyan-300",
    danger: "border-rose-300 bg-rose-100 text-rose-800 hover:bg-rose-200 focus-visible:ring-rose-300",
    ghost: "border-transparent bg-transparent text-slate-700 hover:bg-slate-100 focus-visible:ring-slate-300",
  };

  return [base, size, variantMap[props.variant]].join(" ");
});
</script>

<template>
  <button :type="type" :disabled="disabled" :class="classes">
    <slot />
  </button>
</template>

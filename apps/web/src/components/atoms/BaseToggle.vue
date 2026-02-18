<script setup lang="ts">
import { computed } from "vue";

interface Props {
  checked: boolean;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
});

const emit = defineEmits<{ "update:checked": [value: boolean] }>();

const trackClass = computed(() => {
  const base =
    "inline-flex h-5 w-9 items-center rounded-full border transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-300 disabled:cursor-not-allowed disabled:opacity-60";
  const state = props.checked
    ? "border-blue-500 bg-blue-500"
    : "border-slate-300 bg-slate-200";
  return `${base} ${state}`;
});

const thumbClass = computed(() =>
  props.checked
    ? "h-4 w-4 translate-x-4 rounded-full bg-white shadow-sm transition-transform"
    : "h-4 w-4 translate-x-0.5 rounded-full bg-white shadow-sm transition-transform",
);

function toggle(): void {
  if (props.disabled) {
    return;
  }
  emit("update:checked", !props.checked);
}
</script>

<template>
  <button
    type="button"
    role="switch"
    :aria-checked="checked"
    :disabled="disabled"
    :class="trackClass"
    @click="toggle"
  >
    <span :class="thumbClass" />
  </button>
</template>

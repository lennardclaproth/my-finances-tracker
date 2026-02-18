<script setup lang="ts">
import { computed, ref } from "vue";

interface Props {
  modelValue: string;
  type?: "text" | "search" | "date";
  placeholder?: string;
  disabled?: boolean;
  rounded?: "default" | "pill";
}

const emit = defineEmits<{ "update:modelValue": [value: string] }>();
const props = withDefaults(defineProps<Props>(), {
  type: "text",
  placeholder: "",
  disabled: false,
  rounded: "default",
});

const inputRef = ref<HTMLInputElement | null>(null);

const classes = computed(() => {
  const shape = props.rounded === "pill" ? "rounded-full" : "rounded-md";
  return `w-full ${shape} border border-slate-300 bg-white px-3 py-2 text-sm font-normal text-slate-900 shadow-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-200 disabled:cursor-not-allowed disabled:bg-slate-100`;
});

function focus(): void {
  inputRef.value?.focus();
}

function select(): void {
  inputRef.value?.select();
}

defineExpose({
  focus,
  select,
});
</script>

<template>
  <input
    ref="inputRef"
    :value="modelValue"
    :type="type"
    :placeholder="placeholder"
    :disabled="disabled"
    :class="classes"
    @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
  />
</template>

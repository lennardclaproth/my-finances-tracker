<script setup lang="ts">
import { ref, watch } from "vue";

interface Props {
  checked: boolean;
  indeterminate?: boolean;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  indeterminate: false,
  disabled: false,
});

const emit = defineEmits<{ "update:checked": [value: boolean] }>();
const inputRef = ref<HTMLInputElement | null>(null);

watch(
  () => props.indeterminate,
  (value) => {
    if (inputRef.value) {
      inputRef.value.indeterminate = value;
    }
  },
  { immediate: true },
);
</script>

<template>
  <input
    ref="inputRef"
    type="checkbox"
    :checked="checked"
    :disabled="disabled"
    class="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500 disabled:cursor-not-allowed"
    @change="emit('update:checked', ($event.target as HTMLInputElement).checked)"
  />
</template>
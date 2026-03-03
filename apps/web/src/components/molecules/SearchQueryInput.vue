<script setup lang="ts">
import { MagnifyingGlassIcon } from "@heroicons/vue/24/outline";
import { onBeforeUnmount, ref, watch } from "vue";
import BaseInput from "../atoms/BaseInput.vue";
import InputClearButton from "../atoms/InputClearButton.vue";

interface Props {
  modelValue: string;
  disabled?: boolean;
  debounceMs?: number;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  debounceMs: 300,
});

const emit = defineEmits<{
  "update:modelValue": [value: string];
  "debounced-change": [value: string];
}>();

const localValue = ref(props.modelValue);
let debounceHandle: ReturnType<typeof setTimeout> | null = null;

watch(
  () => props.modelValue,
  (value) => {
    if (value !== localValue.value) {
      localValue.value = value;
    }
  },
);

watch(localValue, (value) => {
  emit("update:modelValue", value);

  if (debounceHandle) {
    clearTimeout(debounceHandle);
  }

  debounceHandle = setTimeout(() => {
    emit("debounced-change", value.trim());
  }, props.debounceMs);
});

function clearValue(): void {
  localValue.value = "";
}

onBeforeUnmount(() => {
  if (debounceHandle) {
    clearTimeout(debounceHandle);
  }
});
</script>

<template>
  <div class="relative w-full max-w-xl">
    <MagnifyingGlassIcon class="pointer-events-none absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
    <BaseInput
      v-model="localValue"
      type="text"
      rounded="pill"
      :disabled="disabled"
      placeholder="Search description, note or tag"
      class="w-full border-slate-200 bg-white/90 py-2 pl-10 pr-9 placeholder:text-slate-400 focus:border-indigo-400 focus:ring-indigo-200"
    />
    <InputClearButton
      v-if="localValue"
      class="absolute right-2.5 top-1.5"
      :disabled="disabled"
      title="Clear search"
      @click="clearValue"
    />
  </div>
</template>

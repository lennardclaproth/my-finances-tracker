<script setup lang="ts">
import { FunnelIcon } from "@heroicons/vue/24/outline";
import { nextTick, ref, watch } from "vue";
import BasePopover from "../atoms/BasePopover.vue";
import BaseInput from "../atoms/BaseInput.vue";
import IconButton from "../atoms/IconButton.vue";
import InputClearButton from "../atoms/InputClearButton.vue";
import BaseToggle from "../atoms/BaseToggle.vue";

interface Props {
  label: string;
  loading?: boolean;
  modelValue: string;
  placeholder?: string;
  supportsUntagged?: boolean;
  untaggedOnly?: boolean;
  untaggedLabel?: string;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  placeholder: "Filter...",
  supportsUntagged: false,
  untaggedOnly: false,
  untaggedLabel: "Untagged",
});

const emit = defineEmits<{
  "update:modelValue": [value: string];
  "update:untaggedOnly": [value: boolean];
}>();

const inputRef = ref<InstanceType<typeof BaseInput> | null>(null);
const isOpen = ref(false);
const isInputFocused = ref(false);
const localValue = ref(props.modelValue);
const localUntaggedOnly = ref(props.untaggedOnly);

watch(
  () => props.modelValue,
  (nextValue) => {
    if (!isInputFocused.value && nextValue !== localValue.value) {
      localValue.value = nextValue;
    }
  },
);

watch(
  () => props.untaggedOnly,
  (nextValue) => {
    if (nextValue !== localUntaggedOnly.value) {
      localUntaggedOnly.value = nextValue;
    }
  },
);

watch(localValue, (value) => {
  emit("update:modelValue", value);
});

watch(localUntaggedOnly, (value) => {
  emit("update:untaggedOnly", value);
  if (value && localValue.value) {
    localValue.value = "";
  }
});

watch(isOpen, (open) => {
  if (!open) {
    return;
  }

  void nextTick(() => {
    if (localUntaggedOnly.value && props.supportsUntagged) {
      return;
    }
    inputRef.value?.focus?.();
    inputRef.value?.select?.();
  });
});

function clearValue(): void {
  localValue.value = "";
  void nextTick(() => {
    inputRef.value?.focus?.();
  });
}
</script>

<template>
  <BasePopover
    v-model:open="isOpen"
    :disabled="loading"
    panel-class="w-56 rounded-lg border border-slate-200 bg-white p-2 shadow-lg"
  >
    <template #trigger="{ toggle }">
      <IconButton
        :disabled="loading"
        :title="`Filter ${label}`"
        :tone="modelValue || localUntaggedOnly ? 'primary' : 'neutral'"
        @click="toggle"
      >
        <FunnelIcon class="h-4 w-4" />
      </IconButton>
    </template>

    <template #default="{ close }">
      <label class="mb-1 block text-[11px] font-semibold uppercase tracking-wide text-slate-500">
        {{ label }}
      </label>
      <div class="relative">
        <BaseInput
          ref="inputRef"
          v-model="localValue"
          type="text"
          rounded="pill"
          :placeholder="placeholder"
          :disabled="loading || (supportsUntagged && localUntaggedOnly)"
          class="w-full border-slate-200 bg-white/90 px-3 py-2 pr-8 placeholder:text-slate-400 focus:border-blue-400 focus:ring-blue-200"
          @focus="isInputFocused = true"
          @blur="isInputFocused = false"
          @keydown.esc.prevent="close"
        />
        <InputClearButton
          v-if="localValue"
          size="sm"
          class="absolute right-2 top-2"
          title="Clear"
          @click="clearValue"
          @mousedown.prevent
        />
      </div>
      <!-- <label
        v-if="supportsUntagged"
        class="mt-2 flex cursor-pointer items-center gap-2 text-xs text-slate-600"
      >
        <span>Untagged</span>
        <BaseToggle
          :checked="localUntaggedOnly"
          :disabled="loading"
          @update:checked="localUntaggedOnly = $event"
        />
      </label> -->
      <label 
        v-if="supportsUntagged"
        class="flex mt-2 cursor-pointer items-center justify-between gap-2 rounded-md bg-slate-50 px-2 py-2 text-xs text-slate-700"
      >
        <span>Untagged only</span>
        <BaseToggle
          :checked="localUntaggedOnly"
          :disabled="loading"
          @update:checked="localUntaggedOnly = $event"
        />
      </label>
    </template>
  </BasePopover>
</template>

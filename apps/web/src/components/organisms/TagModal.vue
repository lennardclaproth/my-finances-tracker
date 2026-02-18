<script setup lang="ts">
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { ref, watch } from "vue";
import BaseButton from "../atoms/BaseButton.vue";
import BaseInput from "../atoms/BaseInput.vue";

interface Props {
  open: boolean;
  loading?: boolean;
  targetLabel: string;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  close: [];
  confirm: [tag: string];
}>();

const tag = ref("");

watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      tag.value = "";
    }
  },
);

function submit(): void {
  emit("confirm", tag.value.trim());
}
</script>

<template>
  <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4" @click.self="emit('close')">
    <div class="w-full max-w-md rounded-lg border border-slate-200 bg-white p-4 shadow-xl">
      <div class="mb-4 flex items-start justify-between">
        <div>
          <h2 class="text-lg font-semibold text-slate-900">Apply Tag</h2>
          <p class="text-sm text-slate-500">{{ targetLabel }}</p>
        </div>
        <button
          type="button"
          class="rounded p-1 text-slate-500 hover:bg-slate-100 hover:text-slate-700"
          :disabled="loading"
          @click="emit('close')"
        >
          <XMarkIcon class="h-5 w-5" />
        </button>
      </div>

      <label class="space-y-1">
        <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Tag</span>
        <BaseInput
          :model-value="tag"
          :disabled="loading"
          placeholder="e.g. groceries"
          @update:model-value="tag = $event"
          @keydown.enter="submit"
        />
      </label>

      <div class="mt-4 flex justify-end gap-2">
        <BaseButton :disabled="loading" variant="ghost" @click="emit('close')">Cancel</BaseButton>
        <BaseButton :disabled="loading || tag.trim() === ''" variant="primary" @click="submit">Save Tag</BaseButton>
      </div>
    </div>
  </div>
</template>
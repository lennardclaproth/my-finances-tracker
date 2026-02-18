<script setup lang="ts">
import {
  CheckCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
  XMarkIcon,
} from "@heroicons/vue/24/outline";
import { computed } from "vue";

type ToastTone = "success" | "danger" | "info";

interface Props {
  tone: ToastTone;
  message: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
}>();

const toneClass = computed(() => {
  const mapping: Record<ToastTone, string> = {
    success: "border-emerald-300 bg-emerald-50 text-emerald-900",
    danger: "border-rose-300 bg-rose-50 text-rose-900",
    info: "border-cyan-300 bg-cyan-50 text-cyan-900",
  };
  return mapping[props.tone];
});
</script>

<template>
  <div class="pointer-events-auto w-full max-w-md rounded-lg border px-3 py-2 shadow-lg" :class="toneClass">
    <div class="flex items-start gap-2">
      <CheckCircleIcon v-if="tone === 'success'" class="mt-0.5 h-5 w-5 shrink-0" />
      <ExclamationTriangleIcon v-else-if="tone === 'danger'" class="mt-0.5 h-5 w-5 shrink-0" />
      <InformationCircleIcon v-else class="mt-0.5 h-5 w-5 shrink-0" />
      <p class="flex-1 text-sm">{{ message }}</p>
      <button
        type="button"
        class="rounded p-0.5 opacity-70 transition hover:bg-white/50 hover:opacity-100"
        title="Dismiss"
        aria-label="Dismiss"
        @click="emit('close')"
      >
        <XMarkIcon class="h-4 w-4" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { computed, ref, watch } from "vue";
import BaseButton from "../atoms/BaseButton.vue";
import BaseInput from "../atoms/BaseInput.vue";

interface Props {
  open: boolean;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  close: [];
  submit: [name: string];
}>();

const name = ref("");
const errorMessage = ref("");

const canSubmit = computed(() => !props.loading && name.value.trim() !== "");

watch(
  () => props.open,
  (open) => {
    if (!open) {
      name.value = "";
      errorMessage.value = "";
    }
  },
);

function onSubmit(): void {
  if (name.value.trim() === "") {
    errorMessage.value = "Class name is required.";
    return;
  }
  errorMessage.value = "";
  emit("submit", name.value.trim());
}
</script>

<template>
  <div
    v-if="open"
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4"
    @click.self="emit('close')"
  >
    <section class="w-full max-w-lg rounded-2xl border border-slate-200 bg-white shadow-2xl">
      <header class="flex items-start justify-between border-b border-slate-100 px-5 py-4">
        <div>
          <h2 class="text-lg font-semibold text-slate-900">Create asset class</h2>
          <p class="text-sm text-slate-500">Set up a category like Property, Art, or Savings.</p>
        </div>
        <button
          type="button"
          class="rounded p-1 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
          title="Close create asset class modal"
          @click="emit('close')"
        >
          <XMarkIcon class="h-5 w-5" />
        </button>
      </header>

      <div class="space-y-3 px-5 py-4">
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Class name *</span>
          <BaseInput
            :model-value="name"
            placeholder="Property"
            @update:model-value="name = $event"
          />
        </label>
        <p v-if="errorMessage" class="rounded-md border border-rose-200 bg-rose-50 px-2 py-1 text-xs text-rose-700">
          {{ errorMessage }}
        </p>
      </div>

      <footer class="flex justify-end gap-2 border-t border-slate-100 px-5 py-4">
        <BaseButton variant="ghost" :disabled="loading" @click="emit('close')">Cancel</BaseButton>
        <BaseButton variant="primary" :disabled="!canSubmit" @click="onSubmit">
          {{ loading ? "Creating..." : "Create class" }}
        </BaseButton>
      </footer>
    </section>
  </div>
</template>

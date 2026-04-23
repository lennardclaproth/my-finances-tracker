<script setup lang="ts">
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { ref, watch } from "vue";
import type { AssetClass } from "../../types/assets";
import BaseButton from "../atoms/BaseButton.vue";
import BaseInput from "../atoms/BaseInput.vue";

interface Props {
  open: boolean;
  row: AssetClass | null;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  close: [];
  save: [payload: { name: string; archived: boolean }];
  delete: [];
}>();

const name = ref("");
const archived = ref(false);
const errorMessage = ref("");

watch(
  () => [props.open, props.row?.id] as const,
  () => {
    if (!props.open || !props.row) {
      return;
    }
    name.value = props.row.name;
    archived.value = props.row.archived;
    errorMessage.value = "";
  },
  { immediate: true },
);

function onSave(): void {
  if (name.value.trim() === "") {
    errorMessage.value = "Class name is required.";
    return;
  }
  errorMessage.value = "";
  emit("save", {
    name: name.value.trim(),
    archived: archived.value,
  });
}

function onDelete(): void {
  if (!confirm("Delete this class and all assets in it?")) {
    return;
  }
  emit("delete");
}
</script>

<template>
  <div
    v-if="open && row"
    class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/50 p-4"
    @click.self="emit('close')"
  >
    <section class="w-full max-w-lg rounded-2xl border border-slate-200 bg-white shadow-2xl">
      <header class="flex items-start justify-between border-b border-slate-100 px-5 py-4">
        <div>
          <h2 class="text-lg font-semibold text-slate-900">Edit asset class</h2>
          <p class="text-sm text-slate-500">Manage class name, state, and lifecycle.</p>
        </div>
        <button
          type="button"
          class="rounded p-1 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
          title="Close class settings modal"
          @click="emit('close')"
        >
          <XMarkIcon class="h-5 w-5" />
        </button>
      </header>

      <div class="space-y-3 px-5 py-4">
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Class name</span>
          <BaseInput :model-value="name" @update:model-value="name = $event" />
        </label>
        <label class="flex items-center justify-between rounded-md border border-slate-200 px-3 py-2">
          <span class="text-sm font-medium text-slate-700">Archived</span>
          <input v-model="archived" type="checkbox" class="h-4 w-4 rounded border-slate-300 text-indigo-600" />
        </label>
        <p v-if="errorMessage" class="text-xs text-rose-700">{{ errorMessage }}</p>
      </div>

      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-4">
        <BaseButton variant="danger" size="sm" :disabled="loading" @click="onDelete">Delete class</BaseButton>
        <div class="flex items-center gap-2">
          <BaseButton variant="ghost" size="sm" :disabled="loading" @click="emit('close')">Cancel</BaseButton>
          <BaseButton variant="primary" size="sm" :disabled="loading" @click="onSave">
            {{ loading ? "Saving..." : "Save changes" }}
          </BaseButton>
        </div>
      </footer>
    </section>
  </div>
</template>

<script setup lang="ts">
import { CloudArrowUpIcon, DocumentTextIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { computed, ref, watch } from "vue";
import { ApiError } from "../../services/http";
import { uploadListingDailies } from "../../services/marketdata";
import type { Listing } from "../../types/listings";
import type { UploadListingDailiesAcceptedResponse } from "../../types/marketdata";
import BaseButton from "../atoms/BaseButton.vue";
import ListingSearchSelect from "./ListingSearchSelect.vue";

interface Props {
  open: boolean;
  initialListing?: Listing | null;
}

const props = withDefaults(defineProps<Props>(), {
  initialListing: null,
});

const emit = defineEmits<{
  close: [];
  uploaded: [payload: UploadListingDailiesAcceptedResponse];
}>();

const fileInputRef = ref<HTMLInputElement | null>(null);
const selectedListing = ref<Listing | null>(null);
const selectedListingId = ref("");
const selectedFile = ref<File | null>(null);
const submitting = ref(false);
const errorMessage = ref("");

const canSubmit = computed(() => {
  return selectedListingId.value !== "" && selectedFile.value !== null && !submitting.value;
});

watch(
  () => props.open,
  (open) => {
    if (!open) {
      resetState();
      return;
    }
    resetState();
    if (props.initialListing) {
      selectedListing.value = props.initialListing;
      selectedListingId.value = props.initialListing.id;
    }
  },
);

function resetState(): void {
  selectedListing.value = null;
  selectedListingId.value = "";
  selectedFile.value = null;
  submitting.value = false;
  errorMessage.value = "";
  if (fileInputRef.value) {
    fileInputRef.value.value = "";
  }
}

function openFilePicker(): void {
  if (submitting.value) {
    return;
  }
  fileInputRef.value?.click();
}

function onFileChange(event: Event): void {
  const input = event.target as HTMLInputElement;
  selectedFile.value = input.files?.[0] ?? null;
  errorMessage.value = "";
}

function clearFile(): void {
  selectedFile.value = null;
  if (fileInputRef.value) {
    fileInputRef.value.value = "";
  }
}

async function submit(): Promise<void> {
  if (!canSubmit.value || !selectedFile.value) {
    return;
  }

  submitting.value = true;
  errorMessage.value = "";
  try {
    const accepted = await uploadListingDailies({
      listingId: selectedListingId.value,
      file: selectedFile.value,
    });
    emit("uploaded", accepted);
    emit("close");
  } catch (error: unknown) {
    if (error instanceof ApiError) {
      errorMessage.value = error.message;
    } else if (error instanceof Error) {
      errorMessage.value = error.message;
    } else {
      errorMessage.value = "Failed to upload daily data file.";
    }
  } finally {
    submitting.value = false;
  }
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
        <div class="space-y-1">
          <h2 class="font-secondary text-lg font-semibold text-slate-900">Upload daily data</h2>
          <p class="text-sm text-slate-500">
            Select a listing and upload a CSV/TXT file for async processing.
          </p>
        </div>
        <button
          type="button"
          class="rounded p-1 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
          title="Close upload daily data modal"
          @click="emit('close')"
        >
          <XMarkIcon class="h-5 w-5" />
        </button>
      </header>

      <div class="space-y-4 px-5 py-4">
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Listing</span>
          <ListingSearchSelect
            :model-value="selectedListingId"
            :selected-listing="selectedListing"
            :disabled="submitting"
            @update:model-value="selectedListingId = $event"
            @select="selectedListing = $event"
          />
        </label>

        <input
          ref="fileInputRef"
          type="file"
          class="hidden"
          accept=".csv,.txt,text/csv,text/plain"
          :disabled="submitting"
          @change="onFileChange"
        >

        <button
          type="button"
          class="flex w-full items-center justify-center gap-2 rounded-xl border border-dashed border-slate-300 bg-slate-50 px-4 py-6 text-sm text-slate-600 transition hover:border-indigo-300 hover:bg-indigo-50 hover:text-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
          :disabled="submitting"
          @click="openFilePicker"
        >
          <CloudArrowUpIcon class="h-5 w-5" />
          <span>{{ selectedFile ? "Replace file" : "Select CSV/TXT file to upload" }}</span>
        </button>

        <div
          v-if="selectedFile"
          class="flex items-center justify-between rounded-xl border border-slate-200 bg-white px-3 py-2"
        >
          <div class="flex min-w-0 items-center gap-2">
            <DocumentTextIcon class="h-4 w-4 shrink-0 text-slate-500" />
            <div class="min-w-0">
              <p class="truncate text-sm font-medium text-slate-700">{{ selectedFile.name }}</p>
              <p class="text-xs text-slate-500">{{ (selectedFile.size / 1024).toFixed(1) }} KB</p>
            </div>
          </div>
          <BaseButton variant="ghost" size="sm" :disabled="submitting" @click="clearFile">Remove</BaseButton>
        </div>

        <p v-if="errorMessage" class="rounded-md border border-rose-200 bg-rose-50 px-2 py-1 text-xs text-rose-700">
          {{ errorMessage }}
        </p>

        <p class="text-xs text-slate-500">
          Supported format: .csv or .txt (max 10 MB).
        </p>
      </div>

      <footer class="flex justify-end gap-2 border-t border-slate-100 px-5 py-4">
        <BaseButton variant="ghost" :disabled="submitting" @click="emit('close')">Cancel</BaseButton>
        <BaseButton :disabled="!canSubmit" variant="primary" @click="void submit()">
          {{ submitting ? "Uploading..." : "Upload" }}
        </BaseButton>
      </footer>
    </section>
  </div>
</template>

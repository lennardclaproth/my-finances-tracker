<script setup lang="ts">
import { MagnifyingGlassIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { searchListings } from "../../services/listings";
import type { Listing } from "../../types/listings";
import { ApiError } from "../../services/http";
import BaseInput from "../atoms/BaseInput.vue";

interface Props {
  modelValue: string;
  selectedListing?: Listing | null;
  disabled?: boolean;
  minChars?: number;
  debounceMs?: number;
}

const props = withDefaults(defineProps<Props>(), {
  selectedListing: null,
  disabled: false,
  minChars: 2,
  debounceMs: 300,
});

const emit = defineEmits<{
  "update:modelValue": [value: string];
  select: [listing: Listing | null];
}>();

const query = ref("");
const loading = ref(false);
const errorMessage = ref("");
const open = ref(false);
const results = ref<Listing[]>([]);
let debounceHandle: ReturnType<typeof setTimeout> | null = null;
let activeRequestId = 0;

const hasSelection = computed(() => Boolean(props.modelValue) && Boolean(props.selectedListing));
const showClearSelection = computed(() => !props.disabled && hasSelection.value);

watch(
  () => props.selectedListing,
  (listing) => {
    if (!listing || query.value.trim() !== "") {
      return;
    }
    query.value = `${listing.symbol} - ${listing.name}`;
  },
  { immediate: true },
);

watch(
  query,
  (value) => {
    if (props.selectedListing) {
      const selectedLabel = `${props.selectedListing.symbol} - ${props.selectedListing.name}`;
      if (value !== selectedLabel) {
        emit("update:modelValue", "");
        emit("select", null);
      }
    }
    if (debounceHandle) {
      clearTimeout(debounceHandle);
      debounceHandle = null;
    }
    const trimmed = value.trim();
    if (trimmed.length < props.minChars) {
      results.value = [];
      errorMessage.value = "";
      open.value = false;
      return;
    }
    debounceHandle = setTimeout(() => {
      void runSearch(trimmed);
    }, props.debounceMs);
  },
);

async function runSearch(term: string): Promise<void> {
  const requestId = ++activeRequestId;
  loading.value = true;
  errorMessage.value = "";
  open.value = true;
  try {
    const response = await searchListings(term, 25, 0);
    if (requestId !== activeRequestId) {
      return;
    }
    results.value = response.data;
    if (response.data.length === 0) {
      errorMessage.value = "No listings found.";
    }
  } catch (error: unknown) {
    if (requestId !== activeRequestId) {
      return;
    }
    if (error instanceof ApiError) {
      errorMessage.value = error.message;
    } else if (error instanceof Error) {
      errorMessage.value = error.message;
    } else {
      errorMessage.value = "Failed to search listings.";
    }
    results.value = [];
  } finally {
    if (requestId === activeRequestId) {
      loading.value = false;
    }
  }
}

function pickListing(listing: Listing): void {
  emit("update:modelValue", listing.id);
  emit("select", listing);
  query.value = `${listing.symbol} - ${listing.name}`;
  results.value = [];
  errorMessage.value = "";
  open.value = false;
}

function clearSelection(): void {
  emit("update:modelValue", "");
  emit("select", null);
  query.value = "";
  results.value = [];
  errorMessage.value = "";
  open.value = false;
}

function onFocus(): void {
  if (results.value.length > 0 || loading.value || errorMessage.value !== "") {
    open.value = true;
  }
}

onBeforeUnmount(() => {
  if (debounceHandle) {
    clearTimeout(debounceHandle);
    debounceHandle = null;
  }
});
</script>

<template>
  <div class="relative">
    <div class="relative">
      <MagnifyingGlassIcon class="pointer-events-none absolute left-3 top-2.5 h-5 w-5 text-slate-400" />
      <BaseInput
        v-model="query"
        type="search"
        rounded="default"
        :disabled="disabled"
        placeholder="Search listing by symbol, name or ISIN"
        class="w-full pl-10 pr-10"
        @focus="onFocus"
      />
      <button
        v-if="showClearSelection"
        type="button"
        class="absolute right-2 top-2 rounded p-1 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
        title="Clear listing"
        @click="clearSelection"
      >
        <XMarkIcon class="h-4 w-4" />
      </button>
    </div>

    <div
      v-if="open"
      class="absolute z-40 mt-1 max-h-60 w-full overflow-auto rounded-md border border-slate-200 bg-white shadow-lg"
    >
      <p v-if="loading" class="px-3 py-2 text-xs text-slate-500">Searching listings...</p>
      <p v-else-if="errorMessage" class="px-3 py-2 text-xs text-slate-500">{{ errorMessage }}</p>
      <template v-else>
        <p v-if="results.length === 0" class="px-3 py-2 text-xs text-slate-500">Type at least 2 characters.</p>
        <button
          v-for="listing in results"
          :key="listing.id"
          type="button"
          class="block w-full border-b border-slate-100 px-3 py-2 text-left text-sm transition hover:bg-slate-50"
          @click="pickListing(listing)"
        >
          <p class="font-medium text-slate-900">{{ listing.symbol }} - {{ listing.name }}</p>
          <p class="text-xs text-slate-500">
            ISIN: {{ listing.isin || "-" }} | {{ listing.exchange || "-" }} | {{ listing.type || "-" }}
          </p>
        </button>
      </template>
    </div>
  </div>
</template>

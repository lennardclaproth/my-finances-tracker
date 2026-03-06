<script setup lang="ts">
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { computed, ref, watch } from "vue";
import type { CreateListingPayload, Listing } from "../../types/listings";
import { createListing } from "../../services/listings";
import { ApiError } from "../../services/http";
import BaseButton from "../atoms/BaseButton.vue";
import BaseInput from "../atoms/BaseInput.vue";
import BaseSelect from "../atoms/BaseSelect.vue";

interface Props {
  open: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  created: [listing: Listing];
}>();

const sourceOptions = [
  { label: "Alpha Vantage", value: "alpha_vantage" },
  { label: "MarketStack", value: "market_stack" },
  { label: "Brand New Day", value: "brandnewday" },
];

const currencyOptions = [
  { label: "No currency", value: "" },
  { label: "EUR", value: "EUR" },
  { label: "USD", value: "USD" },
  { label: "GBP", value: "GBP" },
  { label: "JPY", value: "JPY" },
];

const allowedCurrencies = new Set(["EUR", "USD", "GBP", "JPY"]);

const name = ref("");
const symbol = ref("");
const source = ref("alpha_vantage");
const description = ref("");
const exchange = ref("");
const region = ref("");
const currency = ref("");
const isin = ref("");
const ticker = ref("");
const type = ref("");
const submitting = ref(false);
const errorMessage = ref("");

const canSubmit = computed(() => {
  return !submitting.value && name.value.trim() !== "" && symbol.value.trim() !== "" && source.value.trim() !== "";
});

watch(
  () => props.open,
  (open) => {
    if (!open) {
      reset();
    }
  },
);

function reset(): void {
  name.value = "";
  symbol.value = "";
  source.value = "alpha_vantage";
  description.value = "";
  exchange.value = "";
  region.value = "";
  currency.value = "";
  isin.value = "";
  ticker.value = "";
  type.value = "";
  submitting.value = false;
  errorMessage.value = "";
}

function validate(): string | null {
  if (name.value.trim() === "") {
    return "Name is required.";
  }
  if (symbol.value.trim() === "") {
    return "Symbol is required.";
  }
  if (source.value.trim() === "") {
    return "Source is required.";
  }
  if (currency.value.trim() !== "" && !allowedCurrencies.has(currency.value.trim().toUpperCase())) {
    return "Currency must be one of EUR, USD, GBP, or JPY.";
  }
  return null;
}

function toPayload(): CreateListingPayload {
  return {
    name: name.value,
    symbol: symbol.value.toUpperCase(),
    source: source.value,
    description: description.value,
    exchange: exchange.value,
    region: region.value,
    currency: currency.value.toUpperCase(),
    isin: isin.value.toUpperCase(),
    ticker: ticker.value.toUpperCase(),
    type: type.value,
  };
}

function mapApiError(error: ApiError): string {
  if (error.status === 400) {
    return error.message || "Invalid listing input.";
  }
  if (error.status === 409) {
    return "A listing with this symbol and source already exists.";
  }
  return "Failed to create listing.";
}

async function submit(): Promise<void> {
  const validationError = validate();
  if (validationError) {
    errorMessage.value = validationError;
    return;
  }
  submitting.value = true;
  errorMessage.value = "";
  try {
    const listing = await createListing(toPayload());
    emit("created", listing);
    emit("close");
  } catch (error: unknown) {
    if (error instanceof ApiError) {
      errorMessage.value = mapApiError(error);
    } else if (error instanceof Error) {
      errorMessage.value = error.message;
    } else {
      errorMessage.value = "Failed to create listing.";
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
    <section class="w-full max-w-2xl rounded-2xl border border-slate-200 bg-white shadow-2xl">
      <header class="flex items-start justify-between border-b border-slate-100 px-5 py-4">
        <div class="space-y-1">
          <h2 class="text-lg font-semibold text-slate-900">Add listing</h2>
          <p class="text-sm text-slate-500">
            Create a listing to make it available in market data and portfolio flows.
          </p>
        </div>
        <button
          type="button"
          class="rounded p-1 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
          title="Close create listing modal"
          @click="emit('close')"
        >
          <XMarkIcon class="h-5 w-5" />
        </button>
      </header>

      <div class="grid max-h-[70vh] grid-cols-1 gap-3 overflow-y-auto px-5 py-4 md:grid-cols-2">
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Name *</span>
          <BaseInput :model-value="name" placeholder="VanEck AEX UCITS ETF" @update:model-value="name = $event" />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Symbol *</span>
          <BaseInput :model-value="symbol" placeholder="TDT.AS" @update:model-value="symbol = $event" />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Source *</span>
          <BaseSelect :model-value="source" :options="sourceOptions" rounded="default" @update:model-value="source = $event" />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Currency</span>
          <BaseSelect :model-value="currency" :options="currencyOptions" rounded="default" @update:model-value="currency = $event" />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">ISIN</span>
          <BaseInput :model-value="isin" placeholder="NL..." @update:model-value="isin = $event" />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Ticker</span>
          <BaseInput :model-value="ticker" placeholder="TDT" @update:model-value="ticker = $event" />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Exchange</span>
          <BaseInput :model-value="exchange" placeholder="XAMS" @update:model-value="exchange = $event" />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Region</span>
          <BaseInput :model-value="region" placeholder="NL" @update:model-value="region = $event" />
        </label>
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Type</span>
          <BaseInput :model-value="type" placeholder="ETF" @update:model-value="type = $event" />
        </label>
        <label class="block space-y-1 md:col-span-2">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Description</span>
          <BaseInput :model-value="description" placeholder="Optional listing description" @update:model-value="description = $event" />
        </label>
        <p v-if="errorMessage" class="rounded-md border border-rose-200 bg-rose-50 px-2 py-1 text-xs text-rose-700 md:col-span-2">
          {{ errorMessage }}
        </p>
      </div>

      <footer class="flex justify-end gap-2 border-t border-slate-100 px-5 py-4">
        <BaseButton variant="ghost" :disabled="submitting" @click="emit('close')">Cancel</BaseButton>
        <BaseButton variant="primary" :disabled="!canSubmit" @click="void submit()">
          {{ submitting ? "Creating..." : "Create listing" }}
        </BaseButton>
      </footer>
    </section>
  </div>
</template>

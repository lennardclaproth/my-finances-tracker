<script setup lang="ts">
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { computed, ref, watch } from "vue";
import { ApiError } from "../../services/http";
import { createManualPortfolioTransaction } from "../../services/portfolio";
import { fetchVendors } from "../../services/vendors";
import type { Listing } from "../../types/listings";
import type { CreateManualPortfolioTransactionPayload, PortfolioTransaction } from "../../types/portfolio";
import type { Vendor } from "../../types/vendors";
import BaseButton from "../atoms/BaseButton.vue";
import BaseInput from "../atoms/BaseInput.vue";
import BaseSelect from "../atoms/BaseSelect.vue";
import BaseToggle from "../atoms/BaseToggle.vue";
import ListingSearchSelect from "./ListingSearchSelect.vue";

type ManualTransactionType = "BUY" | "SELL" | "DIVIDEND" | "TAX" | "FEE" | "CASH";

interface Props {
  open: boolean;
  accountId: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  created: [transaction: PortfolioTransaction];
}>();

const transactionTypeOptions = [
  { label: "Buy", value: "BUY" },
  { label: "Sell", value: "SELL" },
  { label: "Dividend", value: "DIVIDEND" },
  { label: "Tax", value: "TAX" },
  { label: "Fee", value: "FEE" },
  { label: "Cash", value: "CASH" },
];

const vendors = ref<Vendor[]>([]);
const loadingVendors = ref(false);
const submitting = ref(false);
const errorMessage = ref("");

const vendorId = ref("");
const occurredAt = ref("");
const type = ref<ManualTransactionType>("BUY");
const listingId = ref("");
const selectedListing = ref<Listing | null>(null);
const buySellMode = ref<"total" | "unit">("total");
const amount = ref("");
const unitPrice = ref("");
const quantity = ref("");
const description = ref("");

const decimalPattern = /^-?\d+(\.\d{1,6})?$/;

const vendorOptions = computed(() => [
  { label: "Select vendor", value: "" },
  ...vendors.value.map((vendor) => ({
    label: `${vendor.name} (${vendor.type})`,
    value: vendor.id,
  })),
]);

const isCash = computed(() => type.value === "CASH");
const isBuySell = computed(() => type.value === "BUY" || type.value === "SELL");
const isUnitPriceMode = computed(() => isBuySell.value && buySellMode.value === "unit");

const canSubmit = computed(() => {
  return !submitting.value && !loadingVendors.value && validate() === null;
});

const computedBuySellTotal = computed(() => {
  if (!isUnitPriceMode.value) {
    return "";
  }
  const quantityValue = Number.parseFloat(quantity.value.trim());
  const unitPriceValue = Number.parseFloat(unitPrice.value.trim());
  if (Number.isNaN(quantityValue) || Number.isNaN(unitPriceValue)) {
    return "";
  }
  if (quantityValue <= 0 || unitPriceValue <= 0) {
    return "";
  }
  return formatDecimalValue(quantityValue * unitPriceValue);
});

watch(
  () => props.open,
  async (open) => {
    if (!open) {
      resetState();
      return;
    }
    resetState();
    await loadVendors();
  },
);

watch(type, (nextType) => {
  if (nextType === "CASH") {
    listingId.value = "";
    selectedListing.value = null;
    quantity.value = "";
    unitPrice.value = "";
  }
  if (!isBuySell.value) {
    quantity.value = "";
    unitPrice.value = "";
    buySellMode.value = "total";
  }
  errorMessage.value = "";
});

function toLocalDateInput(value: Date): string {
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function resetState(): void {
  vendors.value = [];
  loadingVendors.value = false;
  submitting.value = false;
  errorMessage.value = "";
  vendorId.value = "";
  occurredAt.value = toLocalDateInput(new Date());
  type.value = "BUY";
  listingId.value = "";
  selectedListing.value = null;
  buySellMode.value = "total";
  amount.value = "";
  unitPrice.value = "";
  quantity.value = "";
  description.value = "";
}

async function loadVendors(): Promise<void> {
  loadingVendors.value = true;
  errorMessage.value = "";
  try {
    const allVendors = await fetchVendors();
    vendors.value = allVendors.filter((vendor) => vendor.active && vendor.type === "brokerage");
  } catch (error: unknown) {
    if (error instanceof ApiError) {
      errorMessage.value = error.message;
    } else if (error instanceof Error) {
      errorMessage.value = error.message;
    } else {
      errorMessage.value = "Failed to load vendors.";
    }
  } finally {
    loadingVendors.value = false;
  }
}

function validateDecimal(raw: string): boolean {
  return decimalPattern.test(raw.trim());
}

function formatDecimalValue(value: number): string {
  const rounded = Number.parseFloat(value.toFixed(6));
  const raw = rounded.toFixed(6);
  return raw.replace(/\.?0+$/, "");
}

function validate(): string | null {
  if (vendorId.value.trim() === "") {
    return "Vendor is required.";
  }
  if (occurredAt.value.trim() === "") {
    return "Occurred date is required.";
  }
  if (isCash.value) {
    if (amount.value.trim() === "") {
      return "Amount is required.";
    }
    if (!validateDecimal(amount.value)) {
      return "Amount must be a decimal with up to 6 decimals.";
    }
    const parsedAmount = Number.parseFloat(amount.value);
    if (Number.isNaN(parsedAmount)) {
      return "Amount must be valid.";
    }
    if (parsedAmount === 0) {
      return "Cash amount must be non-zero.";
    }
    return null;
  }

  if (listingId.value.trim() === "") {
    return "Listing is required for non-cash transactions.";
  }

  if (isBuySell.value) {
    if (quantity.value.trim() === "") {
      return "Quantity is required for buy/sell.";
    }
    if (!validateDecimal(quantity.value)) {
      return "Quantity must be a decimal with up to 6 decimals.";
    }
    const parsedQuantity = Number.parseFloat(quantity.value);
    if (Number.isNaN(parsedQuantity) || parsedQuantity <= 0) {
      return "Quantity must be positive.";
    }

    if (isUnitPriceMode.value) {
      if (unitPrice.value.trim() === "") {
        return "Unit price is required.";
      }
      if (!validateDecimal(unitPrice.value)) {
        return "Unit price must be a decimal with up to 6 decimals.";
      }
      const parsedUnitPrice = Number.parseFloat(unitPrice.value);
      if (Number.isNaN(parsedUnitPrice) || parsedUnitPrice <= 0) {
        return "Unit price must be positive.";
      }
      return null;
    }
  }

  if (amount.value.trim() === "") {
    return "Amount is required.";
  }
  if (!validateDecimal(amount.value)) {
    return "Amount must be a decimal with up to 6 decimals.";
  }
  const parsedAmount = Number.parseFloat(amount.value);
  if (Number.isNaN(parsedAmount)) {
    return "Amount must be valid.";
  }
  if (parsedAmount <= 0) {
    return "Amount must be positive for non-cash transactions.";
  }
  return null;
}

function toPayload(): CreateManualPortfolioTransactionPayload {
  let amountValue = amount.value.trim();
  if (isBuySell.value && isUnitPriceMode.value) {
    const quantityValue = Number.parseFloat(quantity.value.trim());
    const unitPriceValue = Number.parseFloat(unitPrice.value.trim());
    amountValue = formatDecimalValue(quantityValue * unitPriceValue);
  }
  const payload: CreateManualPortfolioTransactionPayload = {
    account_id: props.accountId,
    vendor_id: vendorId.value.trim(),
    occurred_at: occurredAt.value.trim(),
    type: type.value,
    amount: amountValue,
  };
  if (!isCash.value) {
    payload.listing_id = listingId.value.trim();
  }
  if (isBuySell.value) {
    payload.quantity = quantity.value.trim();
  }
  const descriptionValue = description.value.trim();
  if (descriptionValue !== "") {
    payload.description = descriptionValue;
  }
  return payload;
}

function mapApiError(error: ApiError): string {
  if (error.status === 400) {
    return error.message || "Invalid transaction input.";
  }
  if (error.status === 404) {
    return error.message || "Account, vendor, or listing was not found.";
  }
  if (error.status === 409) {
    return "Duplicate transaction detected.";
  }
  if (error.status === 422) {
    return error.message || "Vendor type is not allowed for portfolio transactions.";
  }
  return "Failed to create transaction.";
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
    const created = await createManualPortfolioTransaction(toPayload());
    emit("created", created);
    emit("close");
  } catch (error: unknown) {
    if (error instanceof ApiError) {
      errorMessage.value = mapApiError(error);
    } else if (error instanceof Error) {
      errorMessage.value = error.message;
    } else {
      errorMessage.value = "Failed to create transaction.";
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
          <h2 class="text-lg font-semibold text-slate-900">Add transaction</h2>
          <p class="text-sm text-slate-500">Create a manual portfolio transaction.</p>
        </div>
        <button
          type="button"
          class="rounded p-1 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
          title="Close create transaction modal"
          @click="emit('close')"
        >
          <XMarkIcon class="h-5 w-5" />
        </button>
      </header>

      <div class="grid max-h-[70vh] grid-cols-1 gap-3 overflow-y-auto px-5 py-4 md:grid-cols-2">
        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Vendor *</span>
          <BaseSelect
            :model-value="vendorId"
            :options="vendorOptions"
            rounded="default"
            :disabled="loadingVendors || submitting"
            @update:model-value="vendorId = $event"
          />
        </label>

        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Occurred at *</span>
          <BaseInput
            :model-value="occurredAt"
            type="date"
            :disabled="submitting"
            @update:model-value="occurredAt = $event"
          />
        </label>

        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Type *</span>
          <BaseSelect
            :model-value="type"
            :options="transactionTypeOptions"
            rounded="default"
            :disabled="submitting"
            @update:model-value="type = $event as ManualTransactionType"
          />
        </label>

        <div v-if="isBuySell" class="flex items-center justify-between rounded-md border border-slate-200 bg-slate-50 px-3 py-2 md:col-span-2">
          <div class="space-y-0.5">
            <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">Input Mode</p>
            <p class="text-xs text-slate-600">
              {{ isUnitPriceMode ? "Unit price x quantity" : "Total amount" }}
            </p>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-slate-600">Total</span>
            <BaseToggle
              :checked="buySellMode === 'unit'"
              :disabled="submitting"
              @update:checked="buySellMode = $event ? 'unit' : 'total'"
            />
            <span class="text-xs text-slate-600">Unit</span>
          </div>
        </div>

        <label class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">
            {{ isUnitPriceMode ? "Unit Price *" : "Amount *" }} {{ isCash ? "(signed)" : "" }}
          </span>
          <BaseInput
            :model-value="isUnitPriceMode ? unitPrice : amount"
            type="text"
            :placeholder="isCash ? '-10.50 or 10.50' : '10.50'"
            :disabled="submitting"
            @update:model-value="isUnitPriceMode ? (unitPrice = $event) : (amount = $event)"
          />
        </label>

        <label v-if="isBuySell" class="block space-y-1">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Quantity *</span>
          <BaseInput
            :model-value="quantity"
            type="text"
            placeholder="1.000000"
            :disabled="submitting"
            @update:model-value="quantity = $event"
          />
          <p v-if="isUnitPriceMode && computedBuySellTotal" class="text-xs text-slate-500">
            Computed total amount: {{ computedBuySellTotal }}
          </p>
        </label>

        <label v-if="!isCash" class="block space-y-1 md:col-span-2">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Listing *</span>
          <ListingSearchSelect
            :model-value="listingId"
            :selected-listing="selectedListing"
            :disabled="submitting"
            @update:model-value="listingId = $event"
            @select="selectedListing = $event"
          />
          <p v-if="selectedListing" class="text-xs text-slate-500">
            Selected: {{ selectedListing.symbol }} (ISIN: {{ selectedListing.isin || "-" }})
          </p>
        </label>

        <label class="block space-y-1 md:col-span-2">
          <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Description</span>
          <BaseInput
            :model-value="description"
            type="text"
            placeholder="Optional"
            :disabled="submitting"
            @update:model-value="description = $event"
          />
        </label>

        <p v-if="errorMessage" class="rounded-md border border-rose-200 bg-rose-50 px-2 py-1 text-xs text-rose-700 md:col-span-2">
          {{ errorMessage }}
        </p>
      </div>

      <footer class="flex justify-end gap-2 border-t border-slate-100 px-5 py-4">
        <BaseButton variant="ghost" :disabled="submitting" @click="emit('close')">Cancel</BaseButton>
        <BaseButton variant="primary" :disabled="!canSubmit" @click="void submit()">
          {{ submitting ? "Creating..." : "Create transaction" }}
        </BaseButton>
      </footer>
    </section>
  </div>
</template>

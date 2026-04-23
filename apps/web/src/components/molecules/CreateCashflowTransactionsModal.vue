<script setup lang="ts">
import { PlusIcon, TrashIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { computed, ref, watch } from "vue";
import { createManualCashflowTransactions } from "../../services/cashflowTransactions";
import { ApiError } from "../../services/http";
import type {
  CreateManualCashflowTransactionInput,
  ManualCashflowTransactionType,
} from "../../types/cashflow";
import BaseButton from "../atoms/BaseButton.vue";
import BaseInput from "../atoms/BaseInput.vue";
import BaseSelect from "../atoms/BaseSelect.vue";
import SingleDatePopover from "./SingleDatePopover.vue";

interface Props {
  open: boolean;
  accountId: string;
}

interface DraftRow {
  id: number;
  date: string;
  amount: string;
  type: ManualCashflowTransactionType;
  description: string;
  note: string;
  tag: string;
  vendor: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  created: [payload: { createdCount: number }];
}>();

const typeOptions = [
  { label: "Outgoing", value: "out" },
  { label: "Incoming", value: "in" },
];

const rows = ref<DraftRow[]>([]);
const submitting = ref(false);
const errorMessage = ref("");
let rowIDCounter = 0;

const decimalPattern = /^\d+(\.\d{1,6})?$/;
const datePattern = /^\d{4}-\d{2}-\d{2}$/;

const canSubmit = computed(() => !submitting.value && rows.value.length > 0 && validateRows() === null);

watch(
  () => props.open,
  (open) => {
    if (!open) {
      return;
    }
    resetState();
  },
);

function todayDate(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function newRow(): DraftRow {
  rowIDCounter += 1;
  return {
    id: rowIDCounter,
    date: todayDate(),
    amount: "",
    type: "out",
    description: "",
    note: "",
    tag: "",
    vendor: "",
  };
}

function resetState(): void {
  rows.value = [newRow()];
  errorMessage.value = "";
  submitting.value = false;
}

function addRow(): void {
  rows.value = [...rows.value, newRow()];
}

function removeRow(id: number): void {
  if (rows.value.length <= 1) {
    return;
  }
  rows.value = rows.value.filter((row) => row.id !== id);
}

function updateRow(id: number, field: keyof Omit<DraftRow, "id">, value: string): void {
  rows.value = rows.value.map((row) => (row.id === id ? { ...row, [field]: value } : row));
}

function validateRows(): string | null {
  if (rows.value.length === 0) {
    return "Add at least one transaction.";
  }
  if (rows.value.length > 100) {
    return "Maximum 100 transactions per submit.";
  }
  for (let index = 0; index < rows.value.length; index += 1) {
    const row = rows.value[index];
    const prefix = `Row ${index + 1}`;
    if (row.date.trim() === "") {
      return `${prefix}: date is required.`;
    }
    if (!datePattern.test(row.date.trim()) || Number.isNaN(new Date(row.date.trim()).getTime())) {
      return `${prefix}: date must use YYYY-MM-DD format.`;
    }
    if (row.amount.trim() === "") {
      return `${prefix}: amount is required.`;
    }
    if (!decimalPattern.test(row.amount.trim())) {
      return `${prefix}: amount must be a positive decimal with up to 6 decimals.`;
    }
    const parsedAmount = Number.parseFloat(row.amount.trim());
    if (Number.isNaN(parsedAmount) || parsedAmount <= 0) {
      return `${prefix}: amount must be greater than 0.`;
    }
    if (row.description.trim() === "") {
      return `${prefix}: description is required.`;
    }
    if (row.note.trim() === "") {
      return `${prefix}: notes are required.`;
    }
    if (row.tag.trim() === "") {
      return `${prefix}: tag is required.`;
    }
  }
  return null;
}

function toPayload(): CreateManualCashflowTransactionInput[] {
  return rows.value.map((row) => ({
    date: row.date.trim(),
    amount: row.amount.trim(),
    type: row.type,
    description: row.description.trim(),
    note: row.note.trim(),
    tag: row.tag.trim(),
    ...(row.vendor.trim() ? { vendor: row.vendor.trim() } : {}),
  }));
}

async function submit(): Promise<void> {
  const validationError = validateRows();
  if (validationError) {
    errorMessage.value = validationError;
    return;
  }

  submitting.value = true;
  errorMessage.value = "";
  try {
    const response = await createManualCashflowTransactions(props.accountId, toPayload());
    emit("created", { createdCount: response.created_count });
    emit("close");
  } catch (error: unknown) {
    if (error instanceof ApiError) {
      errorMessage.value = error.message || "Failed to create transactions.";
    } else if (error instanceof Error) {
      errorMessage.value = error.message;
    } else {
      errorMessage.value = "Failed to create transactions.";
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
    <section class="flex max-h-[84vh] w-full max-w-6xl flex-col rounded-2xl border border-slate-200 bg-white shadow-2xl">
      <header class="flex items-start justify-between border-b border-slate-100 px-5 py-4">
        <div class="space-y-1">
          <h2 class="text-lg font-semibold text-slate-900">Add Cashflow Transactions</h2>
          <p class="text-sm text-slate-500">Create one or more manual transactions for the selected account.</p>
        </div>
        <button
          type="button"
          class="rounded p-1 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
          :disabled="submitting"
          title="Close create cashflow transactions modal"
          @click="emit('close')"
        >
          <XMarkIcon class="h-5 w-5" />
        </button>
      </header>

      <div class="min-h-0 flex-1 overflow-auto px-5 py-4">
        <div class="space-y-3">
          <article
            v-for="(row, index) in rows"
            :key="row.id"
            class="rounded-xl border border-slate-200 bg-slate-50 p-3"
          >
            <div class="mb-2 flex items-center justify-between gap-2">
              <h3 class="text-sm font-semibold text-slate-700">Transaction {{ index + 1 }}</h3>
              <button
                type="button"
                class="inline-flex items-center gap-1 rounded-md border border-slate-200 px-2 py-1 text-xs font-medium text-slate-600 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="submitting || rows.length <= 1"
                @click="removeRow(row.id)"
              >
                <TrashIcon class="h-3.5 w-3.5" />
                Remove
              </button>
            </div>

            <div class="grid grid-cols-1 gap-2 md:grid-cols-2 lg:grid-cols-4">
              <label class="space-y-1">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Date *</span>
                <SingleDatePopover
                  :model-value="row.date"
                  :disabled="submitting"
                  @update:model-value="updateRow(row.id, 'date', $event)"
                />
              </label>

              <label class="space-y-1">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Amount *</span>
                <BaseInput
                  :model-value="row.amount"
                  type="text"
                  placeholder="12.34"
                  :disabled="submitting"
                  @update:model-value="updateRow(row.id, 'amount', $event)"
                />
              </label>

              <label class="space-y-1">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Type *</span>
                <BaseSelect
                  :model-value="row.type"
                  :options="typeOptions"
                  rounded="default"
                  :disabled="submitting"
                  @update:model-value="updateRow(row.id, 'type', $event)"
                />
              </label>

              <label class="space-y-1">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Vendor</span>
                <BaseInput
                  :model-value="row.vendor"
                  type="text"
                  placeholder="Optional (e.g. Cash)"
                  :disabled="submitting"
                  @update:model-value="updateRow(row.id, 'vendor', $event)"
                />
              </label>

              <label class="space-y-1 md:col-span-2">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Description *</span>
                <BaseInput
                  :model-value="row.description"
                  type="text"
                  :disabled="submitting"
                  @update:model-value="updateRow(row.id, 'description', $event)"
                />
              </label>

              <label class="space-y-1 md:col-span-1">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Tag *</span>
                <BaseInput
                  :model-value="row.tag"
                  type="text"
                  :disabled="submitting"
                  @update:model-value="updateRow(row.id, 'tag', $event)"
                />
              </label>

              <label class="space-y-1 md:col-span-1 lg:col-span-1">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Notes *</span>
                <BaseInput
                  :model-value="row.note"
                  type="text"
                  :disabled="submitting"
                  @update:model-value="updateRow(row.id, 'note', $event)"
                />
              </label>
            </div>
          </article>
        </div>
      </div>

      <footer class="flex flex-wrap items-center justify-between gap-2 border-t border-slate-100 px-5 py-4">
        <button
          type="button"
          class="inline-flex items-center gap-1 rounded-md border border-slate-200 px-3 py-2 text-sm font-medium text-slate-700 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
          :disabled="submitting || rows.length >= 100"
          @click="addRow"
        >
          <PlusIcon class="h-4 w-4" />
          Add another
        </button>

        <div class="flex items-center gap-2">
          <BaseButton variant="ghost" :disabled="submitting" @click="emit('close')">Cancel</BaseButton>
          <BaseButton variant="primary" :disabled="!canSubmit" @click="void submit()">
            {{ submitting ? "Creating..." : `Create ${rows.length} transaction${rows.length === 1 ? "" : "s"}` }}
          </BaseButton>
        </div>
      </footer>

      <p
        v-if="errorMessage"
        class="border-t border-rose-100 bg-rose-50 px-5 py-3 text-sm text-rose-700"
      >
        {{ errorMessage }}
      </p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { PencilSquareIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { computed, ref, watch } from "vue";
import type { AssetChangeDirection, AssetClassDetails } from "../../types/assets";
import AssetGrowthLineChart from "../molecules/charts/AssetGrowthLineChart.vue";
import BasePopover from "../atoms/BasePopover.vue";
import BaseButton from "../atoms/BaseButton.vue";
import BaseInput from "../atoms/BaseInput.vue";
import BaseSelect from "../atoms/BaseSelect.vue";
import IconButton from "../atoms/IconButton.vue";
import SingleDatePopover from "../molecules/SingleDatePopover.vue";
import UnrealizedPnLBadge from "../molecules/UnrealizedPnLBadge.vue";

interface Props {
  open: boolean;
  details: AssetClassDetails | null;
  loading?: boolean;
  busy?: boolean;
  errorMessage?: string;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  busy: false,
  errorMessage: "",
});
const HISTORY_PAGE_SIZE = 8;

const emit = defineEmits<{
  close: [];
  "create-item": [payload: { name: string; initialWorth: string; effectiveDate: string; note?: string }];
  "set-worth": [payload: { itemId: string; worth: string; effectiveDate: string; note?: string }];
  "adjust-worth": [payload: { itemId: string; direction: "increase" | "decrease"; amount: string; effectiveDate: string; note?: string }];
}>();

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  minimumFractionDigits: 2,
  maximumFractionDigits: 2,
});

const directionOptions = [
  { label: "Increase", value: "increase" },
  { label: "Decrease", value: "decrease" },
];

const todayIso = computed(() => {
  const now = new Date();
  const year = now.getUTCFullYear();
  const month = String(now.getUTCMonth() + 1).padStart(2, "0");
  const day = String(now.getUTCDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
});

const createItemOpen = ref(false);
const itemName = ref("");
const itemInitialWorth = ref("");
const itemEffectiveDate = ref("");
const itemNote = ref("");
const itemError = ref("");

const worthFormOpen = ref(false);
const worthMode = ref<"set" | "adjust">("set");
const worthItemId = ref("");
const worthValue = ref("");
const worthAmount = ref("");
const worthDirection = ref<"increase" | "decrease">("increase");
const worthEffectiveDate = ref("");
const worthNote = ref("");
const worthError = ref("");
const historyPage = ref(1);

const canManageClass = computed(() => props.details?.classRow.source === "MANUAL");

watch(
  () => [props.open, props.details?.classRow.id] as const,
  () => {
    if (!props.open || !props.details) {
      return;
    }
    resetItemForm();
    resetWorthForm();
    historyPage.value = 1;
  },
  { immediate: true },
);

watch(
  () => props.open,
  (open) => {
    if (open) {
      return;
    }
    resetItemForm();
    resetWorthForm();
    historyPage.value = 1;
  },
);

const worthItem = computed(() => {
  if (!props.details) {
    return null;
  }
  return props.details.items.find((item) => item.id === worthItemId.value) ?? null;
});

const canSubmitItem = computed(() => {
  return (
    !props.busy &&
    itemName.value.trim() !== "" &&
    itemInitialWorth.value.trim() !== "" &&
    itemEffectiveDate.value.trim() !== ""
  );
});

const canSubmitWorth = computed(() => {
  if (props.busy || worthItemId.value.trim() === "" || worthEffectiveDate.value.trim() === "") {
    return false;
  }
  if (worthMode.value === "set") {
    return worthValue.value.trim() !== "";
  }
  return worthAmount.value.trim() !== "";
});

const totalHistoryPages = computed(() => {
  const historyCount = props.details?.history.length ?? 0;
  return Math.max(1, Math.ceil(historyCount / HISTORY_PAGE_SIZE));
});

const pagedHistory = computed(() => {
  if (!props.details) {
    return [];
  }
  const start = (historyPage.value - 1) * HISTORY_PAGE_SIZE;
  return props.details.history.slice(start, start + HISTORY_PAGE_SIZE);
});

const hasHistoryPagination = computed(() => {
  const historyCount = props.details?.history.length ?? 0;
  return historyCount > HISTORY_PAGE_SIZE;
});

function goToPreviousHistoryPage(): void {
  historyPage.value = Math.max(1, historyPage.value - 1);
}

function goToNextHistoryPage(): void {
  historyPage.value = Math.min(totalHistoryPages.value, historyPage.value + 1);
}

function formatWorth(value: string): string {
  const amount = Number.parseFloat(value);
  if (Number.isNaN(amount)) {
    return value;
  }
  return currencyFormatter.format(amount);
}

function formatDate(value?: string): string {
  if (!value) {
    return "n/a";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleDateString();
}

function historyLabel(entry: { changeType: string; direction?: AssetChangeDirection }): string {
  if (entry.changeType === "SET") {
    return "Set worth";
  }
  if (entry.direction === "DECREASE") {
    return "Adjusted down";
  }
  return "Adjusted up";
}

function openWorthForm(mode: "set" | "adjust", itemId: string): void {
  worthMode.value = mode;
  worthItemId.value = itemId;
  worthFormOpen.value = true;
  worthError.value = "";
  worthValue.value = "";
  worthAmount.value = "";
  worthDirection.value = "increase";
  worthEffectiveDate.value = todayIso.value;
  worthNote.value = "";
}

function closeWorthForm(): void {
  worthFormOpen.value = false;
  worthError.value = "";
}

function resetItemForm(): void {
  createItemOpen.value = false;
  itemName.value = "";
  itemInitialWorth.value = "";
  itemEffectiveDate.value = todayIso.value;
  itemNote.value = "";
  itemError.value = "";
}

function resetWorthForm(): void {
  worthFormOpen.value = false;
  worthMode.value = "set";
  worthItemId.value = "";
  worthValue.value = "";
  worthAmount.value = "";
  worthDirection.value = "increase";
  worthEffectiveDate.value = todayIso.value;
  worthNote.value = "";
  worthError.value = "";
}

function submitCreateItem(): void {
  if (itemName.value.trim() === "") {
    itemError.value = "Item name is required.";
    return;
  }
  if (itemInitialWorth.value.trim() === "") {
    itemError.value = "Initial worth is required.";
    return;
  }
  if (itemEffectiveDate.value.trim() === "") {
    itemError.value = "Effective date is required.";
    return;
  }
  itemError.value = "";
  emit("create-item", {
    name: itemName.value.trim(),
    initialWorth: itemInitialWorth.value.trim(),
    effectiveDate: itemEffectiveDate.value.trim(),
    note: itemNote.value.trim() || undefined,
  });
}

function submitWorthChange(): void {
  if (worthItemId.value.trim() === "") {
    worthError.value = "Select an item first.";
    return;
  }
  if (worthEffectiveDate.value.trim() === "") {
    worthError.value = "Effective date is required.";
    return;
  }
  if (worthMode.value === "set") {
    if (worthValue.value.trim() === "") {
      worthError.value = "Worth is required.";
      return;
    }
    worthError.value = "";
    emit("set-worth", {
      itemId: worthItemId.value,
      worth: worthValue.value.trim(),
      effectiveDate: worthEffectiveDate.value.trim(),
      note: worthNote.value.trim() || undefined,
    });
    return;
  }

  if (worthAmount.value.trim() === "") {
    worthError.value = "Amount is required.";
    return;
  }
  worthError.value = "";
  emit("adjust-worth", {
    itemId: worthItemId.value,
    direction: worthDirection.value,
    amount: worthAmount.value.trim(),
    effectiveDate: worthEffectiveDate.value.trim(),
    note: worthNote.value.trim() || undefined,
  });
}

</script>

<template>
  <div
    class="fixed inset-0 z-50 flex justify-end bg-slate-900/40 transition-opacity"
    :class="open ? 'pointer-events-auto opacity-100' : 'pointer-events-none opacity-0'"
    @click.self="emit('close')"
  >
    <aside
      class="flex h-full w-full max-w-2xl flex-col border-l border-slate-300 bg-white shadow-2xl transition-transform duration-200"
      :class="open ? 'translate-x-0' : 'translate-x-full'"
    >
      <header class="flex items-start justify-between border-b border-slate-200 px-5 py-4">
        <div v-if="details" class="space-y-1">
          <h2 class="text-lg font-semibold text-slate-900">{{ details.classRow.name }}</h2>
          <p class="text-sm text-slate-500">
            {{ details.classRow.source === "PORTFOLIO" ? "Portfolio-linked class" : "Manual class" }}
          </p>
        </div>
        <button
          type="button"
          class="rounded p-1 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
          title="Close asset class panel"
          @click="emit('close')"
        >
          <XMarkIcon class="h-5 w-5" />
        </button>
      </header>

      <div class="min-h-0 flex-1 overflow-y-auto px-5 py-4">
        <p v-if="errorMessage" class="mb-4 rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {{ errorMessage }}
        </p>

        <div v-if="loading" class="space-y-3">
          <div class="h-24 animate-pulse rounded-2xl bg-slate-100" />
          <div class="h-56 animate-pulse rounded-2xl bg-slate-100" />
          <div class="h-64 animate-pulse rounded-2xl bg-slate-100" />
        </div>

        <div v-else-if="details" class="space-y-5">
          <section class="grid grid-cols-1 gap-3 rounded-2xl border border-slate-300 bg-white p-3 sm:grid-cols-3">
            <article>
              <p class="text-xs uppercase tracking-wide text-slate-500">Current worth</p>
              <p class="text-lg font-semibold text-slate-900">{{ formatWorth(details.classRow.currentWorth) }}</p>
            </article>
            <article>
              <p class="text-xs uppercase tracking-wide text-slate-500">Growth (Inception)</p>
              <div class="pt-1">
                <UnrealizedPnLBadge :value="details.classRow.growthPct" />
              </div>
            </article>
            <article>
              <p class="text-xs uppercase tracking-wide text-slate-500">Last change</p>
              <p class="text-sm font-medium text-slate-700">{{ formatDate(details.classRow.lastChangeAt) }}</p>
            </article>
          </section>

          <section class="space-y-2 rounded-2xl border border-slate-300 bg-white p-3">
            <div class="h-56">
              <AssetGrowthLineChart :loading="loading" :data="details.growth" series-label="Class Worth" />
            </div>
          </section>

          <section class="space-y-3 rounded-2xl border border-slate-300 bg-white p-3">
            <div v-if="canManageClass" class="flex items-center justify-end">
              <BaseButton
                variant="primary"
                size="sm"
                :disabled="busy"
                @click="createItemOpen = !createItemOpen"
              >
                {{ createItemOpen ? "Close" : "Add asset" }}
              </BaseButton>
            </div>

            <div v-if="createItemOpen" class="space-y-2 rounded-xl border border-slate-300 bg-slate-50 p-3">
              <label class="block space-y-1">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Name *</span>
                <BaseInput :model-value="itemName" placeholder="Savings account A" @update:model-value="itemName = $event" />
              </label>
              <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
                <label class="block space-y-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Initial worth *</span>
                  <BaseInput :model-value="itemInitialWorth" type="text" placeholder="10000" @update:model-value="itemInitialWorth = $event" />
                </label>
                <label class="block space-y-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Effective date *</span>
                  <SingleDatePopover
                    :model-value="itemEffectiveDate"
                    :max-date="todayIso"
                    @update:model-value="itemEffectiveDate = $event"
                  />
                </label>
              </div>
              <label class="block space-y-1">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Notes</span>
                <BaseInput :model-value="itemNote" type="text" placeholder="Optional note" @update:model-value="itemNote = $event" />
              </label>
              <p v-if="itemError" class="text-xs text-rose-700">{{ itemError }}</p>
              <div class="flex items-center gap-2">
                <BaseButton variant="primary" size="sm" :disabled="!canSubmitItem" @click="submitCreateItem">
                  {{ busy ? "Saving..." : "Create asset" }}
                </BaseButton>
                <BaseButton variant="ghost" size="sm" :disabled="busy" @click="resetItemForm">Cancel</BaseButton>
              </div>
            </div>

            <div class="overflow-auto bg-slate-100">
              <table class="w-full min-w-[760px] border-separate border-spacing-0 bg-white text-sm">
                <thead class="sticky top-0 z-20 bg-white/95 text-left backdrop-blur">
                  <tr>
                    <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Name</th>
                    <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Worth</th>
                    <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Updated</th>
                    <th class="border-b border-slate-200 px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-slate-500">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in details.items" :key="item.id" class="transition hover:bg-slate-50">
                    <td class="border-b border-slate-100 px-3 py-2 font-medium text-slate-800">{{ item.name }}</td>
                    <td class="border-b border-slate-100 px-3 py-2 text-slate-700">{{ formatWorth(item.currentWorth) }}</td>
                    <td class="border-b border-slate-100 px-3 py-2 text-slate-600">{{ formatDate(item.updatedAt) }}</td>
                    <td class="border-b border-slate-100 px-3 py-2">
                      <BasePopover
                        v-if="canManageClass"
                        align="right"
                        offset-class="mt-1"
                        panel-class="w-40"
                      >
                        <template #trigger="{ toggle }">
                          <IconButton
                            tone="neutral"
                            size="sm"
                            title="Edit asset"
                            :disabled="busy"
                            @click.stop="toggle"
                          >
                            <PencilSquareIcon class="h-4 w-4" />
                          </IconButton>
                        </template>
                        <template #default="{ close }">
                          <div class="translate-y-1 rounded-xl border border-slate-200 bg-white p-1 shadow-lg">
                            <button
                              type="button"
                              class="flex w-full items-center rounded-lg px-2 py-1.5 text-left text-xs font-medium text-slate-700 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                              :disabled="busy"
                              @click="openWorthForm('set', item.id); close()"
                            >
                              Set worth
                            </button>
                            <button
                              type="button"
                              class="mt-1 flex w-full items-center rounded-lg px-2 py-1.5 text-left text-xs font-medium text-slate-700 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                              :disabled="busy"
                              @click="openWorthForm('adjust', item.id); close()"
                            >
                              Adjust worth
                            </button>
                          </div>
                        </template>
                      </BasePopover>
                    </td>
                  </tr>
                  <tr v-if="details.items.length === 0">
                    <td colspan="4" class="px-3 py-6 text-center text-sm text-slate-500">No assets in this class yet.</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-if="worthFormOpen" class="space-y-2 rounded-xl border border-slate-300 bg-slate-50 p-3">
              <p v-if="worthItem" class="text-xs font-medium text-slate-500">
                Editing {{ worthItem.name }}
              </p>
              <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
                <label class="block space-y-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Item *</span>
                  <BaseSelect
                    rounded="default"
                    :model-value="worthItemId"
                    :options="details.items.map((item) => ({ label: item.name, value: item.id }))"
                    @update:model-value="worthItemId = $event"
                  />
                </label>
                <label class="block space-y-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Effective date *</span>
                  <SingleDatePopover
                    :model-value="worthEffectiveDate"
                    :max-date="todayIso"
                    @update:model-value="worthEffectiveDate = $event"
                  />
                </label>
              </div>

              <div v-if="worthMode === 'set'" class="grid grid-cols-1 gap-2 sm:grid-cols-2">
                <label class="block space-y-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Worth *</span>
                  <BaseInput :model-value="worthValue" placeholder="125000" @update:model-value="worthValue = $event" />
                </label>
              </div>

              <div v-else class="grid grid-cols-1 gap-2 sm:grid-cols-2">
                <label class="block space-y-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Direction *</span>
                  <BaseSelect
                    rounded="default"
                    :model-value="worthDirection"
                    :options="directionOptions"
                    @update:model-value="worthDirection = $event as 'increase' | 'decrease'"
                  />
                </label>
                <label class="block space-y-1">
                  <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Amount *</span>
                  <BaseInput :model-value="worthAmount" placeholder="5000" @update:model-value="worthAmount = $event" />
                </label>
              </div>

              <label class="block space-y-1">
                <span class="text-xs font-semibold uppercase tracking-wide text-slate-500">Notes</span>
                <BaseInput :model-value="worthNote" placeholder="Optional note" @update:model-value="worthNote = $event" />
              </label>
              <p v-if="worthError" class="text-xs text-rose-700">{{ worthError }}</p>
              <div class="flex items-center gap-2">
                <BaseButton variant="primary" size="sm" :disabled="!canSubmitWorth" @click="submitWorthChange">
                  {{ busy ? "Saving..." : "Save change" }}
                </BaseButton>
                <BaseButton variant="ghost" size="sm" :disabled="busy" @click="closeWorthForm">Cancel</BaseButton>
              </div>
            </div>
          </section>

          <section class="space-y-2 rounded-2xl border border-slate-300 bg-white p-3">
            <ul class="space-y-2">
              <li
                v-for="entry in pagedHistory"
                :key="entry.id"
                class="rounded-xl border border-slate-300 bg-slate-50 px-3 py-2"
              >
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <p class="text-sm font-medium text-slate-800">{{ historyLabel(entry) }}</p>
                  <p class="text-xs text-slate-500">{{ formatDate(entry.effectiveDate) }}</p>
                </div>
                <p class="text-sm text-slate-700">
                  {{ formatWorth(entry.previousWorth) }} -> {{ formatWorth(entry.newWorth) }}
                  (class: {{ formatWorth(entry.classTotalWorth) }})
                </p>
                <p v-if="entry.note" class="text-xs text-slate-600">{{ entry.note }}</p>
              </li>
              <li v-if="details.history.length === 0" class="text-sm text-slate-500">No history yet.</li>
            </ul>
            <div
              v-if="hasHistoryPagination"
              class="flex items-center justify-between border-t border-slate-100 pt-2"
            >
              <BaseButton
                variant="secondary"
                size="sm"
                :disabled="historyPage <= 1"
                @click="goToPreviousHistoryPage"
              >
                Previous
              </BaseButton>
              <span class="text-xs font-medium text-slate-600">
                Page {{ historyPage }} / {{ totalHistoryPages }}
              </span>
              <BaseButton
                variant="secondary"
                size="sm"
                :disabled="historyPage >= totalHistoryPages"
                @click="goToNextHistoryPage"
              >
                Next
              </BaseButton>
            </div>
          </section>
        </div>
      </div>
    </aside>
  </div>
</template>

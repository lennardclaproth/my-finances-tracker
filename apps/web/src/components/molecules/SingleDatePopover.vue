<script setup lang="ts">
import { CalendarDaysIcon, ChevronLeftIcon, ChevronRightIcon } from "@heroicons/vue/24/outline";
import { computed, ref } from "vue";
import BaseButton from "../atoms/BaseButton.vue";
import BasePopover from "../atoms/BasePopover.vue";
import IconButton from "../atoms/IconButton.vue";

interface Props {
  modelValue: string;
  disabled?: boolean;
  minDate?: string;
  maxDate?: string;
  placeholder?: string;
}

interface CalendarDay {
  date: Date;
  inCurrentMonth: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  minDate: "",
  maxDate: "",
  placeholder: "Select date",
});

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const weekLabels = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const monthLabelFormatter = new Intl.DateTimeFormat(undefined, {
  month: "long",
  year: "numeric",
});
const dateLabelFormatter = new Intl.DateTimeFormat(undefined, {
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
});

const isOpen = ref(false);
const visibleMonth = ref(startOfMonth(new Date()));
const draftDate = ref<Date | null>(null);

const activeDate = computed(() => (isOpen.value ? draftDate.value : parseISODate(props.modelValue)));
const displayLabel = computed(() => {
  if (activeDate.value) {
    return dateLabelFormatter.format(activeDate.value);
  }
  return props.placeholder;
});

const days = computed(() => buildCalendarGrid(visibleMonth.value));
const minDateValue = computed(() => parseISODate(props.minDate));
const maxDateValue = computed(() => parseISODate(props.maxDate));

function toISODate(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function parseISODate(value: string): Date | null {
  if (!value) {
    return null;
  }
  const parsed = new Date(`${value}T00:00:00`);
  if (Number.isNaN(parsed.getTime())) {
    return null;
  }
  parsed.setHours(0, 0, 0, 0);
  return parsed;
}

function startOfMonth(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

function addMonths(date: Date, months: number): Date {
  return new Date(date.getFullYear(), date.getMonth() + months, 1);
}

function startOfWeekMonday(date: Date): Date {
  const day = date.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  return new Date(date.getFullYear(), date.getMonth(), date.getDate() + diff);
}

function buildCalendarGrid(monthStart: Date): CalendarDay[] {
  const firstDay = new Date(monthStart.getFullYear(), monthStart.getMonth(), 1);
  const gridStart = startOfWeekMonday(firstDay);
  const output: CalendarDay[] = [];
  for (let index = 0; index < 42; index += 1) {
    const day = new Date(gridStart.getFullYear(), gridStart.getMonth(), gridStart.getDate() + index);
    output.push({
      date: day,
      inCurrentMonth: day.getMonth() === monthStart.getMonth(),
    });
  }
  return output;
}

function compareDate(a: Date, b: Date): number {
  const aValue = new Date(a.getFullYear(), a.getMonth(), a.getDate()).getTime();
  const bValue = new Date(b.getFullYear(), b.getMonth(), b.getDate()).getTime();
  return aValue - bValue;
}

function isSameDate(a: Date | null, b: Date): boolean {
  if (!a) {
    return false;
  }
  return compareDate(a, b) === 0;
}

function isDisabledDay(day: Date): boolean {
  if (minDateValue.value && compareDate(day, minDateValue.value) < 0) {
    return true;
  }
  if (maxDateValue.value && compareDate(day, maxDateValue.value) > 0) {
    return true;
  }
  return false;
}

function dayCellClass(day: CalendarDay): string {
  if (isSameDate(draftDate.value, day.date)) {
    return "bg-indigo-600 text-white";
  }
  if (isDisabledDay(day.date)) {
    return "cursor-not-allowed text-slate-300";
  }
  if (!day.inCurrentMonth) {
    return "text-slate-300";
  }
  return "text-slate-700 hover:bg-slate-100";
}

function initializeDraft(): void {
  draftDate.value = parseISODate(props.modelValue);
  visibleMonth.value = startOfMonth(draftDate.value ?? new Date());
}

function onTriggerClick(open: () => void, close: () => void): void {
  if (isOpen.value) {
    close();
    return;
  }
  initializeDraft();
  open();
}

function selectDay(day: Date): void {
  if (isDisabledDay(day)) {
    return;
  }
  draftDate.value = new Date(day.getFullYear(), day.getMonth(), day.getDate());
}

function apply(close: () => void): void {
  emit("update:modelValue", draftDate.value ? toISODate(draftDate.value) : "");
  close();
}

function clearDate(close: () => void): void {
  draftDate.value = null;
  emit("update:modelValue", "");
  close();
}

function prevMonth(): void {
  visibleMonth.value = addMonths(visibleMonth.value, -1);
}

function nextMonth(): void {
  visibleMonth.value = addMonths(visibleMonth.value, 1);
}
</script>

<template>
  <BasePopover
    v-model:open="isOpen"
    :disabled="disabled"
    panel-class="w-[20rem] max-w-[calc(100vw-2rem)] rounded-xl border border-slate-300 bg-white p-3 shadow-2xl"
    z-index-class="z-40"
  >
    <template #trigger="{ open, close }">
      <BaseButton
        unstyled
        class="inline-flex w-full items-center justify-between gap-2 rounded-md border border-slate-300 bg-white/90 px-3 py-2 text-sm text-slate-700 shadow-sm transition hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-300 disabled:cursor-not-allowed disabled:opacity-60"
        :disabled="disabled"
        :title="displayLabel"
        @click="onTriggerClick(open, close)"
      >
        <span class="inline-flex min-w-0 items-center gap-2">
          <CalendarDaysIcon class="h-5 w-5 text-slate-500" />
          <span class="truncate">{{ displayLabel }}</span>
        </span>
      </BaseButton>
    </template>

    <template #default="{ close }">
      <div class="mb-2 flex items-center justify-between">
        <IconButton
          unstyled
          class="rounded-full border border-slate-300 p-1 text-slate-600 hover:bg-slate-100"
          title="Previous month"
          @click="prevMonth"
        >
          <ChevronLeftIcon class="h-4 w-4" />
        </IconButton>
        <p class="text-sm font-semibold text-slate-800">{{ monthLabelFormatter.format(visibleMonth) }}</p>
        <IconButton
          unstyled
          class="rounded-full border border-slate-300 p-1 text-slate-600 hover:bg-slate-100"
          title="Next month"
          @click="nextMonth"
        >
          <ChevronRightIcon class="h-4 w-4" />
        </IconButton>
      </div>

      <div class="mb-3 grid grid-cols-7 text-center text-xs font-semibold uppercase text-slate-400">
        <span v-for="label in weekLabels" :key="label">{{ label }}</span>
      </div>
      <div class="grid grid-cols-7 gap-1">
        <BaseButton
          unstyled
          v-for="day in days"
          :key="toISODate(day.date)"
          class="h-8 rounded-md text-sm transition"
          :class="dayCellClass(day)"
          :disabled="isDisabledDay(day.date)"
          @click="selectDay(day.date)"
        >
          {{ day.date.getDate() }}
        </BaseButton>
      </div>

      <div class="mt-3 flex justify-end gap-2 border-t border-slate-100 pt-3">
        <BaseButton variant="ghost" size="sm" @click="clearDate(close)">Clear</BaseButton>
        <BaseButton variant="primary" size="sm" :disabled="!draftDate" @click="apply(close)">Apply</BaseButton>
      </div>
    </template>
  </BasePopover>
</template>

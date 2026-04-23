<script setup lang="ts">
import {
  ArrowRightIcon,
  CalendarDaysIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
} from "@heroicons/vue/24/outline";
import { computed, ref } from "vue";
import BaseButton from "../atoms/BaseButton.vue";
import BasePopover from "../atoms/BasePopover.vue";
import IconButton from "../atoms/IconButton.vue";

interface Props {
  from: string;
  to: string;
  disabled?: boolean;
}

interface CalendarDay {
  date: Date;
  inCurrentMonth: boolean;
}

type PresetKey = "YTD" | "1W" | "1M" | "3M" | "1Y" | "3Y";

interface PresetOption {
  key: PresetKey;
  label: string;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
});

const emit = defineEmits<{
  apply: [from: string, to: string];
  clear: [];
}>();

const weekLabels = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const presetOptions: PresetOption[] = [
  { key: "YTD", label: "YTD" },
  { key: "1W", label: "1W" },
  { key: "1M", label: "1M" },
  { key: "3M", label: "3M" },
  { key: "1Y", label: "1Y" },
  { key: "3Y", label: "3Y" },
];
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
const draftFrom = ref<Date | null>(null);
const draftTo = ref<Date | null>(null);
const selectedPreset = ref<PresetKey | null>(null);

const leftMonth = computed(() => visibleMonth.value);
const rightMonth = computed(() => addMonths(visibleMonth.value, 1));
const leftDays = computed(() => buildCalendarGrid(leftMonth.value));
const rightDays = computed(() => buildCalendarGrid(rightMonth.value));
const activeFrom = computed(() => (isOpen.value ? draftFrom.value : parseISODate(props.from)));
const activeTo = computed(() => (isOpen.value ? draftTo.value : parseISODate(props.to)));
const activePreset = computed<PresetKey | null>(() => {
  if (isOpen.value) {
    return selectedPreset.value ?? detectPreset(activeFrom.value, activeTo.value);
  }
  return detectPreset(activeFrom.value, activeTo.value);
});

const displayLabel = computed(() => {
  if (activePreset.value) {
    return activePreset.value;
  }
  if (activeFrom.value && activeTo.value) {
    return `${dateLabelFormatter.format(activeFrom.value)} -> ${dateLabelFormatter.format(activeTo.value)}`;
  }
  if (activeFrom.value) {
    return `${dateLabelFormatter.format(activeFrom.value)} -> ...`;
  }
  return "Select date range";
});

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

function todayAtStartOfDay(): Date {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  return today;
}

function dateOnly(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
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

  const days: CalendarDay[] = [];
  for (let index = 0; index < 42; index += 1) {
    const day = new Date(gridStart.getFullYear(), gridStart.getMonth(), gridStart.getDate() + index);
    days.push({
      date: day,
      inCurrentMonth: day.getMonth() === monthStart.getMonth(),
    });
  }

  return days;
}

function compareDate(a: Date, b: Date): number {
  const aValue = new Date(a.getFullYear(), a.getMonth(), a.getDate()).getTime();
  const bValue = new Date(b.getFullYear(), b.getMonth(), b.getDate()).getTime();
  return aValue - bValue;
}

function rangeForPreset(preset: PresetKey, today: Date): { from: Date; to: Date } {
  const to = dateOnly(today);
  switch (preset) {
    case "YTD":
      return {
        from: new Date(to.getFullYear(), 0, 1),
        to,
      };
    case "1W":
      return {
        from: new Date(to.getFullYear(), to.getMonth(), to.getDate() - 6),
        to,
      };
    case "1M":
      return {
        from: new Date(to.getFullYear(), to.getMonth() - 1, to.getDate()),
        to,
      };
    case "3M":
      return {
        from: new Date(to.getFullYear(), to.getMonth() - 3, to.getDate()),
        to,
      };
    case "1Y":
      return {
        from: new Date(to.getFullYear() - 1, to.getMonth(), to.getDate()),
        to,
      };
    case "3Y":
      return {
        from: new Date(to.getFullYear() - 3, to.getMonth(), to.getDate()),
        to,
      };
  }
}

function detectPreset(from: Date | null, to: Date | null): PresetKey | null {
  if (!from || !to) {
    return null;
  }
  const today = todayAtStartOfDay();
  for (const preset of presetOptions) {
    const range = rangeForPreset(preset.key, today);
    if (compareDate(from, range.from) === 0 && compareDate(to, range.to) === 0) {
      return preset.key;
    }
  }
  return null;
}

function isSameDate(a: Date | null, b: Date): boolean {
  if (!a) {
    return false;
  }
  return compareDate(a, b) === 0;
}

function isInRange(day: Date): boolean {
  if (!draftFrom.value || !draftTo.value) {
    return false;
  }
  return compareDate(day, draftFrom.value) >= 0 && compareDate(day, draftTo.value) <= 0;
}

function dayCellClass(day: CalendarDay): string {
  const selected = isSameDate(draftFrom.value, day.date) || isSameDate(draftTo.value, day.date);
  const inRange = !selected && isInRange(day.date);

  if (selected) {
    return "bg-indigo-600 text-white";
  }
  if (inRange) {
    return "bg-indigo-100 text-indigo-800";
  }
  if (!day.inCurrentMonth) {
    return "text-slate-300";
  }
  return "text-slate-700 hover:bg-slate-100";
}

function selectDay(day: Date): void {
  const normalized = new Date(day.getFullYear(), day.getMonth(), day.getDate());
  selectedPreset.value = null;

  if (!draftFrom.value || (draftFrom.value && draftTo.value)) {
    draftFrom.value = normalized;
    draftTo.value = null;
    return;
  }

  if (compareDate(normalized, draftFrom.value) < 0) {
    draftTo.value = draftFrom.value;
    draftFrom.value = normalized;
    return;
  }

  draftTo.value = normalized;
}

function initializeDraftRange(): void {
  draftFrom.value = parseISODate(props.from);
  draftTo.value = parseISODate(props.to);
  selectedPreset.value = detectPreset(draftFrom.value, draftTo.value);
  visibleMonth.value = startOfMonth(draftFrom.value ?? new Date());
}

function selectPreset(preset: PresetKey): void {
  const range = rangeForPreset(preset, todayAtStartOfDay());
  draftFrom.value = range.from;
  draftTo.value = range.to;
  selectedPreset.value = preset;
  visibleMonth.value = startOfMonth(range.from);
}

function onTriggerClick(open: () => void, close: () => void): void {
  if (isOpen.value) {
    close();
    return;
  }
  initializeDraftRange();
  open();
}

function apply(close: () => void): void {
  selectedPreset.value = detectPreset(draftFrom.value, draftTo.value);
  emit("apply", draftFrom.value ? toISODate(draftFrom.value) : "", draftTo.value ? toISODate(draftTo.value) : "");
  close();
}

function clearDates(close: () => void): void {
  draftFrom.value = null;
  draftTo.value = null;
  selectedPreset.value = null;
  emit("clear");
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
    panel-class="w-[44rem] max-w-[calc(100vw-2rem)] rounded-xl border border-slate-300 bg-white p-4 shadow-2xl"
    z-index-class="z-40"
  >
    <template #trigger="{ open, close }">
      <BaseButton
        unstyled
        class="inline-flex items-center gap-2 rounded-full border border-slate-300 bg-white/90 px-4 py-2 text-sm text-slate-700 shadow-sm transition hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-300 disabled:cursor-not-allowed disabled:opacity-60"
        :disabled="disabled"
        :title="displayLabel"
        @click="onTriggerClick(open, close)"
      >
        <CalendarDaysIcon class="h-5 w-5 text-slate-500" />
        <span class="whitespace-nowrap">{{ displayLabel }}</span>
        <ArrowRightIcon class="h-4 w-4 text-slate-400" />
      </BaseButton>
    </template>

    <template #default="{ close }">
      <div class="mb-3 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <IconButton
            unstyled
            class="rounded-full border border-slate-300 p-1 text-slate-600 hover:bg-slate-100"
            title="Previous month"
            @click="prevMonth"
          >
            <ChevronLeftIcon class="h-4 w-4" />
          </IconButton>
          <IconButton
            unstyled
            class="rounded-full border border-slate-300 p-1 text-slate-600 hover:bg-slate-100"
            title="Next month"
            @click="nextMonth"
          >
            <ChevronRightIcon class="h-4 w-4" />
          </IconButton>
        </div>

        <div class="text-sm text-slate-600">Click start and end dates to select a range</div>
      </div>

      <div class="mb-4 flex flex-wrap items-center gap-2">
        <BaseButton
          unstyled
          v-for="preset in presetOptions"
          :key="preset.key"
          class="rounded-full border px-3 py-1 text-xs font-semibold uppercase tracking-wide transition"
          :class="activePreset === preset.key
            ? 'border-indigo-500 bg-indigo-50 text-indigo-700'
            : 'border-slate-300 bg-white text-slate-600 hover:bg-slate-100'"
          @click="selectPreset(preset.key)"
        >
          {{ preset.label }}
        </BaseButton>
      </div>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <article class="rounded-lg border border-slate-300 p-3">
          <h3 class="mb-2 text-sm font-semibold text-slate-800">{{ monthLabelFormatter.format(leftMonth) }}</h3>
          <div class="mb-1 grid grid-cols-7 text-center text-xs font-semibold uppercase text-slate-400">
            <span v-for="label in weekLabels" :key="`left-${label}`">{{ label }}</span>
          </div>
          <div class="grid grid-cols-7 gap-1">
            <BaseButton
              unstyled
              v-for="day in leftDays"
              :key="`left-${toISODate(day.date)}`"
              class="h-8 rounded-md text-sm transition"
              :class="dayCellClass(day)"
              @click="selectDay(day.date)"
            >
              {{ day.date.getDate() }}
            </BaseButton>
          </div>
        </article>

        <article class="rounded-lg border border-slate-300 p-3">
          <h3 class="mb-2 text-sm font-semibold text-slate-800">{{ monthLabelFormatter.format(rightMonth) }}</h3>
          <div class="mb-1 grid grid-cols-7 text-center text-xs font-semibold uppercase text-slate-400">
            <span v-for="label in weekLabels" :key="`right-${label}`">{{ label }}</span>
          </div>
          <div class="grid grid-cols-7 gap-1">
            <BaseButton
              unstyled
              v-for="day in rightDays"
              :key="`right-${toISODate(day.date)}`"
              class="h-8 rounded-md text-sm transition"
              :class="dayCellClass(day)"
              @click="selectDay(day.date)"
            >
              {{ day.date.getDate() }}
            </BaseButton>
          </div>
        </article>
      </div>

      <div class="mt-4 flex justify-end gap-2">
        <BaseButton variant="ghost" @click="clearDates(close)">Clear</BaseButton>
        <BaseButton variant="primary" @click="apply(close)">Apply</BaseButton>
      </div>
    </template>
  </BasePopover>
</template>

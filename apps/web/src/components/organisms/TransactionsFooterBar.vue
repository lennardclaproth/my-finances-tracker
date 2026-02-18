<script setup lang="ts">
import {
  BackwardIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  CheckCircleIcon,
  EyeIcon,
  ForwardIcon,
  NoSymbolIcon,
  TagIcon,
  XMarkIcon,
} from "@heroicons/vue/24/outline";
import { computed } from "vue";
import BaseButton from "../atoms/BaseButton.vue";
import BaseSelect from "../atoms/BaseSelect.vue";
import IconButton from "../atoms/IconButton.vue";

interface Props {
  limit: number;
  offset: number;
  count: number;
  total: number;
  selectedCount: number;
  allMatchingSelected: boolean;
  actionsEnabled: boolean;
  loading?: boolean;
  pageSizeOptions?: number[];
}

const props = withDefaults(defineProps<Props>(), {
  actionsEnabled: false,
  loading: false,
  pageSizeOptions: () => [10, 25, 50, 100],
});

const emit = defineEmits<{
  "change-limit": [value: number];
  "go-first": [];
  "go-prev": [];
  "go-next": [];
  "go-last": [];
  "open-tag": [];
  ignore: [];
  unignore: [];
  "clear-selection": [];
  "toggle-all-matching": [];
}>();

const currentPage = computed(() => Math.floor(props.offset / props.limit) + 1);
const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.limit)));
const rangeStart = computed(() => (props.total === 0 ? 0 : props.offset + 1));
const rangeEnd = computed(() => (props.total === 0 ? 0 : props.offset + props.count));
const hasPrevious = computed(() => props.offset > 0);
const hasNext = computed(() => props.offset + props.limit < props.total);

const pageSizeSelectOptions = computed(() =>
  props.pageSizeOptions.map((option) => ({
    label: `${option} / page`,
    value: option,
  })),
);

function onLimitChange(rawValue: string): void {
  const nextLimit = Number.parseInt(rawValue, 10);
  if (!Number.isNaN(nextLimit)) {
    emit("change-limit", nextLimit);
  }
}
</script>

<template>
  <div class="bg-transparent pt-2">
    <div class="grid grid-cols-1 gap-3 lg:grid-cols-[1fr_auto_1fr] lg:items-center">
      <div class="flex flex-wrap items-center gap-3 text-sm text-slate-600">
        <BaseSelect
          :model-value="limit"
          :options="pageSizeSelectOptions"
          :disabled="loading"
          @update:model-value="onLimitChange"
        />
        <span>Showing {{ rangeStart }}-{{ rangeEnd }} of {{ total }}</span>
        <span class="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700">
          {{ selectedCount }} selected
        </span>
        <BaseButton
          variant="ghost"
          size="sm"
          class="rounded-full border border-slate-200 bg-white !px-2.5 !py-1 !text-xs !font-medium text-slate-600"
          :disabled="loading || total === 0"
          :title="allMatchingSelected ? 'Disable select all matching' : 'Select all matching filters'"
          @click="emit('toggle-all-matching')"
        >
          <CheckCircleIcon class="h-3.5 w-3.5" />
          {{ allMatchingSelected ? "All matching selected" : "Select all matching" }}
        </BaseButton>
      </div>

      <div class="flex items-center justify-center gap-2">
        <IconButton 
          title="Tag selected/filtered" 
          tone="primary" 
          :disabled="loading || !actionsEnabled" 
          @click="emit('open-tag')"
        >
          <TagIcon class="h-4 w-4" />
        </IconButton>
        <IconButton
          title="Ignore selected/filtered"
          tone="warning"
          :disabled="loading || !actionsEnabled"
          @click="emit('ignore')"
        >
          <NoSymbolIcon class="h-4 w-4" />
        </IconButton>
        <IconButton
          title="Unignore selected/filtered"
          tone="info"
          :disabled="loading || !actionsEnabled"
          @click="emit('unignore')"
        >
          <EyeIcon class="h-4 w-4" />
        </IconButton>
        <IconButton
          title="Clear selection"
          :disabled="loading || selectedCount === 0"
          @click="emit('clear-selection')"
        >
          <XMarkIcon class="h-4 w-4" />
        </IconButton>
      </div>

      <div class="flex items-center justify-start gap-2 lg:justify-end">
        <span class="text-sm text-slate-600">Page {{ currentPage }} / {{ totalPages }}</span>
        <IconButton title="First page" :disabled="loading || !hasPrevious" @click="emit('go-first')">
          <BackwardIcon class="h-4 w-4" />
        </IconButton>
        <IconButton title="Previous page" :disabled="loading || !hasPrevious" @click="emit('go-prev')">
          <ChevronLeftIcon class="h-4 w-4" />
        </IconButton>
        <IconButton title="Next page" :disabled="loading || !hasNext" @click="emit('go-next')">
          <ChevronRightIcon class="h-4 w-4" />
        </IconButton>
        <IconButton title="Last page" :disabled="loading || !hasNext" @click="emit('go-last')">
          <ForwardIcon class="h-4 w-4" />
        </IconButton>
      </div>
    </div>
  </div>
</template>

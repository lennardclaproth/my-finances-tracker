<script setup lang="ts">
import {
  BackwardIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ForwardIcon,
} from "@heroicons/vue/24/outline";
import { computed } from "vue";
import BaseSelect from "../atoms/BaseSelect.vue";
import IconButton from "../atoms/IconButton.vue";

interface Props {
  limit: number;
  offset: number;
  count: number;
  total: number;
  loading?: boolean;
  pageSizeOptions?: number[];
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  pageSizeOptions: () => [10, 25, 50, 100],
});

const emit = defineEmits<{
  "change-limit": [value: number];
  "go-first": [];
  "go-prev": [];
  "go-next": [];
  "go-last": [];
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
    <div class="grid grid-cols-1 gap-3 lg:grid-cols-[1fr_auto] lg:items-center">
      <div class="flex flex-wrap items-center gap-3 text-sm text-slate-600">
        <BaseSelect
          :model-value="limit"
          :options="pageSizeSelectOptions"
          :disabled="loading"
          @update:model-value="onLimitChange"
        />
        <span>Showing {{ rangeStart }}-{{ rangeEnd }} of {{ total }}</span>
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

<script setup lang="ts">
import { FunnelIcon } from "@heroicons/vue/24/outline";
import type { DirectionFilter } from "../../types/cashflow";
import BasePopover from "../atoms/BasePopover.vue";
import BaseSelect from "../atoms/BaseSelect.vue";
import IconButton from "../atoms/IconButton.vue";

interface Props {
  modelValue: DirectionFilter;
  loading?: boolean;
}

withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  "update:modelValue": [value: DirectionFilter];
}>();

const options: Array<{ label: string; value: DirectionFilter }> = [
  { label: "All", value: "" },
  { label: "In", value: "in" },
  { label: "Out", value: "out" },
];
</script>

<template>
  <BasePopover
    :disabled="loading"
    panel-class="w-44 rounded-lg border border-slate-200 bg-white p-2 shadow-lg"
  >
    <template #trigger="{ toggle }">
      <IconButton
        :disabled="loading"
        title="Filter Direction"
        :tone="modelValue ? 'primary' : 'neutral'"
        @click="toggle"
      >
        <FunnelIcon class="h-4 w-4" />
      </IconButton>
    </template>

    <template #default>
      <label class="mb-1 block text-[11px] font-semibold uppercase tracking-wide text-slate-500">
        Direction
      </label>
      <BaseSelect
        :model-value="modelValue"
        :options="options"
        :disabled="loading"
        class="w-full"
        @update:model-value="emit('update:modelValue', $event as DirectionFilter)"
      />
    </template>
  </BasePopover>
</template>

<script setup lang="ts">
import { FunnelIcon } from "@heroicons/vue/24/outline";
import BasePopover from "../atoms/BasePopover.vue";
import BaseToggle from "../atoms/BaseToggle.vue";
import IconButton from "../atoms/IconButton.vue";

interface Props {
  showHidden: boolean;
  loading?: boolean;
}

withDefaults(defineProps<Props>(), {
  loading: false,
});

const emit = defineEmits<{
  "update:showHidden": [value: boolean];
}>();
</script>

<template>
  <BasePopover
    :disabled="loading"
    :portal="true"
    z-index-class="z-[60]"
    panel-class="w-44 rounded-lg border border-slate-200 bg-white p-2 shadow-lg"
  >
    <template #trigger="{ toggle }">
      <IconButton
        :disabled="loading"
        title="Filter Visibility"
        :tone="showHidden ? 'neutral' : 'primary'"
        @click="toggle"
      >
        <FunnelIcon class="h-4 w-4" />
      </IconButton>
    </template>

    <template #default>
      <label class="mb-1 block text-[11px] font-semibold uppercase tracking-wide text-slate-500">
        Visibility
      </label>
      <label class="flex cursor-pointer items-center justify-between gap-2 rounded-md bg-slate-50 px-2 py-2 text-xs text-slate-700">
        <span>Show hidden</span>
        <BaseToggle
          :checked="showHidden"
          :disabled="loading"
          @update:checked="emit('update:showHidden', $event)"
        />
      </label>
    </template>
  </BasePopover>
</template>

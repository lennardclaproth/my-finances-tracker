<script setup lang="ts">
import { ref } from "vue";
import ActionMenuPopover from "../molecules/ActionMenuPopover.vue";
import ImportDataModal from "../molecules/ImportDataModal.vue";
import SearchQueryInput from "../molecules/SearchQueryInput.vue";
import DateRangePopover from "../molecules/DateRangePopover.vue";

interface Props {
  searchValue?: string;
  from?: string;
  to?: string;
  loading?: boolean;
  showFilterControls?: boolean;
}

withDefaults(defineProps<Props>(), {
  searchValue: "",
  from: "",
  to: "",
  loading: false,
  showFilterControls: true,
});

const emit = defineEmits<{
  "update:searchValue": [value: string];
  "search-debounced": [value: string];
  "date-apply": [from: string, to: string];
  "date-clear": [];
}>();

const importModalOpen = ref(false);

function openImportModal(): void {
  importModalOpen.value = true;
}

function closeImportModal(): void {
  importModalOpen.value = false;
}
</script>

<template>
  <header class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
    <ActionMenuPopover @open-import="openImportModal" />

    <div v-if="showFilterControls" class="flex flex-col gap-2 sm:flex-row sm:items-center">
      <SearchQueryInput
        :model-value="searchValue"
        :disabled="loading"
        @update:model-value="emit('update:searchValue', $event)"
        @debounced-change="emit('search-debounced', $event)"
      />

      <DateRangePopover
        :from="from"
        :to="to"
        :disabled="loading"
        @apply="(fromValue, toValue) => emit('date-apply', fromValue, toValue)"
        @clear="emit('date-clear')"
      />
    </div>
  </header>

  <ImportDataModal :open="importModalOpen" @close="closeImportModal" />
</template>

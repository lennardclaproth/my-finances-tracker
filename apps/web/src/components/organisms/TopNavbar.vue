<script setup lang="ts">
import { ref } from "vue";
import ActionMenuPopover from "../molecules/ActionMenuPopover.vue";
import AvatarAdminPopover from "../molecules/AvatarAdminPopover.vue";
import ImportDataModal from "../molecules/ImportDataModal.vue";
import SearchQueryInput from "../molecules/SearchQueryInput.vue";
import DateRangePopover from "../molecules/DateRangePopover.vue";
import PageBreadcrumb from "../molecules/PageBreadcrumb.vue";

interface Props {
  searchValue?: string;
  from?: string;
  to?: string;
  loading?: boolean;
  showFilterControls?: boolean;
  showSearchControl?: boolean;
  showDateControl?: boolean;
}

withDefaults(defineProps<Props>(), {
  searchValue: "",
  from: "",
  to: "",
  loading: false,
  showFilterControls: true,
  showSearchControl: true,
  showDateControl: true,
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
    <div class="flex min-w-0 items-center gap-3">
      <ActionMenuPopover @open-import="openImportModal" />
      <PageBreadcrumb />
    </div>

    <div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-end">
      <template v-if="showFilterControls">
        <SearchQueryInput
          v-if="showSearchControl"
          :model-value="searchValue"
          :disabled="loading"
          @update:model-value="emit('update:searchValue', $event)"
          @debounced-change="emit('search-debounced', $event)"
        />

        <DateRangePopover
          v-if="showDateControl"
          :from="from"
          :to="to"
          :disabled="loading"
          @apply="(fromValue, toValue) => emit('date-apply', fromValue, toValue)"
          @clear="emit('date-clear')"
        />
      </template>

      <AvatarAdminPopover />
    </div>
  </header>

  <ImportDataModal :open="importModalOpen" @close="closeImportModal" />
</template>

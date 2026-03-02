<script setup lang="ts">
import { PlusIcon } from "@heroicons/vue/24/outline";
import { onBeforeUnmount, onMounted, ref } from "vue";
import CreateListingModal from "../components/molecules/CreateListingModal.vue";
import ToastMessage from "../components/molecules/ToastMessage.vue";
import TopNavbar from "../components/organisms/TopNavbar.vue";
import ListingsTable from "../components/organisms/ListingsTable.vue";
import AppShellTemplate from "../components/templates/AppShellTemplate.vue";
import { fetchListings } from "../services/listings";
import { ApiError } from "../services/http";
import type { Listing } from "../types/listings";
import IconButton from "../components/atoms/IconButton.vue";

type AlertTone = "success" | "danger" | "info";

const listings = ref<Listing[]>([]);
const loading = ref(false);
const errorMessage = ref("");
const createModalOpen = ref(false);
const toast = ref<{ tone: AlertTone; message: string } | null>(null);
let toastTimer: ReturnType<typeof setTimeout> | null = null;

function showToast(tone: AlertTone, message: string): void {
  toast.value = { tone, message };
  if (toastTimer) {
    clearTimeout(toastTimer);
  }
  toastTimer = setTimeout(() => {
    toast.value = null;
    toastTimer = null;
  }, 4500);
}

function toErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    return error.message;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "Failed to load listings.";
}

async function loadListings(): Promise<void> {
  loading.value = true;
  errorMessage.value = "";
  try {
    listings.value = await fetchListings();
  } catch (error: unknown) {
    errorMessage.value = toErrorMessage(error);
  } finally {
    loading.value = false;
  }
}

async function onListingCreated(): Promise<void> {
  showToast("success", "Listing created successfully.");
  await loadListings();
}

onMounted(() => {
  void loadListings();
});

onBeforeUnmount(() => {
  if (toastTimer) {
    clearTimeout(toastTimer);
    toastTimer = null;
  }
});
</script>

<template>
  <AppShellTemplate>
    <template #top>
      <TopNavbar :show-filter-controls="false" />
    </template>

    <div class="flex h-full min-h-0 flex-col px-4 pb-24">
      <ListingsTable
        class="min-h-0 flex-1"
        :rows="listings"
        :loading="loading"
        :error-message="errorMessage"
        @retry="void loadListings()"
      />
    </div>

    <CreateListingModal :open="createModalOpen" @close="createModalOpen = false" @created="void onListingCreated()" />

    <div class="fixed bottom-6 right-6 z-40">
      <IconButton
        tone="primary"
        size="fab"
        title="Add listing"
        @click="createModalOpen = true"
      >
        <PlusIcon class="h-6 w-6" />
      </IconButton>
    </div>

    <div class="pointer-events-none fixed right-4 top-4 z-50">
      <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="translate-y-2 opacity-0"
        enter-to-class="translate-y-0 opacity-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="translate-y-0 opacity-100"
        leave-to-class="translate-y-2 opacity-0"
      >
        <ToastMessage
          v-if="toast"
          :tone="toast.tone"
          :message="toast.message"
          @close="toast = null"
        />
      </Transition>
    </div>
  </AppShellTemplate>
</template>

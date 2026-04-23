<script setup lang="ts">
import { ChevronDownIcon } from "@heroicons/vue/24/outline";
import { computed, onBeforeUnmount, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { BOOTSTRAP_ACCOUNT_NAME } from "../../config/app";
import { useAppSession } from "../../composables/useAppSession";
import BasePopover from "../atoms/BasePopover.vue";
import BaseToggle from "../atoms/BaseToggle.vue";

interface Props {
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
});

const route = useRoute();
const router = useRouter();
const session = useAppSession();
const isOpen = ref(false);
let closeTimer: ReturnType<typeof setTimeout> | null = null;

const initials = computed(() => {
  const parts = BOOTSTRAP_ACCOUNT_NAME
    .trim()
    .split(/\s+/)
    .filter(Boolean);
  if (parts.length === 0) {
    return "U";
  }
  if (parts.length === 1) {
    return parts[0].slice(0, 2).toUpperCase();
  }
  return `${parts[0][0]}${parts[parts.length - 1][0]}`.toUpperCase();
});

function openOnHover(): void {
  if (closeTimer) {
    clearTimeout(closeTimer);
    closeTimer = null;
  }
  isOpen.value = true;
}

function scheduleCloseOnLeave(): void {
  if (closeTimer) {
    clearTimeout(closeTimer);
  }
  closeTimer = setTimeout(() => {
    isOpen.value = false;
    closeTimer = null;
  }, 160);
}

function toggleFromClick(): void {
  isOpen.value = !isOpen.value;
}

function onToggleAdminMode(value: boolean): void {
  session.setAdminMode(value);
  if (!value && route.path.startsWith("/admin")) {
    void router.push({ path: "/cashflow" });
  }
}

onBeforeUnmount(() => {
  if (closeTimer) {
    clearTimeout(closeTimer);
    closeTimer = null;
  }
});
</script>

<template>
  <div class="relative" @mouseenter="openOnHover" @mouseleave="scheduleCloseOnLeave">
    <BasePopover
      v-model:open="isOpen"
      :disabled="props.disabled"
      align="right"
      offset-class="mt-0"
      panel-class="w-72 max-w-[calc(100vw-2rem)]"
      z-index-class="z-40"
    >
      <template #trigger>
        <button
          type="button"
          class="group inline-flex items-center gap-2 rounded-full border border-slate-300 bg-white/90 px-2 py-1.5 text-sm text-slate-700 shadow-sm transition hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-300"
          :aria-expanded="isOpen"
          aria-haspopup="menu"
          title="Open account menu"
          @click="toggleFromClick"
        >
          <span class="inline-flex h-8 w-8 items-center justify-center rounded-full bg-slate-900 text-xs font-semibold text-white">
            {{ initials }}
          </span>
          <ChevronDownIcon class="h-4 w-4 text-slate-400 transition-transform duration-200" :class="isOpen ? 'rotate-180' : ''" />
        </button>
      </template>

      <template #default="{ close }">
        <section
          class="menu-panel translate-y-2 overflow-hidden rounded-2xl border border-slate-300 bg-white/95 shadow-xl backdrop-blur"
          @mouseenter="openOnHover"
          @mouseleave="scheduleCloseOnLeave"
        >
          <header class="border-b border-slate-100 px-4 py-3">
            <p class="font-secondary text-[11px] font-semibold uppercase tracking-[0.14em] text-slate-500">Account</p>
            <p class="font-secondary mt-1 truncate text-sm font-semibold text-slate-800">{{ BOOTSTRAP_ACCOUNT_NAME }}</p>
          </header>

          <div class="p-2">
            <div
              class="flex w-full items-center justify-between gap-3 rounded-xl px-3 py-2.5 text-left text-slate-700 transition hover:bg-slate-100"
              @click="onToggleAdminMode(!session.adminMode.value)"
            >
              <span>
                <span class="block text-sm font-semibold">Admin mode</span>
                <span class="block text-xs text-slate-500">Switch Action menu to admin pages.</span>
              </span>
              <span @click.stop>
                <BaseToggle
                  :checked="session.adminMode.value"
                  @update:checked="onToggleAdminMode"
                />
              </span>
            </div>
          </div>

          <footer class="flex justify-end border-t border-slate-100 px-3 py-2">
            <button
              type="button"
              class="rounded-md px-2 py-1 text-xs font-medium text-slate-500 transition hover:bg-slate-100 hover:text-slate-700"
              @click="close"
            >
              Close
            </button>
          </footer>
        </section>
      </template>
    </BasePopover>
  </div>
</template>

<style scoped>
@keyframes menu-panel-open {
  0% {
    opacity: 0;
    transform: translateY(-8px) scaleY(0.22);
  }
  100% {
    opacity: 1;
    transform: translateY(0) scaleY(1);
  }
}

.menu-panel {
  transform-origin: top;
  animation: menu-panel-open 220ms cubic-bezier(0.16, 1, 0.3, 1);
}
</style>

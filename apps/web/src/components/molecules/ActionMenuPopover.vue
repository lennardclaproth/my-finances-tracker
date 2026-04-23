<script setup lang="ts">
import {
  BuildingOffice2Icon,
  Bars3Icon,
  CalendarDaysIcon,
  ChartBarSquareIcon,
  ChevronDownIcon,
  CloudArrowUpIcon,
  QueueListIcon,
  TagIcon,
} from "@heroicons/vue/24/outline";
import type { Component } from "vue";
import { computed, onBeforeUnmount, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAppSession } from "../../composables/useAppSession";
import BasePopover from "../atoms/BasePopover.vue";

interface Props {
  disabled?: boolean;
}

interface RouteActionItem {
  id: string;
  kind: "route";
  label: string;
  description: string;
  to: string;
  icon: Component;
}

interface ModalActionItem {
  id: string;
  kind: "modal";
  label: string;
  description: string;
  icon: Component;
}

type ActionItem = RouteActionItem | ModalActionItem;

withDefaults(defineProps<Props>(), {
  disabled: false,
});

const emit = defineEmits<{
  "open-import": [];
}>();

const route = useRoute();
const router = useRouter();
const session = useAppSession();
const isOpen = ref(false);
let closeTimer: ReturnType<typeof setTimeout> | null = null;

const userActions: ActionItem[] = [
  {
    id: "cashflow",
    kind: "route",
    label: "Cashflow",
    description: "Review and tag transactions.",
    to: "/cashflow",
    icon: TagIcon,
  },
  {
    id: "portfolio",
    kind: "route",
    label: "Portfolio",
    description: "View positions and portfolio growth.",
    to: "/portfolio",
    icon: ChartBarSquareIcon,
  },
  {
    id: "assets",
    kind: "route",
    label: "Assets",
    description: "Track asset classes and growth.",
    to: "/assets",
    icon: BuildingOffice2Icon,
  },
  {
    id: "import",
    kind: "modal",
    label: "Import data",
    description: "Upload files into your workspace.",
    icon: CloudArrowUpIcon,
  },
];

const adminActions: ActionItem[] = [
  {
    id: "admin-listings",
    kind: "route",
    label: "Listings",
    description: "Create and review listing metadata.",
    to: "/admin/listings",
    icon: QueueListIcon,
  },
  {
    id: "admin-dailies",
    kind: "route",
    label: "Dailies",
    description: "Review listing daily price history.",
    to: "/admin/dailies",
    icon: CalendarDaysIcon,
  },
];

const actions = computed<ActionItem[]>(() => (session.adminMode.value ? adminActions : userActions));

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

function isActiveAction(item: ActionItem): boolean {
  return item.kind === "route" && route.path === item.to;
}

async function runAction(item: ActionItem, close: () => void): Promise<void> {
  if (item.kind === "route") {
    if (route.path !== item.to) {
      await router.push({ path: item.to });
    }
    close();
    return;
  }

  emit("open-import");
  close();
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
      :disabled="disabled"
      align="left"
      offset-class="mt-0"
      panel-class="w-80 max-w-[calc(100vw-2rem)]"
      z-index-class="z-40"
    >
      <template #trigger>
        <button
          type="button"
          class="group inline-flex items-center gap-3 rounded-full border border-slate-300 bg-white/90 px-3 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:bg-slate-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-300"
          :aria-expanded="isOpen"
          aria-haspopup="menu"
          title="Open actions"
          @click="toggleFromClick"
        >
          <span
            class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-indigo-200 bg-indigo-50 text-indigo-700 transition-transform duration-500 ease-out group-hover:rotate-[360deg]"
            :class="isOpen ? 'rotate-[360deg]' : ''"
          >
            <Bars3Icon class="h-5 w-5" />
          </span>
          <span class="whitespace-nowrap">Actions</span>
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
            <p class="font-secondary text-[11px] font-semibold uppercase tracking-[0.14em] text-slate-500">
              {{ session.adminMode.value ? "Admin actions" : "Quick actions" }}
            </p>
          </header>

          <div class="p-2">
            <button
              v-for="item in actions"
              :key="item.id"
              type="button"
              class="menu-item flex w-full items-start gap-3 rounded-xl px-3 py-2.5 text-left transition"
              :class="
                isActiveAction(item)
                  ? 'bg-indigo-50 text-indigo-800'
                  : 'text-slate-700 hover:bg-slate-100'
              "
              @click="void runAction(item, close)"
            >
              <span
                class="mt-0.5 inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border"
                :class="isActiveAction(item) ? 'border-indigo-200 bg-indigo-100 text-indigo-700' : 'border-slate-300 bg-white text-slate-500'"
              >
                <component :is="item.icon" class="h-4 w-4" />
              </span>
              <span class="min-w-0">
                <span class="block text-sm font-semibold">{{ item.label }}</span>
                <span class="block text-xs text-slate-500">{{ item.description }}</span>
              </span>
            </button>
          </div>
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

@keyframes menu-item-enter {
  0% {
    opacity: 0;
    transform: translateY(-6px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

.menu-panel {
  transform-origin: top;
  animation: menu-panel-open 220ms cubic-bezier(0.16, 1, 0.3, 1);
}

.menu-item {
  opacity: 0;
  animation: menu-item-enter 200ms ease-out forwards;
}

.menu-item:nth-child(1) {
  animation-delay: 50ms;
}

.menu-item:nth-child(2) {
  animation-delay: 90ms;
}

.menu-item:nth-child(3) {
  animation-delay: 130ms;
}

.menu-item:nth-child(4) {
  animation-delay: 170ms;
}
</style>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch, type CSSProperties } from "vue";

interface Props {
  open?: boolean;
  disabled?: boolean;
  panelClass?: string;
  align?: "left" | "right";
  side?: "bottom" | "top";
  offsetClass?: string;
  zIndexClass?: string;
  closeOnOutside?: boolean;
  portal?: boolean;
  offsetPx?: number;
}

const props = withDefaults(defineProps<Props>(), {
  open: undefined,
  disabled: false,
  panelClass: "",
  align: "right",
  side: "bottom",
  offsetClass: "mt-2",
  zIndexClass: "z-30",
  closeOnOutside: true,
  portal: false,
  offsetPx: 8,
});

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

const rootRef = ref<HTMLElement | null>(null);
const panelRef = ref<HTMLElement | null>(null);
const uncontrolledOpen = ref(false);
const portalPanelStyle = ref<CSSProperties>({});

const isControlled = computed(() => props.open !== undefined);
const isOpen = computed(() =>
  isControlled.value ? Boolean(props.open) : uncontrolledOpen.value,
);

const panelPositionClass = computed(() =>
  props.align === "left" ? "left-0" : "right-0",
);

const panelSideClass = computed(() =>
  props.side === "top" ? "bottom-full" : "top-full",
);

function setOpen(value: boolean): void {
  if (props.disabled && value) {
    return;
  }

  if (!isControlled.value) {
    uncontrolledOpen.value = value;
  }
  emit("update:open", value);
}

function openPopover(): void {
  setOpen(true);
}

function closePopover(): void {
  setOpen(false);
}

function togglePopover(): void {
  setOpen(!isOpen.value);
}

function onDocumentClick(event: MouseEvent): void {
  if (!props.closeOnOutside || !isOpen.value) {
    return;
  }

  const target = event.target as Node | null;
  if (!target) {
    closePopover();
    return;
  }

  const isInsideRoot = Boolean(rootRef.value?.contains(target));
  const isInsidePanel = Boolean(panelRef.value?.contains(target));
  if (!isInsideRoot && !isInsidePanel) {
    closePopover();
  }
}

function updatePortalPosition(): void {
  if (!props.portal || !isOpen.value || !rootRef.value) {
    return;
  }

  const rect = rootRef.value.getBoundingClientRect();
  const viewportWidth = window.innerWidth;
  const viewportHeight = window.innerHeight;
  const nextStyle: CSSProperties = {
    position: "fixed",
  };

  if (props.align === "left") {
    nextStyle.left = `${Math.max(8, rect.left)}px`;
    nextStyle.right = "auto";
  } else {
    nextStyle.right = `${Math.max(8, viewportWidth - rect.right)}px`;
    nextStyle.left = "auto";
  }

  if (props.side === "top") {
    nextStyle.bottom = `${Math.max(8, viewportHeight - rect.top + props.offsetPx)}px`;
    nextStyle.top = "auto";
  } else {
    nextStyle.top = `${Math.max(8, rect.bottom + props.offsetPx)}px`;
    nextStyle.bottom = "auto";
  }

  portalPanelStyle.value = nextStyle;
}

onMounted(() => {
  document.addEventListener("mousedown", onDocumentClick);
});

onBeforeUnmount(() => {
  document.removeEventListener("mousedown", onDocumentClick);
  window.removeEventListener("resize", updatePortalPosition);
  window.removeEventListener("scroll", updatePortalPosition, true);
});

watch(isOpen, (open) => {
  if (!props.portal) {
    return;
  }

  if (open) {
    void nextTick(() => {
      updatePortalPosition();
    });
    window.addEventListener("resize", updatePortalPosition);
    window.addEventListener("scroll", updatePortalPosition, true);
    return;
  }

  window.removeEventListener("resize", updatePortalPosition);
  window.removeEventListener("scroll", updatePortalPosition, true);
});

watch(
  () => [props.portal, props.align, props.side, props.offsetPx],
  () => {
    if (!props.portal || !isOpen.value) {
      return;
    }
    void nextTick(() => {
      updatePortalPosition();
    });
  },
);
</script>

<template>
  <div ref="rootRef" class="relative">
    <slot
      name="trigger"
      :is-open="isOpen"
      :open="openPopover"
      :close="closePopover"
      :toggle="togglePopover"
    />

    <Teleport v-if="portal" to="body">
      <div v-if="isOpen" ref="panelRef" :class="[zIndexClass, panelClass]" :style="portalPanelStyle">
        <slot :is-open="isOpen" :close="closePopover" :open="openPopover" :toggle="togglePopover" />
      </div>
    </Teleport>

    <div
      v-else-if="isOpen"
      ref="panelRef"
      :class="['absolute', panelPositionClass, panelSideClass, offsetClass, zIndexClass, panelClass]"
    >
      <slot :is-open="isOpen" :close="closePopover" :open="openPopover" :toggle="togglePopover" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from "vue";

interface Props {
  open?: boolean;
  disabled?: boolean;
  panelClass?: string;
  align?: "left" | "right";
  side?: "bottom" | "top";
  offsetClass?: string;
  zIndexClass?: string;
  closeOnOutside?: boolean;
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
});

const emit = defineEmits<{
  "update:open": [value: boolean];
}>();

const rootRef = ref<HTMLElement | null>(null);
const uncontrolledOpen = ref(false);

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
  if (target && rootRef.value && !rootRef.value.contains(target)) {
    closePopover();
  }
}

onMounted(() => {
  document.addEventListener("mousedown", onDocumentClick);
});

onBeforeUnmount(() => {
  document.removeEventListener("mousedown", onDocumentClick);
});
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

    <div
      v-if="isOpen"
      :class="['absolute', panelPositionClass, panelSideClass, offsetClass, zIndexClass, panelClass]"
    >
      <slot :is-open="isOpen" :close="closePopover" :open="openPopover" :toggle="togglePopover" />
    </div>
  </div>
</template>

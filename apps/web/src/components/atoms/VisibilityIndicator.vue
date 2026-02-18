<script setup lang="ts">
import { EyeIcon, EyeSlashIcon } from "@heroicons/vue/24/outline";

interface Props {
  visible: boolean;
}

defineProps<Props>();
</script>

<template>
  <EyeIcon
    v-if="visible"
    class="h-4 w-4 text-slate-400"
    title="Visible"
    aria-label="Visible"
  />
  <EyeSlashIcon
    v-else
    class="h-4 w-4 text-slate-400"
    title="Not visible"
    aria-label="Not visible"
  />
</template>

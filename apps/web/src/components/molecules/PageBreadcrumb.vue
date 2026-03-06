<script setup lang="ts">
import { ChevronRightIcon } from "@heroicons/vue/24/solid";
import { computed } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

const segmentLabelMap: Record<string, string> = {
  tagging: "Cashflow",
  analyze: "Portfolio",
  admin: "Admin",
  listings: "Listings",
  dailies: "Dailies",
  cashflow: "Cashflow",
  portfolio: "Portfolio",
};

const crumbs = computed(() => {
  const segments = route.path.split("/").filter(Boolean);
  if (segments.length === 0) {
    return ["Home"];
  }
  return segments.map((segment) => segmentLabelMap[segment] ?? toTitle(segment));
});

function toTitle(input: string): string {
  return input
    .split("-")
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}
</script>

<template>
  <nav class="flex min-w-0 items-center" aria-label="Breadcrumb">
    <ol class="flex min-w-0 items-center gap-1 overflow-x-auto rounded-full border border-slate-200 bg-white/80 px-3 py-2 text-xs shadow-sm">
      <li
        v-for="(crumb, index) in crumbs"
        :key="`${crumb}-${index}`"
        class="inline-flex shrink-0 items-center gap-1"
      >
        <ChevronRightIcon v-if="index > 0" class="h-3 w-3 text-slate-400" />
        <span :class="index === crumbs.length - 1 ? 'font-semibold text-slate-800' : 'text-slate-500'">
          {{ crumb }}
        </span>
      </li>
    </ol>
  </nav>
</template>

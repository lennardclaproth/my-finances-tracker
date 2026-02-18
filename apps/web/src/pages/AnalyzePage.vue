<script setup lang="ts">
import TopNavbar from "../components/organisms/TopNavbar.vue";
import AppShellTemplate from "../components/templates/AppShellTemplate.vue";

const metricCards = [
  {
    title: "Income this month",
    value: "$0.00",
    hint: "Waiting for imported and tagged transactions.",
  },
  {
    title: "Spending this month",
    value: "$0.00",
    hint: "Link more actions to categorize expenses.",
  },
  {
    title: "Savings rate",
    value: "0%",
    hint: "Will update once income and expenses are classified.",
  },
];
</script>

<template>
  <AppShellTemplate>
    <template #top>
      <TopNavbar :show-filter-controls="false" />
    </template>

    <div class="h-full overflow-y-auto px-4 pb-4">
      <section class="rounded-xl border border-slate-200 bg-white px-4 py-4 shadow-sm">
        <h1 class="text-lg font-semibold text-slate-900">Analyze</h1>
        <p class="mt-1 text-sm text-slate-500">
          Overview workspace for complete finances, trends, and allocation insights.
        </p>
      </section>

      <section class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-3">
        <article
          v-for="card in metricCards"
          :key="card.title"
          class="rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
        >
          <p class="text-xs font-semibold uppercase tracking-wide text-slate-500">{{ card.title }}</p>
          <p class="mt-2 text-2xl font-semibold text-slate-900">{{ card.value }}</p>
          <p class="mt-1 text-xs text-slate-500">{{ card.hint }}</p>
        </article>
      </section>

      <!-- <GraphPlaceholderCards class="mt-4" /> -->
    </div>
  </AppShellTemplate>
</template>

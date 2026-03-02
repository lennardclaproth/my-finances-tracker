<script setup lang="ts">
import { computed } from "vue";
import PortfolioKpiCard from "../molecules/PortfolioKpiCard.vue";

interface Props {
  marketValue?: number;
  totalPnl?: number;
  totalPnlPct?: number;
  loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
});

const MONEY_SCALE = 1_000_000;
const moneyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 2,
});

function formatMoney(value?: number): string {
  if (value === undefined || value === null) {
    return "-";
  }
  return moneyFormatter.format(value / MONEY_SCALE);
}

function formatPct(value?: number): string {
  if (value === undefined || value === null) {
    return "-";
  }
  return `${value.toFixed(2)}%`;
}

function toneFromNumber(value?: number): "neutral" | "positive" | "negative" {
  if (value === undefined || value === null || value === 0) {
    return "neutral";
  }
  return value > 0 ? "positive" : "negative";
}

const totalPnLTone = computed(() => toneFromNumber(props.totalPnl));
const totalPnLPctTone = computed(() => toneFromNumber(props.totalPnlPct));
</script>

<template>
  <section class="grid grid-cols-1 gap-2 md:grid-cols-3">
    <PortfolioKpiCard label="Total PnL" :value="formatMoney(props.totalPnl)" :tone="totalPnLTone" :loading="props.loading" />
    <PortfolioKpiCard label="Total PnL %" :value="formatPct(props.totalPnlPct)" :tone="totalPnLPctTone" :loading="props.loading" />
    <PortfolioKpiCard label="Market Value" :value="formatMoney(props.marketValue)" tone="neutral" :loading="props.loading" />
  </section>
</template>

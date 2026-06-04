<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import StatCard from './StatCard.svelte';

  import { trendIndicatorFormats } from '$lib/components/molecules/trend-indicator/trend-indicator.types';

  const sampleTrend = [120, 128, 124, 140, 138, 150, 162];

  const { Story } = defineMeta({
    title: 'Organisms/StatCard',
    component: StatCard,
    tags: ['autodocs'],
    argTypes: {
      label: { control: 'text' },
      amount: { control: 'number' },
      currency: { control: 'text' },
      change: { control: 'number' },
      changeFormat: { control: 'select', options: trendIndicatorFormats }
    }
  });
</script>

<Story
  name="Playground"
  args={{
    label: 'Total balance',
    amount: 16240.55,
    currency: 'EUR',
    change: 4.2,
    changeFormat: 'percent',
    trend: sampleTrend
  }}
/>

<Story name="Dashboard Grid" asChild>
  <div class="grid max-w-3xl grid-cols-1 gap-4 sm:grid-cols-3">
    <StatCard label="Total balance" amount={16240.55} change={4.2} trend={sampleTrend} />
    <StatCard
      label="Spending (30d)"
      amount={-2310.18}
      change={-1.8}
      trend={[60, 58, 62, 70, 66, 64, 72]}
    />
    <StatCard label="Net worth" amount={84120} change={0} trend={[80, 81, 80, 82, 83, 84, 84]} />
  </div>
</Story>

<Story name="Without Trend" asChild>
  <div class="max-w-xs">
    <StatCard label="Available credit" amount={3500} change={120} changeFormat="currency" />
  </div>
</Story>

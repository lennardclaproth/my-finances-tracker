<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import TrendIndicator from './TrendIndicator.svelte';

  import { trendIndicatorFormats, trendIndicatorSizes } from './trend-indicator.types';

  const { Story } = defineMeta({
    title: 'Molecules/TrendIndicator',
    component: TrendIndicator,
    tags: ['autodocs'],
    argTypes: {
      value: { control: 'number' },
      format: { control: 'select', options: trendIndicatorFormats },
      currency: { control: 'text' },
      size: { control: 'select', options: trendIndicatorSizes },
      showArrow: { control: 'boolean' }
    }
  });
</script>

<Story name="Playground" args={{ value: 2.34, format: 'percent', size: 'md', showArrow: true }} />

<Story name="Directions" asChild>
  <div class="flex items-center gap-4">
    <TrendIndicator value={2.34} />
    <TrendIndicator value={-1.12} />
    <TrendIndicator value={0} />
  </div>
</Story>

<Story name="Formats" asChild>
  <div class="flex items-center gap-4">
    <TrendIndicator value={3.5} format="percent" />
    <TrendIndicator value={1240.5} format="currency" />
    <TrendIndicator value={-87} format="number" />
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="flex items-center gap-4">
    {#each trendIndicatorSizes as size}
      <TrendIndicator value={4.2} {size} />
    {/each}
  </div>
</Story>

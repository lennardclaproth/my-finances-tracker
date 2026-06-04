<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Money from './Money.svelte';

  import {
    moneySignDisplays,
    moneySizes,
    moneyWeights
  } from './money.types';

  const { Story } = defineMeta({
    title: 'Atoms/Money',
    component: Money,
    tags: ['autodocs'],
    argTypes: {
      amount: { control: 'number' },
      currency: { control: 'text' },
      locale: { control: 'text' },
      size: { control: 'select', options: moneySizes },
      weight: { control: 'select', options: moneyWeights },
      signDisplay: { control: 'select', options: moneySignDisplays },
      colored: { control: 'boolean' }
    }
  });
</script>

<Story
  name="Playground"
  args={{
    amount: 1234.56,
    currency: 'EUR',
    locale: 'en-US',
    size: 'md',
    weight: 'medium',
    signDisplay: 'auto',
    colored: false
  }}
/>

<Story name="Colored By Sign" asChild>
  <div class="flex flex-col gap-2">
    <Money amount={1820.42} colored />
    <Money amount={-340.18} colored />
    <Money amount={0} colored />
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="flex flex-col gap-2">
    {#each moneySizes as size}
      <Money amount={1234.56} {size} weight="semibold" />
    {/each}
  </div>
</Story>

<Story name="Currencies" asChild>
  <div class="flex flex-col gap-2">
    <Money amount={1234.56} currency="EUR" />
    <Money amount={1234.56} currency="USD" />
    <Money amount={1234.56} currency="GBP" />
    <Money amount={1234} currency="JPY" />
  </div>
</Story>

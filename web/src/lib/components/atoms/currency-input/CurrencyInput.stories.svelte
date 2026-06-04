<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import CurrencyInput from './CurrencyInput.svelte';

  import {
    currencyInputIntents,
    currencyInputShapes,
    currencyInputSizes
  } from './currency-input.types';

  const { Story } = defineMeta({
    title: 'Atoms/CurrencyInput',
    component: CurrencyInput,
    tags: ['autodocs'],
    args: { ariaLabel: 'Amount' },
    argTypes: {
      value: { control: 'text' },
      currencySymbol: { control: 'text' },
      placeholder: { control: 'text' },
      size: { control: 'select', options: currencyInputSizes },
      intent: { control: 'select', options: currencyInputIntents },
      shape: { control: 'select', options: currencyInputShapes },
      disabled: { control: 'boolean' },
      readonly: { control: 'boolean' },
      required: { control: 'boolean' }
    }
  });
</script>

<Story
  name="Playground"
  args={{
    value: '1234.56',
    currencySymbol: '€',
    placeholder: '0.00',
    size: 'md',
    intent: 'default',
    shape: 'default',
    disabled: false,
    readonly: false,
    required: false
  }}
/>

<Story name="Symbols" asChild>
  <div class="grid max-w-md gap-4">
    <CurrencyInput currencySymbol="€" value="1234.56" ariaLabel="Euro amount" />
    <CurrencyInput currencySymbol="$" value="1234.56" ariaLabel="Dollar amount" />
    <CurrencyInput currencySymbol="£" value="1234.56" ariaLabel="Pound amount" />
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="grid max-w-md gap-4">
    {#each currencyInputSizes as size}
      <CurrencyInput {size} value="1234.56" ariaLabel={`${size} amount`} />
    {/each}
  </div>
</Story>

<Story name="Intents" asChild>
  <div class="grid max-w-md gap-4">
    {#each currencyInputIntents as intent}
      <CurrencyInput {intent} value="1234.56" ariaLabel={`${intent} amount`} />
    {/each}
  </div>
</Story>

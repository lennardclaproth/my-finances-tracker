<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Radio from './Radio.svelte';

  import { radioSizes } from './radio.types';

  const { Story } = defineMeta({
    title: 'Atoms/Radio',
    component: Radio,
    tags: ['autodocs'],
    argTypes: {
      group: { control: 'text' },
      value: { control: 'text' },
      size: { control: 'select', options: radioSizes },
      invalid: { control: 'boolean' },
      disabled: { control: 'boolean' }
    }
  });
</script>

<Story
  name="Playground"
  args={{ value: 'a', group: 'a', size: 'md', invalid: false, disabled: false }}
/>

<Story name="States" asChild>
  <div class="flex items-center gap-4">
    <Radio value="a" group="a" />
    <Radio value="b" group="a" />
    <Radio value="c" group="a" invalid />
    <Radio value="d" group="a" disabled />
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="flex items-center gap-4">
    {#each radioSizes as size}
      <Radio {size} value="a" group="a" />
    {/each}
  </div>
</Story>

<Story name="Group" asChild>
  <fieldset class="flex flex-col gap-2">
    <label class="flex items-center gap-2">
      <Radio value="income" group="income" /> Income
    </label>
    <label class="flex items-center gap-2">
      <Radio value="expense" group="income" /> Expense
    </label>
    <label class="flex items-center gap-2">
      <Radio value="transfer" group="income" /> Transfer
    </label>
  </fieldset>
</Story>

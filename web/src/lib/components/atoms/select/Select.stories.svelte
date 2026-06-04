<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Select from './Select.svelte';

  import { selectIntents, selectShapes, selectSizes } from './select.types';

  const sampleOptions = [
    { value: 'checking', label: 'Checking account' },
    { value: 'savings', label: 'Savings account' },
    { value: 'investment', label: 'Investment account' },
    { value: 'closed', label: 'Closed account', disabled: true }
  ];

  const { Story } = defineMeta({
    title: 'Atoms/Select',
    component: Select,
    tags: ['autodocs'],
    args: {
      options: sampleOptions,
      ariaLabel: 'Account type'
    },
    argTypes: {
      size: { control: 'select', options: selectSizes },
      intent: { control: 'select', options: selectIntents },
      shape: { control: 'select', options: selectShapes },
      placeholder: { control: 'text' },
      disabled: { control: 'boolean' },
      required: { control: 'boolean' }
    }
  });
</script>

<Story
  name="Playground"
  args={{
    placeholder: 'Select an account',
    size: 'md',
    intent: 'default',
    shape: 'default',
    disabled: false,
    required: false
  }}
/>

<Story name="Sizes" asChild>
  <div class="grid max-w-md gap-4">
    {#each selectSizes as size}
      <Select {size} options={sampleOptions} ariaLabel={`${size} select`} placeholder="Select…" />
    {/each}
  </div>
</Story>

<Story name="Intents" asChild>
  <div class="grid max-w-md gap-4">
    {#each selectIntents as intent}
      <Select
        {intent}
        options={sampleOptions}
        ariaLabel={`${intent} select`}
        placeholder="Select…"
      />
    {/each}
  </div>
</Story>

<Story name="Disabled" asChild>
  <Select disabled options={sampleOptions} ariaLabel="Disabled select" placeholder="Select…" />
</Story>

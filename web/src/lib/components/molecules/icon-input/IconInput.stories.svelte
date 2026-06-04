<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import IconInput from './IconInput.svelte';

  import {
    inputIntents,
    inputSizes
  } from '$lib/components/atoms/input/input.types';

  const { Story } = defineMeta({
    title: 'Molecules/IconInput',
    component: IconInput,
    tags: ['autodocs'],
    argTypes: {
      value: {
        control: 'text'
      },
      placeholder: {
        control: 'text'
      },
      size: {
        control: 'select',
        options: inputSizes
      },
      intent: {
        control: 'select',
        options: inputIntents
      },
      leftIcon: {
        control: 'text'
      },
      rightIcon: {
        control: 'text'
      },
      disabled: {
        control: 'boolean'
      },
      readonly: {
        control: 'boolean'
      },
      required: {
        control: 'boolean'
      }
    }
  });
</script>

<Story
  name="Playground"
  args={{
    value: '',
    placeholder: 'Search...',
    size: 'md',
    intent: 'default',
    leftIcon: 'heroicons:magnifying-glass',
    rightIcon: '',
    disabled: false,
    readonly: false,
    required: false
  }}
/>

<Story name="Left Icon" asChild>
  <div class="grid max-w-md gap-4">
    <IconInput
      leftIcon="heroicons:magnifying-glass"
      placeholder="Search..."
    />

    <IconInput
      leftIcon="heroicons:envelope"
      placeholder="Email address"
    />
  </div>
</Story>

<Story name="Right Icon" asChild>
  <div class="grid max-w-md gap-4">
    <IconInput
      rightIcon="heroicons:calendar-days"
      placeholder="Pick a date"
    />
  </div>
</Story>

<Story name="Validation Icons" asChild>
  <div class="grid max-w-md gap-4">
    <IconInput
      leftIcon="heroicons:envelope"
      intent="success"
      value="looks.good@example.com"
      placeholder="Email address"
    />

    <IconInput
      leftIcon="heroicons:envelope"
      intent="error"
      value="invalid-email"
      placeholder="Email address"
    />
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="grid max-w-md gap-4">
    {#each inputSizes as size}
      <IconInput
        {size}
        leftIcon="heroicons:magnifying-glass"
        placeholder={`${size} input`}
      />
    {/each}
  </div>
</Story>

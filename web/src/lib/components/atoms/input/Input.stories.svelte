<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Input from './Input.svelte';

  import {
    inputIntents,
    inputSizes,
  } from './input.types';

  const { Story } = defineMeta({
    title: 'Atoms/Input',
    component: Input,
    tags: ['autodocs'],
    argTypes: {
      value: {
        control: 'text'
      },
      placeholder: {
        control: 'text'
      },
      type: {
        control: 'select',
        options: [
          'text',
          'email',
          'password',
          'number',
          'search',
          'tel',
          'url',
          'date',
          'datetime-local',
          'time'
        ]
      },
      size: {
        control: 'select',
        options: inputSizes
      },
    //   variant: {
    //     control: 'select',
    //     options: inputVariants
    //   },
      intent: {
        control: 'select',
        options: inputIntents
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
    placeholder: 'Enter text...',
    type: 'text',
    size: 'md',
    intent: 'default',
    disabled: false,
    readonly: false,
    required: false
  }}
/>
<!-- 
<Story name="Variants" asChild>
  <div class="grid max-w-md gap-4">
    {#each inputVariants as variant}
      <Input
        {variant}
        placeholder={`${variant} input`}
      />
    {/each}
  </div>
</Story> -->

<Story name="Sizes" asChild>
  <div class="grid max-w-md gap-4">
    {#each inputSizes as size}
      <Input
        {size}
        placeholder={`${size} input`}
      />
    {/each}
  </div>
</Story>

<Story name="Intents" asChild>
  <div class="grid max-w-md gap-4">
    <Input
      intent="default"
      placeholder="Default input"
    />

    <Input
      intent="success"
      value="Looks good"
      placeholder="Success input"
    />

    <Input
      intent="error"
      value="Invalid value"
      placeholder="Error input"
    />
  </div>
</Story>

<Story name="States" asChild>
  <div class="grid max-w-md gap-4">
    <Input placeholder="Default" />

    <Input
      placeholder="Disabled"
      disabled
    />

    <Input
      value="Read-only value"
      readonly
    />

    <Input
      intent="error"
      value="Error state"
    />

    <Input
      intent="success"
      value="Success state"
    />
  </div>
</Story>

<Story name="Shapes" asChild>
  <div class="grid max-w-md gap-4">
    <Input placeholder="Default" />

    <Input
        shape="rounded"
        placeholder="Rounded input"
    />

    <Input
      shape="pill"
      placeholder="Pill input"
    />
  </div>
</Story>
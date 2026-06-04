<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import type { ComponentProps } from 'svelte';

  import Button from './Button.svelte';
  import Icon from '$lib/components/atoms/icon/Icon.svelte';

  import {
    buttonIntents,
    buttonVariants,
    buttonSizes,
    buttonShapes
  } from './button.types';

  type ButtonProps = ComponentProps<typeof Button>;

  const { Story } = defineMeta({
    title: 'Atoms/Button',
    component: Button,
    tags: ['autodocs'],
    argTypes: {
      intent: {
        control: 'select',
        options: buttonIntents
      },
      variant: {
        control: 'select',
        options: buttonVariants
      },
      size: {
        control: 'select',
        options: buttonSizes
      },
      shape: {
        control: 'select',
        options: buttonShapes
      },
      disabled: {
        control: 'boolean'
      },
      loading: {
        control: 'boolean'
      },
      pressed: {
        control: 'boolean'
      }
    }
  });
</script>

<!--
  Button is a pure atom: it renders a native <button> and its children. Icons are composed by the
  consumer via the Icon atom in the button's children (see the "With Icon" stories). For icon-only
  buttons use the Molecules/IconButton component.
-->

{#snippet playground(args: ButtonProps)}
  <Button {...args} onclick={() => console.log('Button clicked')}>
    Button
  </Button>
{/snippet}

<Story
  name="Playground"
  args={{
    intent: 'primary',
    variant: 'solid',
    size: 'md',
    shape: 'rounded',
    disabled: false,
    loading: false,
    pressed: false
  }}
  template={playground}
/>

<Story name="Text Only" asChild>
  <div class="flex flex-wrap items-center gap-3">
    <Button>Default</Button>
    <Button intent="secondary">Secondary</Button>
    <Button intent="success">Success</Button>
    <Button intent="error">Error</Button>
  </div>
</Story>

<Story name="With Icon Left" asChild>
  <div class="flex flex-wrap items-center gap-3">
    <Button intent="success">
      <Icon icon="heroicons:check" />
      Save
    </Button>

    <Button intent="info" variant="outline">
      <Icon icon="heroicons:pencil-square" />
      Edit
    </Button>

    <Button intent="error" variant="ghost">
      <Icon icon="heroicons:trash" />
      Delete
    </Button>
  </div>
</Story>

<Story name="With Icon Right" asChild>
  <div class="flex flex-wrap items-center gap-3">
    <Button>
      Continue
      <Icon icon="heroicons:arrow-right" />
    </Button>

    <Button intent="secondary" variant="outline">
      Open
      <Icon icon="heroicons:arrow-top-right-on-square" />
    </Button>
  </div>
</Story>

<Story name="Variants" asChild>
  <div class="flex flex-col gap-8">
    {#each buttonVariants as variant}
      <section class="space-y-3">
        <h3 class="text-sm font-semibold capitalize">{variant}</h3>

        <div class="flex flex-wrap gap-3">
          {#each buttonIntents as intent}
            <Button {intent} {variant}>
              <Icon icon="heroicons:sparkles" />
              {intent}
            </Button>
          {/each}
        </div>
      </section>
    {/each}
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="flex items-center gap-3">
    {#each buttonSizes as size}
      <Button {size}>
        <Icon icon="heroicons:check" {size} />
        {size}
      </Button>
    {/each}
  </div>
</Story>

<Story name="States" asChild>
  <div class="flex flex-wrap items-center gap-3">
    <Button>
      <Icon icon="heroicons:check" />
      Default
    </Button>

    <Button disabled>
      <Icon icon="heroicons:check" />
      Disabled
    </Button>

    <Button loading>
      Loading
    </Button>

    <Button pressed>
      <Icon icon="heroicons:check" />
      Pressed
    </Button>

    <Button intent="error" variant="outline">
      <Icon icon="heroicons:trash" />
      Error
    </Button>
  </div>
</Story>

<Story name="All Combinations" asChild>
  <div class="flex flex-col gap-10">
    {#each buttonVariants as variant}
      <section class="space-y-4">
        <h3 class="text-sm font-semibold capitalize">{variant}</h3>

        <div class="grid gap-4">
          {#each buttonIntents as intent}
            <div class="flex flex-wrap items-center gap-3">
              {#each buttonSizes as size}
                <Button {intent} {variant} {size}>
                  <Icon icon="heroicons:check" {size} />
                  {intent} {size}
                </Button>
              {/each}
            </div>
          {/each}
        </div>
      </section>
    {/each}
  </div>
</Story>

<Story name="Shapes" asChild>
  <div class="flex flex-wrap items-center gap-3">
    {#each buttonShapes as shape}
      <Button {shape}>
        <Icon icon="heroicons:check" />
        {shape}
      </Button>
    {/each}
  </div>
</Story>

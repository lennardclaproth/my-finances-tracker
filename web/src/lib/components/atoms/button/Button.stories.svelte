<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import type { ComponentProps } from 'svelte';

  import Button from './Button.svelte';

  import {
    buttonIntents,
    buttonVariants,
    buttonSizes,
    buttonShapes,
    iconPositions
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
      iconPosition: {
        control: 'select',
        options: iconPositions
      },
      icon: {
        control: 'text'
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
    icon: 'heroicons:check',
    shape: 'rounded',
    iconPosition: 'left',
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
    <Button icon="heroicons:check" intent="success">
      Save
    </Button>

    <Button icon="heroicons:pencil-square" intent="info" variant="outline">
      Edit
    </Button>

    <Button icon="heroicons:trash" intent="error" variant="ghost">
      Delete
    </Button>
  </div>
</Story>

<Story name="With Icon Right" asChild>
  <div class="flex flex-wrap items-center gap-3">
    <Button icon="heroicons:arrow-right" iconPosition="right">
      Continue
    </Button>

    <Button
      icon="heroicons:arrow-top-right-on-square"
      iconPosition="right"
      intent="secondary"
      variant="outline"
    >
      Open
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
            <Button {intent} {variant} icon="heroicons:sparkles">
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
      <Button {size} icon="heroicons:check">
        {size}
      </Button>
    {/each}
  </div>
</Story>

<Story name="States" asChild>
  <div class="flex flex-wrap items-center gap-3">
    <Button icon="heroicons:check">Default</Button>

    <Button icon="heroicons:check" disabled>
      Disabled
    </Button>

    <Button icon="heroicons:check" loading>
      Loading
    </Button>

    <Button icon="heroicons:check" pressed>
      Pressed
    </Button>

    <Button icon="heroicons:trash" intent="error" variant="outline">
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
                <Button
                  {intent}
                  {variant}
                  {size}
                  icon="heroicons:check"
                >
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
      <Button {shape} icon="heroicons:check">
        {shape}
      </Button>
    {/each}
  </div>
</Story>
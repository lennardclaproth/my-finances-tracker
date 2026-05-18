<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import IconButton from './IconButton.svelte';

  import {
    buttonIntents,
    buttonVariants,
    buttonSizes
  } from './button.types';

  const { Story } = defineMeta({
    title: 'Atoms/IconButton',
    component: IconButton,
    tags: ['autodocs'],
    argTypes: {
      icon: {
        control: 'text'
      },
      ariaLabel: {
        control: 'text'
      },
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

<Story
  name="Playground"
  args={{
    icon: 'heroicons:trash',
    ariaLabel: 'Delete item',
    intent: 'secondary',
    variant: 'ghost',
    size: 'md',
    disabled: false,
    loading: false,
    pressed: false
  }}
/>

<Story name="Variants" asChild>
  <div class="flex flex-col gap-8">
    {#each buttonVariants as variant}
      <section class="space-y-3">
        <h3 class="text-sm font-semibold capitalize">{variant}</h3>

        <div class="flex flex-wrap gap-3">
          {#each buttonIntents as intent}
            <IconButton
              icon="heroicons:trash"
              ariaLabel={`${intent} icon button`}
              {intent}
              {variant}
            />
          {/each}
        </div>
      </section>
    {/each}
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="flex items-center gap-3">
    {#each buttonSizes as size}
      <IconButton
        icon="heroicons:trash"
        ariaLabel={`${size} icon button`}
        {size}
      />
    {/each}
  </div>
</Story>

<Story name="States" asChild>
  <div class="flex flex-wrap items-center gap-3">
    <IconButton
      icon="heroicons:trash"
      ariaLabel="Default icon button"
    />

    <IconButton
      icon="heroicons:trash"
      ariaLabel="Disabled icon button"
      disabled
    />

    <IconButton
      icon="heroicons:trash"
      ariaLabel="Loading icon button"
      loading
    />

    <IconButton
      icon="heroicons:trash"
      ariaLabel="Pressed icon button"
      pressed
    />
  </div>
</Story>

<Story name="Common Icons" asChild>
  <div class="flex flex-wrap items-center gap-3">
    <IconButton
      icon="heroicons:pencil-square"
      ariaLabel="Edit item"
      intent="info"
      variant="ghost"
    />

    <IconButton
      icon="heroicons:trash"
      ariaLabel="Delete item"
      intent="error"
      variant="ghost"
    />

    <IconButton
      icon="heroicons:check"
      ariaLabel="Confirm item"
      intent="success"
      variant="ghost"
    />

    <IconButton
      icon="heroicons:x-mark"
      ariaLabel="Close"
      intent="secondary"
      variant="ghost"
    />

    <IconButton
      icon="heroicons:cog-6-tooth"
      ariaLabel="Open settings"
      intent="secondary"
      variant="outline"
    />
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
                <IconButton
                  icon="heroicons:trash"
                  ariaLabel={`${intent} ${variant} ${size} icon button`}
                  {intent}
                  {variant}
                  {size}
                />
              {/each}
            </div>
          {/each}
        </div>
      </section>
    {/each}
  </div>
</Story>
<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import type { ComponentProps } from 'svelte';

  import Panel from './Panel.svelte';

  import Heading from '$lib/components/atoms/typography/Heading.svelte';
  import Text from '$lib/components/atoms/typography/Text.svelte';
  import Input from '$lib/components/atoms/input/Input.svelte';
  import Button from '$lib/components/atoms/button/Button.svelte';

  import {
    panelPaddings,
    panelShapes,
    panelShadows,
    panelVariants
  } from './panel.types';

  type PanelProps = ComponentProps<typeof Panel>;

  const { Story } = defineMeta({
    title: 'Atoms/Panel',
    component: Panel,
    tags: ['autodocs'],
    argTypes: {
      variant: {
        control: 'select',
        options: panelVariants
      },
      padding: {
        control: 'select',
        options: panelPaddings
      },
      shape: {
        control: 'select',
        options: panelShapes
      },
      shadow: {
        control: 'select',
        options: panelShadows
      },
      bordered: {
        control: 'boolean'
      },
      interactive: {
        control: 'boolean'
      }
    }
  });
</script>

{#snippet playground(args: PanelProps)}
  <Panel {...args}>
    <Heading level="h3" size="md">
      Panel title
    </Heading>

    <Text size="sm" tone="muted" class="mt-2">
      This is a reusable surface component. It can be used as a card,
      popover panel, or section container.
    </Text>
  </Panel>
{/snippet}

<Story
  name="Playground"
  args={{
    variant: 'default',
    padding: 'md',
    shape: 'md',
    shadow: 'sm',
    bordered: true,
    interactive: false
  }}
  template={playground}
/>

<Story name="Variants" asChild>
  <div class="grid max-w-3xl gap-4">
    {#each panelVariants as variant}
      <Panel
        {variant}
        shadow={variant === 'floating' ? 'md' : 'none'}
      >
        <Heading level="h3" size="md">
          {variant}
        </Heading>

        <Text size="sm" tone="default" class="mt-1">
          A {variant} panel surface.
        </Text>
      </Panel>
    {/each}
  </div>
</Story>

<Story name="Padding" asChild>
  <div class="grid max-w-3xl gap-4">
    {#each panelPaddings as padding}
      <Panel {padding}>
        <Text size="sm" tone="muted">
          Padding: {padding}
        </Text>
      </Panel>
    {/each}
  </div>
</Story>

<Story name="Shapes" asChild>
  <div class="grid max-w-3xl gap-4">
    {#each panelShapes as shape}
      <Panel {shape}>
        <Text size="sm" tone="muted">
          Shape: {shape}
        </Text>
      </Panel>
    {/each}
  </div>
</Story>

<Story name="Popover Example" asChild>
  <div class="max-w-sm">
    <Panel variant="floating" padding="md" shape="xl" shadow="md">
      <Heading level="h3" size="md" uppercase>
        Filters
      </Heading>

      <Text size="sm" tone="muted" class="mt-1">
        This is the kind of content that could live inside a popover.
      </Text>

      <div class="mt-4 space-y-3">
        <Input
          placeholder="Search..."
          leftIcon="heroicons:magnifying-glass"
        />

        <div class="flex justify-end gap-2">
          <Button
            intent="primary"
            variant="ghost"
            size="sm"
          >
            Reset
          </Button>

          <Button
            intent="primary"
            variant="solid"
            size="sm"
          >
            Apply
          </Button>
        </div>
      </div>
    </Panel>
  </div>
</Story>
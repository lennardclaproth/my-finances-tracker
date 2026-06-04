<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import type { ComponentProps } from 'svelte';

  import Chip from './Chip.svelte';

  import {
    badgeIntents,
    badgeShapes,
    badgeSizes,
    badgeVariants
  } from '$lib/components/atoms/badge/badge.types';

  type ChipProps = ComponentProps<typeof Chip>;

  const { Story } = defineMeta({
    title: 'Molecules/Chip',
    component: Chip,
    tags: ['autodocs'],
    argTypes: {
      intent: { control: 'select', options: badgeIntents },
      variant: { control: 'select', options: badgeVariants },
      size: { control: 'select', options: badgeSizes },
      shape: { control: 'select', options: badgeShapes },
      removable: { control: 'boolean' }
    }
  });
</script>

{#snippet playground(args: ChipProps)}
  <Chip {...args}>Groceries</Chip>
{/snippet}

<Story
  name="Playground"
  args={{ intent: 'neutral', variant: 'soft', size: 'md', shape: 'pill', removable: true }}
  template={playground}
/>

<Story name="Intents" asChild>
  <div class="flex flex-wrap items-center gap-2">
    {#each badgeIntents as intent}
      <Chip {intent}>{intent}</Chip>
    {/each}
  </div>
</Story>

<Story name="Category Filters" asChild>
  <div class="flex flex-wrap items-center gap-2">
    <Chip intent="secondary" removeLabel="Remove Groceries">Groceries</Chip>
    <Chip intent="info" removeLabel="Remove Rent">Rent</Chip>
    <Chip intent="warning" removeLabel="Remove Travel">Travel</Chip>
    <Chip intent="success" removeLabel="Remove Salary">Salary</Chip>
  </div>
</Story>

<Story name="Not Removable" asChild>
  <div class="flex flex-wrap items-center gap-2">
    <Chip removable={false} intent="primary">Read only</Chip>
  </div>
</Story>

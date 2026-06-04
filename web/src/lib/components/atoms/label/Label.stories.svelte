<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import type { ComponentProps } from 'svelte';

  import Label from './Label.svelte';

  import { labelSizes, labelTones } from './label.types';

  type LabelProps = ComponentProps<typeof Label>;

  const { Story } = defineMeta({
    title: 'Atoms/Label',
    component: Label,
    tags: ['autodocs'],
    argTypes: {
      size: { control: 'select', options: labelSizes },
      tone: { control: 'select', options: labelTones },
      required: { control: 'boolean' },
      disabled: { control: 'boolean' }
    }
  });
</script>

{#snippet playground(args: LabelProps)}
  <Label {...args}>Email address</Label>
{/snippet}

<Story
  name="Playground"
  args={{ size: 'md', tone: 'default', required: false, disabled: false }}
  template={playground}
/>

<Story name="Required" asChild>
  <Label required>Account name</Label>
</Story>

<Story name="Sizes" asChild>
  <div class="flex flex-col gap-2">
    {#each labelSizes as size}
      <Label {size}>Label {size}</Label>
    {/each}
  </div>
</Story>

<Story name="Disabled" asChild>
  <Label disabled>Disabled label</Label>
</Story>

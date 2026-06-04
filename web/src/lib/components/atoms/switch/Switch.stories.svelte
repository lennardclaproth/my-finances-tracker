<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Switch from './Switch.svelte';

  import { switchIntents, switchSizes } from './switch.types';

  const { Story } = defineMeta({
    title: 'Atoms/Switch',
    component: Switch,
    tags: ['autodocs'],
    args: { 'aria-label': 'Toggle setting' },
    argTypes: {
      checked: { control: 'boolean' },
      size: { control: 'select', options: switchSizes },
      intent: { control: 'select', options: switchIntents },
      disabled: { control: 'boolean' }
    }
  });
</script>

<Story
  name="Playground"
  args={{ checked: true, size: 'md', intent: 'secondary', disabled: false }}
/>

<Story name="Off And On" asChild>
  <div class="flex items-center gap-4">
    <Switch aria-label="Off" />
    <Switch aria-label="On" checked />
    <Switch aria-label="Disabled" disabled />
    <Switch aria-label="Disabled on" checked disabled />
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="flex items-center gap-4">
    {#each switchSizes as size}
      <Switch {size} checked aria-label={`${size} switch`} />
    {/each}
  </div>
</Story>

<Story name="Intents" asChild>
  <div class="flex items-center gap-4">
    {#each switchIntents as intent}
      <Switch {intent} checked aria-label={`${intent} switch`} />
    {/each}
  </div>
</Story>

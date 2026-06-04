<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Slider from './Slider.svelte';

  import { sliderIntents } from './slider.types';

  const { Story } = defineMeta({
    title: 'Atoms/Slider',
    component: Slider,
    tags: ['autodocs'],
    args: { ariaLabel: 'Amount' },
    argTypes: {
      value: { control: 'number' },
      min: { control: 'number' },
      max: { control: 'number' },
      step: { control: 'number' },
      intent: { control: 'select', options: sliderIntents },
      disabled: { control: 'boolean' }
    }
  });
</script>

<Story
  name="Playground"
  args={{ value: 40, min: 0, max: 100, step: 1, intent: 'secondary', disabled: false }}
/>

<Story name="Intents" asChild>
  <div class="flex max-w-md flex-col gap-4">
    {#each sliderIntents as intent}
      <Slider value={60} {intent} ariaLabel={`${intent} slider`} />
    {/each}
  </div>
</Story>

<Story name="Disabled" asChild>
  <div class="max-w-md">
    <Slider value={50} disabled ariaLabel="Disabled slider" />
  </div>
</Story>

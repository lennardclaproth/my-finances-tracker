<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import ProgressBar from './ProgressBar.svelte';

  import { progressBarIntents, progressBarSizes } from './progress-bar.types';

  const { Story } = defineMeta({
    title: 'Atoms/ProgressBar',
    component: ProgressBar,
    tags: ['autodocs'],
    args: { ariaLabel: 'Budget used' },
    argTypes: {
      value: { control: { type: 'range', min: 0, max: 100, step: 1 } },
      max: { control: 'number' },
      intent: { control: 'select', options: progressBarIntents },
      size: { control: 'select', options: progressBarSizes }
    }
  });
</script>

<Story name="Playground" args={{ value: 64, max: 100, intent: 'secondary', size: 'md' }} />

<Story name="Sizes" asChild>
  <div class="grid max-w-md gap-4">
    {#each progressBarSizes as size}
      <ProgressBar value={64} {size} ariaLabel={`${size} progress`} />
    {/each}
  </div>
</Story>

<Story name="Intents" asChild>
  <div class="grid max-w-md gap-4">
    {#each progressBarIntents as intent}
      <ProgressBar value={64} {intent} ariaLabel={`${intent} progress`} />
    {/each}
  </div>
</Story>

<Story name="Levels" asChild>
  <div class="grid max-w-md gap-4">
    <ProgressBar value={25} intent="success" ariaLabel="25 percent" />
    <ProgressBar value={70} intent="warning" ariaLabel="70 percent" />
    <ProgressBar value={95} intent="error" ariaLabel="95 percent" />
  </div>
</Story>

<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Spinner from './Spinner.svelte';

  import { spinnerSizes } from './spinner.types';

  const { Story } = defineMeta({
    title: 'Atoms/Spinner',
    component: Spinner,
    tags: ['autodocs'],
    argTypes: {
      size: { control: 'select', options: spinnerSizes },
      label: { control: 'text' }
    }
  });
</script>

<Story name="Playground" args={{ size: 'md', label: 'Loading' }} />

<Story name="Sizes" asChild>
  <div class="flex items-center gap-4 text-slate-700">
    {#each spinnerSizes as size}
      <Spinner {size} />
    {/each}
  </div>
</Story>

<Story name="Colors" asChild>
  <div class="flex items-center gap-4">
    <Spinner class="text-slate-700" />
    <Spinner class="text-emerald-600" />
    <Spinner class="text-red-600" />
    <Spinner class="text-sky-600" />
  </div>
</Story>

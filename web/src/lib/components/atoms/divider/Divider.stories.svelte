<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Divider from './Divider.svelte';

  import { dividerOrientations } from './divider.types';

  const { Story } = defineMeta({
    title: 'Atoms/Divider',
    component: Divider,
    tags: ['autodocs'],
    argTypes: {
      orientation: { control: 'select', options: dividerOrientations }
    }
  });
</script>

<Story name="Playground" args={{ orientation: 'horizontal' }} />

<Story name="Horizontal" asChild>
  <div class="max-w-md">
    <p class="text-sm text-slate-700">Above the divider</p>
    <Divider class="my-3" />
    <p class="text-sm text-slate-700">Below the divider</p>
  </div>
</Story>

<Story name="Vertical" asChild>
  <div class="flex h-8 items-center gap-3 text-sm text-slate-700">
    <span>Left</span>
    <Divider orientation="vertical" />
    <span>Right</span>
  </div>
</Story>

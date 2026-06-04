<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Skeleton from './Skeleton.svelte';

  import { skeletonVariants } from './skeleton.types';

  const { Story } = defineMeta({
    title: 'Atoms/Skeleton',
    component: Skeleton,
    tags: ['autodocs'],
    argTypes: {
      variant: { control: 'select', options: skeletonVariants },
      width: { control: 'text' },
      height: { control: 'text' }
    }
  });
</script>

<Story name="Playground" args={{ variant: 'text', width: '12rem' }} />

<Story name="Text Lines" asChild>
  <div class="flex max-w-md flex-col gap-2">
    <Skeleton width="60%" />
    <Skeleton />
    <Skeleton width="80%" />
  </div>
</Story>

<Story name="Shapes" asChild>
  <div class="flex items-center gap-4">
    <Skeleton variant="circle" width="3rem" height="3rem" />
    <Skeleton variant="rect" width="12rem" height="6rem" />
  </div>
</Story>

<Story name="List Row" asChild>
  <div class="flex max-w-md items-center gap-3">
    <Skeleton variant="circle" width="2.5rem" height="2.5rem" />
    <div class="flex flex-1 flex-col gap-2">
      <Skeleton width="50%" />
      <Skeleton width="30%" />
    </div>
    <Skeleton variant="rect" width="4rem" height="1.5rem" />
  </div>
</Story>

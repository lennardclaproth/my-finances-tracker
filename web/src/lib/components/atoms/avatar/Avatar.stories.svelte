<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Avatar from './Avatar.svelte';

  import { avatarShapes, avatarSizes } from './avatar.types';

  // Inline data URI so the image always loads without a network dependency.
  const sampleImage =
    "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='80' height='80'%3E%3Crect width='80' height='80' fill='%2310b981'/%3E%3C/svg%3E";

  const { Story } = defineMeta({
    title: 'Atoms/Avatar',
    component: Avatar,
    tags: ['autodocs'],
    argTypes: {
      src: { control: 'text' },
      alt: { control: 'text' },
      name: { control: 'text' },
      initials: { control: 'text' },
      size: { control: 'select', options: avatarSizes },
      shape: { control: 'select', options: avatarShapes }
    }
  });
</script>

<Story
  name="Playground"
  args={{ src: sampleImage, alt: 'Jane Doe', name: 'Jane Doe', size: 'md', shape: 'circle' }}
/>

<Story name="Initials Fallback" asChild>
  <div class="flex items-center gap-3">
    <Avatar name="Jane Doe" />
    <Avatar name="ING Bank" />
    <Avatar initials="N26" />
    <Avatar name="Single" />
  </div>
</Story>

<Story name="Image" asChild>
  <div class="flex items-center gap-3">
    <Avatar src={sampleImage} alt="Account avatar" />
    <Avatar src="https://invalid.example/missing.png" name="Falls Back" />
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="flex items-center gap-3">
    {#each avatarSizes as size}
      <Avatar {size} name="Jane Doe" />
    {/each}
  </div>
</Story>

<Story name="Shapes" asChild>
  <div class="flex items-center gap-3">
    {#each avatarShapes as shape}
      <Avatar {shape} src={sampleImage} alt="Avatar" />
    {/each}
  </div>
</Story>

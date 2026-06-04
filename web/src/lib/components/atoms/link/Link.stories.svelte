<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import type { ComponentProps } from 'svelte';

  import Link from './Link.svelte';

  import { linkSizes, linkTones, linkUnderlines } from './link.types';

  type LinkProps = ComponentProps<typeof Link>;

  const { Story } = defineMeta({
    title: 'Atoms/Link',
    component: Link,
    tags: ['autodocs'],
    args: { href: '#' },
    argTypes: {
      href: { control: 'text' },
      tone: { control: 'select', options: linkTones },
      size: { control: 'select', options: linkSizes },
      underline: { control: 'select', options: linkUnderlines },
      external: { control: 'boolean' }
    }
  });
</script>

{#snippet playground(args: LinkProps)}
  <Link {...args}>View transactions</Link>
{/snippet}

<Story
  name="Playground"
  args={{ href: '#', tone: 'primary', size: 'md', underline: 'hover', external: false }}
  template={playground}
/>

<Story name="Tones" asChild>
  <div class="flex items-center gap-4">
    {#each linkTones as tone}
      <Link href="#" {tone}>{tone} link</Link>
    {/each}
  </div>
</Story>

<Story name="Underlines" asChild>
  <div class="flex items-center gap-4">
    {#each linkUnderlines as underline}
      <Link href="#" {underline}>{underline}</Link>
    {/each}
  </div>
</Story>

<Story name="External" asChild>
  <Link href="https://svelte.dev" external>Open svelte.dev</Link>
</Story>

<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import type { ComponentProps } from 'svelte';

  import Heading from './Heading.svelte';
  import Text from './Text.svelte';

  import {
    headingLevels,
    headingSizes,
    headingTones,
    headingWeights,
    textSizes,
    textTones,
    textWeights
  } from './typography.types';

  type HeadingProps = ComponentProps<typeof Heading>;

  const { Story } = defineMeta({
    title: 'Atoms/Typography',
    tags: ['autodocs'],
    argTypes: {
      level: {
        control: 'select',
        options: headingLevels
      },
      size: {
        control: 'select',
        options: headingSizes
      },
      tone: {
        control: 'select',
        options: headingTones
      },
      weight: {
        control: 'select',
        options: headingWeights
      },
      uppercase: {
        control: 'boolean'
      }
    }
  });
</script>

{#snippet headingPlayground(args: HeadingProps)}
  <Heading {...args}>
    The quick brown fox jumps over the lazy dog
  </Heading>
{/snippet}

<Story
  name="Heading Playground"
  args={{
    level: 'h2',
    size: 'lg',
    tone: 'default',
    weight: 'semibold',
    uppercase: false
  }}
  template={headingPlayground}
/>

<Story name="Heading Uppercase" asChild>
  <div class="max-w-3xl space-y-6">
    <div>
      <Text size="sm" tone="muted" class="mb-2">
        Normal
      </Text>

      <Heading level="h2" size="lg">
        Account summary
      </Heading>
    </div>

    <div>
      <Text size="sm" tone="muted" class="mb-2">
        Uppercase
      </Text>

      <Heading level="h2" size="lg" uppercase>
        Account summary
      </Heading>
    </div>
  </div>
</Story>

<Story name="Headings" asChild>
  <div class="max-w-3xl space-y-6">
    {#each headingSizes as size}
      <div>
        <Text size="sm" tone="muted" class="mb-2">
          Heading size: {size}
        </Text>

        <Heading level="h2" {size}>
          The quick brown fox jumps over the lazy dog
        </Heading>
      </div>
    {/each}
  </div>
</Story>

<Story name="Heading Levels" asChild>
  <div class="max-w-3xl space-y-6">
    {#each headingLevels as level}
      <div>
        <Text size="sm" tone="muted" class="mb-2">
          {level}
        </Text>

        <Heading {level} size="xl">
          The quick brown fox jumps over the lazy dog
        </Heading>
      </div>
    {/each}
  </div>
</Story>

<Story name="Heading Tones" asChild>
  <div class="max-w-3xl space-y-6">
    {#each headingTones as tone}
      <Heading level="h2" size="xl" {tone}>
        {tone} heading
      </Heading>
    {/each}
  </div>
</Story>

<Story name="Heading Weights" asChild>
  <div class="max-w-3xl space-y-6">
    {#each headingWeights as weight}
      <Heading level="h2" size="xl" {weight}>
        {weight} heading
      </Heading>
    {/each}
  </div>
</Story>

<Story name="Text Sizes" asChild>
  <div class="max-w-2xl space-y-4">
    {#each textSizes as size}
      <Text {size}>
        This is {size} body text. The quick brown fox jumps over the lazy dog.
      </Text>
    {/each}
  </div>
</Story>

<Story name="Text Tones" asChild>
  <div class="max-w-2xl space-y-4">
    {#each textTones as tone}
      <Text {tone}>
        This is {tone} text. The quick brown fox jumps over the lazy dog.
      </Text>
    {/each}
  </div>
</Story>

<Story name="Text Weights" asChild>
  <div class="max-w-2xl space-y-4">
    {#each textWeights as weight}
      <Text {weight}>
        This is {weight} text. The quick brown fox jumps over the lazy dog.
      </Text>
    {/each}
  </div>
</Story>

<Story name="Composition" asChild>
  <div class="max-w-2xl space-y-3">
    <Heading level="h1" size="2xl">
      Account summary
    </Heading>

    <Text tone="muted">
      This section gives you a quick overview of your current financial state.
    </Text>

    <Text size="sm" tone="subtle">
      Last updated two minutes ago.
    </Text>
  </div>
</Story>
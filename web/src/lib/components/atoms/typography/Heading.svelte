<script lang="ts">
  import type { Snippet } from 'svelte';

  import type {
    HeadingLevel,
    HeadingSize,
    HeadingTone,
    HeadingWeight
  } from './typography.types';

  import {
    headingBaseClasses,
    headingSizeClasses,
    headingToneClasses,
    headingUppercaseClasses,
    headingWeightClasses
  } from './typography.variants';

  type Props = {
    level?: HeadingLevel;
    size?: HeadingSize;
    tone?: HeadingTone;
    weight?: HeadingWeight;
    uppercase?: boolean;
    class?: string;
    children?: Snippet;
  };

  let {
    level = 'h2',
    size = 'lg',
    tone = 'default',
    weight = 'semibold',
    uppercase = false,
    class: className = '',
    children
  }: Props = $props();

  const classes = $derived([
    headingBaseClasses,
    headingSizeClasses[size],
    headingToneClasses[tone],
    headingWeightClasses[weight],
    uppercase ? headingUppercaseClasses : '',
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<svelte:element this={level} class={classes}>
  {@render children?.()}
</svelte:element>
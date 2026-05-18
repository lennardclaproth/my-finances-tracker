<script lang="ts">
  import type { Snippet } from 'svelte';

  import type {
    TextElement,
    TextSize,
    TextTone,
    TextWeight
  } from './typography.types';

  import {
    textBaseClasses,
    textSizeClasses,
    textToneClasses,
    textWeightClasses
  } from './typography.variants';

  type Props = {
    as?: TextElement;
    size?: TextSize;
    tone?: TextTone;
    weight?: TextWeight;
    class?: string;
    children?: Snippet;
  };

  let {
    as = 'p',
    size = 'md',
    tone = 'default',
    weight = 'normal',
    class: className = '',
    children
  }: Props = $props();

  const classes = $derived([
    textBaseClasses,
    textSizeClasses[size],
    textToneClasses[tone],
    textWeightClasses[weight],
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<svelte:element this={as} class={classes}>
  {@render children?.()}
</svelte:element>
<script lang="ts">
  import type { Snippet } from 'svelte';

  import type {
    ButtonIntent,
    ButtonShape,
    ButtonSize,
    ButtonType,
    ButtonVariant
  } from './button.types';

  import {
    baseButtonClasses,
    buttonShapeClasses,
    buttonSizeClasses,
    iconSizeClasses,
    intentVariantClasses
  } from './button.variants';

  type Props = {
    intent?: ButtonIntent;
    variant?: ButtonVariant;
    size?: ButtonSize;
    shape?: ButtonShape;
    type?: ButtonType;
    disabled?: boolean;
    loading?: boolean;
    pressed?: boolean;
    ariaLabel?: string;
    class?: string;
    onclick?: (event: MouseEvent) => void;
    children?: Snippet;
  };

  let {
    intent = 'primary',
    variant = 'solid',
    size = 'md',
    shape = 'rounded',
    type = 'button',
    disabled = false,
    loading = false,
    pressed = false,
    ariaLabel,
    class: className = '',
    onclick,
    children
  }: Props = $props();

  const spinnerClasses = $derived([
    'animate-spin rounded-full border-2 border-current border-t-transparent',
    iconSizeClasses[size]
  ].join(' '));

  const classes = $derived([
    baseButtonClasses,
    buttonSizeClasses[size],
    buttonShapeClasses[shape],
    intentVariantClasses[intent][variant],
    loading ? 'cursor-wait' : '',
    className
  ]
    .filter(Boolean)
    .join(' '));

  function handleClick(event: MouseEvent) {
    if (disabled || loading) return;
    onclick?.(event);
  }
</script>

<button
  class={classes}
  {type}
  {disabled}
  aria-label={ariaLabel}
  aria-busy={loading}
  aria-pressed={pressed || undefined}
  onclick={handleClick}
>
  {#if loading}
    <span class={spinnerClasses} aria-hidden="true"></span>
  {/if}

  {@render children?.()}
</button>

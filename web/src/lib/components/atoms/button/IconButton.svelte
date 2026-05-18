<script lang="ts">
  import Icon from '$lib/components/atoms/icon/Icon.svelte';

  import type {
    ButtonIntent,
    ButtonSize,
    ButtonType,
    ButtonVariant
  } from './button.types';

  import {
    baseButtonClasses,
    iconButtonSizeClasses,
    iconSizeClasses,
    intentVariantClasses
  } from './button.variants';

  type Props = {
    icon: string;
    ariaLabel: string;
    intent?: ButtonIntent;
    variant?: ButtonVariant;
    size?: ButtonSize;
    type?: ButtonType;
    disabled?: boolean;
    loading?: boolean;
    pressed?: boolean;
    class?: string;
    onclick?: (event: MouseEvent) => void;
  };

  let {
    icon,
    ariaLabel,
    intent = 'secondary',
    variant = 'ghost',
    size = 'md',
    type = 'button',
    disabled = false,
    loading = false,
    pressed = false,
    class: className = '',
    onclick
  }: Props = $props();

  const spinnerClasses = $derived([
    'animate-spin rounded-full border-2 border-current border-t-transparent',
    iconSizeClasses[size]
  ].join(' '));

  const classes = $derived([
    baseButtonClasses,
    iconButtonSizeClasses[size],
    'rounded-full',
    intentVariantClasses[intent][variant],
    pressed ? 'ring-2 ring-offset-2' : '',
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
  {:else}
    <Icon {icon} {size} />
  {/if}
</button>
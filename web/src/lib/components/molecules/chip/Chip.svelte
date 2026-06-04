<script lang="ts">
  import type { Snippet } from 'svelte';

  import Icon from '$lib/components/atoms/icon/Icon.svelte';

  // Chip reuses the Badge atom's visual tokens (data) and adds a remove affordance.
  import type {
    BadgeIntent,
    BadgeShape,
    BadgeSize,
    BadgeVariant
  } from '$lib/components/atoms/badge/badge.types';

  import {
    badgeIntentVariantClasses,
    badgeShapeClasses,
    badgeSizeClasses,
    baseBadgeClasses
  } from '$lib/components/atoms/badge/badge.variants';

  type Props = {
    intent?: BadgeIntent;
    variant?: BadgeVariant;
    size?: BadgeSize;
    shape?: BadgeShape;
    removable?: boolean;
    onRemove?: () => void;
    /** Accessible label for the remove button, e.g. "Remove Groceries". */
    removeLabel?: string;
    class?: string;
    children?: Snippet;
  };

  let {
    intent = 'neutral',
    variant = 'soft',
    size = 'md',
    shape = 'pill',
    removable = true,
    onRemove,
    removeLabel = 'Remove',
    class: className = '',
    children
  }: Props = $props();

  const classes = $derived([
    baseBadgeClasses,
    badgeSizeClasses[size],
    badgeShapeClasses[shape],
    badgeIntentVariantClasses[intent][variant],
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<span class={classes}>
  {@render children?.()}

  {#if removable}
    <button
      type="button"
      class="-mr-0.5 ml-0.5 inline-flex items-center justify-center rounded-full opacity-70 transition-opacity hover:opacity-100 focus-visible:ring-2 focus-visible:ring-current focus-visible:outline-none"
      aria-label={removeLabel}
      onclick={onRemove}
    >
      <Icon icon="heroicons:x-mark" size="sm" />
    </button>
  {/if}
</span>

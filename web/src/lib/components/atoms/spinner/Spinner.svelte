<script lang="ts">
  import type { SpinnerSize } from './spinner.types';

  type Props = {
    size?: SpinnerSize;
    /** Accessible status text announced to screen readers. */
    label?: string;
    class?: string;
  };

  let { size = 'md', label = 'Loading', class: className = '' }: Props = $props();

  const sizeClasses = {
    sm: 'size-4 border-2',
    md: 'size-6 border-2',
    lg: 'size-8 border-[3px]'
  } satisfies Record<SpinnerSize, string>;

  // Uses currentColor, so the caller controls the colour via a text-* class.
  const classes = $derived([
    'inline-block animate-spin rounded-full border-current border-t-transparent',
    sizeClasses[size],
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<span role="status" class="inline-flex">
  <span class={classes} aria-hidden="true"></span>
  <span class="sr-only">{label}</span>
</span>

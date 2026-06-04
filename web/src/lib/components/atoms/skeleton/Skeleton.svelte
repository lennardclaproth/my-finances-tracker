<script lang="ts">
  import type { SkeletonVariant } from './skeleton.types';

  type Props = {
    variant?: SkeletonVariant;
    /** CSS width, e.g. '100%' or '4rem'. */
    width?: string;
    /** CSS height, e.g. '1rem' or '160px'. */
    height?: string;
    class?: string;
  };

  let { variant = 'text', width, height, class: className = '' }: Props = $props();

  const variantClasses = {
    text: 'h-4 w-full rounded',
    rect: 'rounded-md',
    circle: 'rounded-full'
  } satisfies Record<SkeletonVariant, string>;

  const classes = $derived([
    'block animate-pulse bg-slate-200',
    variantClasses[variant],
    className
  ]
    .filter(Boolean)
    .join(' '));

  const style = $derived(
    [width ? `width:${width}` : '', height ? `height:${height}` : ''].filter(Boolean).join(';')
  );
</script>

<div class={classes} {style} aria-hidden="true"></div>

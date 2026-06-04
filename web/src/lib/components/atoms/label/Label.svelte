<script lang="ts">
  import type { Snippet } from 'svelte';

  import type { LabelSize, LabelTone } from './label.types';

  type Props = {
    /** Id of the form control this label is bound to. */
    for?: string;
    size?: LabelSize;
    tone?: LabelTone;
    required?: boolean;
    disabled?: boolean;
    class?: string;
    children?: Snippet;
  };

  let {
    for: forId,
    size = 'md',
    tone = 'default',
    required = false,
    disabled = false,
    class: className = '',
    children
  }: Props = $props();

  const sizeClasses = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-base'
  } satisfies Record<LabelSize, string>;

  const toneClasses = {
    default: 'text-slate-800',
    muted: 'text-slate-500'
  } satisfies Record<LabelTone, string>;

  const classes = $derived([
    'inline-flex select-none items-center gap-1 font-medium',
    sizeClasses[size],
    disabled ? 'opacity-50' : toneClasses[tone],
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<label for={forId} class={classes}>
  {@render children?.()}
  {#if required}
    <span class="text-red-600" aria-hidden="true">*</span>
  {/if}
</label>

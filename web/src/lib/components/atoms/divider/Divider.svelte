<script lang="ts">
  import type { DividerOrientation } from './divider.types';

  type Props = {
    orientation?: DividerOrientation;
    class?: string;
  };

  let { orientation = 'horizontal', class: className = '' }: Props = $props();

  const classes = $derived([
    'border-slate-200',
    orientation === 'vertical' ? 'inline-block h-full border-l' : 'w-full border-t',
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<div role="separator" aria-orientation={orientation} class={classes}></div>

<script lang="ts">
  import type { ProgressBarIntent, ProgressBarSize } from './progress-bar.types';

  import { progressBarFillClasses, progressBarSizeClasses } from './progress-bar.variants';

  type Props = {
    /** Current value. */
    value: number;
    /** Maximum value the bar represents. */
    max?: number;
    intent?: ProgressBarIntent;
    size?: ProgressBarSize;
    /** Accessible name for the progress bar. */
    ariaLabel?: string;
    class?: string;
  };

  let {
    value,
    max = 100,
    intent = 'secondary',
    size = 'md',
    ariaLabel,
    class: className = ''
  }: Props = $props();

  const percent = $derived(max <= 0 ? 0 : Math.min(100, Math.max(0, (value / max) * 100)));

  const trackClasses = $derived([
    'w-full overflow-hidden rounded-full bg-slate-200',
    progressBarSizeClasses[size],
    className
  ]
    .filter(Boolean)
    .join(' '));

  const fillClasses = $derived([
    'h-full rounded-full transition-[width] duration-300 ease-out',
    progressBarFillClasses[intent]
  ].join(' '));
</script>

<div
  class={trackClasses}
  role="progressbar"
  aria-label={ariaLabel}
  aria-valuenow={Math.round(value)}
  aria-valuemin={0}
  aria-valuemax={max}
>
  <div class={fillClasses} style={`width: ${percent}%`}></div>
</div>

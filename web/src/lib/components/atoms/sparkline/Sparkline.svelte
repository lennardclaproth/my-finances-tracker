<script lang="ts">
  import type { ResolvedSparklineTone, SparklineTone } from './sparkline.types';

  import { sparklineToneClasses } from './sparkline.variants';

  type Props = {
    /** Series of values, oldest to newest. */
    data: number[];
    width?: number;
    height?: number;
    strokeWidth?: number;
    /** 'auto' colours by trend (last vs first); otherwise a fixed tone. */
    tone?: SparklineTone;
    /** Render a subtle area fill under the line. */
    fill?: boolean;
    /** Accessible label; when omitted the chart is treated as decorative. */
    ariaLabel?: string;
    class?: string;
  };

  let {
    data,
    width = 100,
    height = 32,
    strokeWidth = 2,
    tone = 'auto',
    fill = false,
    ariaLabel,
    class: className = ''
  }: Props = $props();

  const pad = $derived(strokeWidth);

  const points = $derived(computePoints(data, width, height, pad));

  function computePoints(
    values: number[],
    w: number,
    h: number,
    p: number
  ): Array<{ x: number; y: number }> {
    if (values.length === 0) return [];
    const min = Math.min(...values);
    const max = Math.max(...values);
    const range = max - min || 1;
    const innerW = w - p * 2;
    const innerH = h - p * 2;
    const step = values.length > 1 ? innerW / (values.length - 1) : 0;
    return values.map((value, i) => ({
      x: p + i * step,
      y: p + innerH - ((value - min) / range) * innerH
    }));
  }

  const linePath = $derived(
    points.map((pt, i) => `${i === 0 ? 'M' : 'L'}${pt.x.toFixed(2)},${pt.y.toFixed(2)}`).join(' ')
  );

  const areaPath = $derived(
    points.length
      ? `${linePath} L${points[points.length - 1].x.toFixed(2)},${(height - pad).toFixed(2)}` +
          ` L${points[0].x.toFixed(2)},${(height - pad).toFixed(2)} Z`
      : ''
  );

  const resolvedTone: ResolvedSparklineTone = $derived(
    tone !== 'auto'
      ? tone
      : data.length > 1 && data[data.length - 1] >= data[0]
        ? 'positive'
        : 'negative'
  );

  const colorClass = $derived(sparklineToneClasses[resolvedTone]);

  const classes = $derived(['inline-block', className].filter(Boolean).join(' '));
</script>

<svg
  class={classes}
  {width}
  {height}
  viewBox={`0 0 ${width} ${height}`}
  fill="none"
  role={ariaLabel ? 'img' : undefined}
  aria-label={ariaLabel}
  aria-hidden={ariaLabel ? undefined : 'true'}
>
  {#if fill && areaPath}
    <path d={areaPath} class={colorClass} fill="currentColor" fill-opacity="0.12" stroke="none" />
  {/if}

  <path
    d={linePath}
    class={colorClass}
    stroke="currentColor"
    stroke-width={strokeWidth}
    stroke-linecap="round"
    stroke-linejoin="round"
  />
</svg>

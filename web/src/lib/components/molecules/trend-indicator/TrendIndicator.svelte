<script lang="ts">
  import Icon from '$lib/components/atoms/icon/Icon.svelte';

  import type { TrendIndicatorFormat, TrendIndicatorSize } from './trend-indicator.types';

  import {
    trendDirectionClasses,
    trendDirectionIcons,
    trendIconSizeClasses,
    trendTextSizeClasses
  } from './trend-indicator.variants';

  type Props = {
    /** The change value. Sign drives direction and colour. */
    value: number;
    /** How to format the value. `percent` treats e.g. 2.34 as 2.34%. */
    format?: TrendIndicatorFormat;
    currency?: string;
    locale?: string;
    size?: TrendIndicatorSize;
    showArrow?: boolean;
    class?: string;
  };

  let {
    value,
    format = 'percent',
    currency = 'EUR',
    locale = 'en-US',
    size = 'md',
    showArrow = true,
    class: className = ''
  }: Props = $props();

  const direction = $derived(value > 0 ? 'up' : value < 0 ? 'down' : 'flat');

  const formatted = $derived(formatValue(value, format, currency, locale));

  function formatValue(v: number, f: TrendIndicatorFormat, c: string, l: string): string {
    if (f === 'percent') {
      return new Intl.NumberFormat(l, {
        style: 'percent',
        signDisplay: 'always',
        minimumFractionDigits: 1,
        maximumFractionDigits: 2
      }).format(v / 100);
    }
    if (f === 'currency') {
      return new Intl.NumberFormat(l, {
        style: 'currency',
        currency: c,
        signDisplay: 'always'
      }).format(v);
    }
    return new Intl.NumberFormat(l, { signDisplay: 'always', maximumFractionDigits: 2 }).format(v);
  }

  const classes = $derived([
    'inline-flex items-center gap-0.5 font-medium tabular-nums',
    trendTextSizeClasses[size],
    trendDirectionClasses[direction],
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<span class={classes}>
  {#if showArrow}
    <Icon icon={trendDirectionIcons[direction]} size={trendIconSizeClasses[size]} />
  {/if}
  {formatted}
</span>

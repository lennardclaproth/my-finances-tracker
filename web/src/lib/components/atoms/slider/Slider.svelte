<script lang="ts">
  import type { SliderIntent } from './slider.types';

  import { sliderAccentClasses } from './slider.variants';

  type Props = {
    value?: number;
    min?: number;
    max?: number;
    step?: number;
    intent?: SliderIntent;
    disabled?: boolean;
    ariaLabel?: string;
    class?: string;
    oninput?: (event: Event) => void;
    onchange?: (event: Event) => void;
  };

  let {
    value = $bindable(0),
    min = 0,
    max = 100,
    step = 1,
    intent = 'secondary',
    disabled = false,
    ariaLabel,
    class: className = '',
    oninput,
    onchange
  }: Props = $props();

  const classes = $derived([
    'h-2 w-full cursor-pointer',
    sliderAccentClasses[intent],
    'disabled:cursor-not-allowed disabled:opacity-50',
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<input
  type="range"
  bind:value
  {min}
  {max}
  {step}
  {disabled}
  aria-label={ariaLabel}
  class={classes}
  {oninput}
  {onchange}
/>

<script lang="ts">
  // Field styling tokens are shared from the Input atom (data, not a component import).
  import {
    baseInputClasses,
    inputIntentClasses,
    inputShapeClasses,
    inputSizeClasses
  } from '$lib/components/atoms/input/input.variants';

  import type {
    CurrencyInputIntent,
    CurrencyInputShape,
    CurrencyInputSize
  } from './currency-input.types';

  import {
    currencySymbolContainerClasses,
    currencySymbolPaddingClasses
  } from './currency-input.variants';

  type Props = {
    id?: string;
    name?: string;
    value?: string;
    /** Currency symbol shown as a leading affix. */
    currencySymbol?: string;
    placeholder?: string;
    size?: CurrencyInputSize;
    intent?: CurrencyInputIntent;
    shape?: CurrencyInputShape;
    disabled?: boolean;
    readonly?: boolean;
    required?: boolean;
    ariaLabel?: string;
    ariaDescribedby?: string;
    class?: string;
    oninput?: (event: Event) => void;
    onchange?: (event: Event) => void;
    onblur?: (event: FocusEvent) => void;
    onfocus?: (event: FocusEvent) => void;
  };

  let {
    id,
    name,
    value = $bindable(''),
    currencySymbol = '€',
    placeholder = '0.00',
    size = 'md',
    intent = 'default',
    shape = 'default',
    disabled = false,
    readonly = false,
    required = false,
    ariaLabel,
    ariaDescribedby,
    class: className = '',
    oninput,
    onchange,
    onblur,
    onfocus
  }: Props = $props();

  const classes = $derived([
    baseInputClasses,
    inputSizeClasses[size],
    inputShapeClasses[shape],
    inputIntentClasses[intent],
    'text-right tabular-nums',
    currencySymbolPaddingClasses[size],
    className
  ]
    .filter(Boolean)
    .join(' '));

  const symbolClasses = $derived([
    'pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center text-slate-500',
    currencySymbolContainerClasses[size]
  ].join(' '));
</script>

<div class="relative w-full">
  <span class={symbolClasses} aria-hidden="true">{currencySymbol}</span>

  <input
    {id}
    {name}
    type="text"
    inputmode="decimal"
    bind:value
    {placeholder}
    class={classes}
    {disabled}
    readonly={readonly}
    {required}
    aria-label={ariaLabel}
    aria-invalid={intent === 'error' ? 'true' : undefined}
    aria-describedby={ariaDescribedby}
    {oninput}
    {onchange}
    {onblur}
    {onfocus}
  />
</div>

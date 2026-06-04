<script lang="ts">
  // Field styling tokens are shared from the Input atom (data, not a component import).
  import {
    baseInputClasses,
    inputIntentClasses,
    inputShapeClasses,
    inputSizeClasses
  } from '$lib/components/atoms/input/input.variants';

  import type {
    SelectIntent,
    SelectOption,
    SelectShape,
    SelectSize
  } from './select.types';

  import {
    selectChevronContainerClasses,
    selectChevronPaddingClasses
  } from './select.variants';

  type Props = {
    id?: string;
    name?: string;
    value?: string;
    options?: SelectOption[];
    placeholder?: string;
    size?: SelectSize;
    intent?: SelectIntent;
    shape?: SelectShape;
    disabled?: boolean;
    required?: boolean;
    ariaLabel?: string;
    ariaDescribedby?: string;
    class?: string;
    onchange?: (event: Event) => void;
    onblur?: (event: FocusEvent) => void;
    onfocus?: (event: FocusEvent) => void;
  };

  let {
    id,
    name,
    value = $bindable(''),
    options = [],
    placeholder,
    size = 'md',
    intent = 'default',
    shape = 'default',
    disabled = false,
    required = false,
    ariaLabel,
    ariaDescribedby,
    class: className = '',
    onchange,
    onblur,
    onfocus
  }: Props = $props();

  const classes = $derived([
    baseInputClasses,
    inputSizeClasses[size],
    inputShapeClasses[shape],
    inputIntentClasses[intent],
    'cursor-pointer appearance-none',
    selectChevronPaddingClasses[size],
    className
  ]
    .filter(Boolean)
    .join(' '));

  const chevronClasses = $derived([
    'pointer-events-none absolute inset-y-0 right-0 flex items-center justify-center text-slate-400',
    selectChevronContainerClasses[size]
  ].join(' '));
</script>

<div class="relative w-full">
  <select
    {id}
    {name}
    bind:value
    class={classes}
    {disabled}
    {required}
    aria-label={ariaLabel}
    aria-invalid={intent === 'error' ? 'true' : undefined}
    aria-describedby={ariaDescribedby}
    {onchange}
    {onblur}
    {onfocus}
  >
    {#if placeholder}
      <option value="" disabled>{placeholder}</option>
    {/if}

    {#each options as option (option.value)}
      <option value={option.value} disabled={option.disabled}>
        {option.label}
      </option>
    {/each}
  </select>

  <span class={chevronClasses} aria-hidden="true">
    <svg
      class="size-4"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      stroke-width="1.75"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path d="M6 8l4 4 4-4" />
    </svg>
  </span>
</div>

<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements';

  import type { RadioSize } from './radio.types';

  interface Props
    extends Omit<
      HTMLInputAttributes,
      'type' | 'checked' | 'size' | 'class' | 'disabled' | 'value'
    > {
    /**
     * Shared selected value for the radio group. Supports bind:group.
     */
    group?: string;

    /**
     * Value contributed to the group when this radio is selected.
     */
    value?: string;

    /**
     * Displays an invalid/error state.
     */
    invalid?: boolean;

    /**
     * Prevents user interaction.
     */
    disabled?: boolean;

    /**
     * Controls the visual size of the radio.
     */
    size?: RadioSize;

    /**
     * Additional classes applied to the wrapper.
     */
    class?: string;
  }

  let {
    group = $bindable(undefined),
    value,
    invalid = false,
    disabled = false,
    size = 'md',
    class: className = '',
    ...rest
  }: Props = $props();

  const selected = $derived(group === value);

  const sizeClasses = {
    sm: 'size-4',
    md: 'size-5',
    lg: 'size-6'
  } satisfies Record<RadioSize, string>;

  const dotSizeClasses = {
    sm: 'size-1.5',
    md: 'size-2',
    lg: 'size-2.5'
  } satisfies Record<RadioSize, string>;

  const baseClasses = [
    'pointer-events-none inline-flex items-center justify-center rounded-full',
    'border bg-white transition-all duration-150 ease-out',
    'peer-focus-visible:ring-2 peer-focus-visible:ring-offset-2',
    'peer-disabled:opacity-50'
  ].join(' ');

  const ringClasses = $derived(
    invalid
      ? 'border-red-600 peer-hover:bg-red-50 peer-focus-visible:ring-red-50 peer-checked:border-red-600'
      : 'border-slate-400 peer-hover:bg-emerald-50 peer-focus-visible:ring-emerald-50 peer-checked:border-slate-800'
  );

  // The dot is a descendant of the indicator, so its visibility is driven by `selected`
  // (Svelte state) rather than the peer-checked selector, which only reaches siblings.
  const dotClasses = $derived([
    'rounded-full transition-transform duration-150 ease-out',
    invalid ? 'bg-red-600' : 'bg-emerald-500',
    dotSizeClasses[size],
    selected ? 'scale-100' : 'scale-0'
  ].join(' '));
</script>

<span
  class={[
    'relative inline-flex shrink-0 align-middle',
    disabled ? 'cursor-not-allowed' : 'cursor-pointer',
    className
  ]
    .filter(Boolean)
    .join(' ')}
>
  <input {...rest} type="radio" bind:group {value} {disabled} class="peer sr-only" />

  <span aria-hidden="true" class={[baseClasses, ringClasses, sizeClasses[size]].filter(Boolean).join(' ')}>
    <span class={dotClasses}></span>
  </span>
</span>

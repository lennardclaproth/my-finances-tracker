<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements';

  import type { SwitchIntent, SwitchSize } from './switch.types';

  import {
    switchOnTrackClasses,
    switchThumbSizeClasses,
    switchThumbTranslateClasses,
    switchTrackSizeClasses
  } from './switch.variants';

  interface Props
    extends Omit<
      HTMLInputAttributes,
      'type' | 'checked' | 'size' | 'class' | 'disabled' | 'role'
    > {
    /**
     * Whether the switch is on. Supports bind:checked.
     */
    checked?: boolean;

    /**
     * Prevents user interaction.
     */
    disabled?: boolean;

    /**
     * Controls the visual size of the switch.
     */
    size?: SwitchSize;

    /**
     * Track colour when the switch is on.
     */
    intent?: SwitchIntent;

    /**
     * Additional classes applied to the wrapper.
     */
    class?: string;
  }

  let {
    checked = $bindable(false),
    disabled = false,
    size = 'md',
    intent = 'secondary',
    class: className = '',
    ...rest
  }: Props = $props();

  const trackClasses = $derived([
    'relative inline-flex shrink-0 items-center rounded-full',
    'transition-colors duration-150 ease-out',
    'peer-focus-visible:ring-2 peer-focus-visible:ring-slate-300 peer-focus-visible:ring-offset-2',
    'peer-disabled:opacity-50',
    switchTrackSizeClasses[size],
    checked ? switchOnTrackClasses[intent] : 'bg-slate-300'
  ].join(' '));

  // Thumb is a descendant of the track, so its travel is driven by `checked` (Svelte state)
  // rather than peer-checked, which only reaches siblings.
  const thumbClasses = $derived([
    'pointer-events-none ml-0.5 inline-block rounded-full bg-white shadow-sm',
    'transition-transform duration-150 ease-out',
    switchThumbSizeClasses[size],
    checked ? switchThumbTranslateClasses[size] : 'translate-x-0'
  ].join(' '));
</script>

<label
  class={[
    'inline-flex items-center align-middle',
    disabled ? 'cursor-not-allowed' : 'cursor-pointer',
    className
  ]
    .filter(Boolean)
    .join(' ')}
>
  <input {...rest} type="checkbox" role="switch" bind:checked {disabled} class="peer sr-only" />

  <span class={trackClasses}>
    <span class={thumbClasses}></span>
  </span>
</label>

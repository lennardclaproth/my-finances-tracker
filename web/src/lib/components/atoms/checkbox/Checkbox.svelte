<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements';

  import type { CheckboxSize } from './checkbox.types';

  interface Props
    extends Omit<
      HTMLInputAttributes,
      'type' | 'checked' | 'size' | 'class' | 'disabled'
    > {
    /**
     * Whether the checkbox is selected.
     * Supports bind:checked.
     */
    checked?: boolean;

    /**
     * Whether the checkbox is partially selected.
     * Supports bind:indeterminate.
     */
    indeterminate?: boolean;

    /**
     * Displays an invalid/error state.
     */
    invalid?: boolean;

    /**
     * Prevents user interaction.
     */
    disabled?: boolean;

    /**
     * Controls the visual size of the checkbox.
     */
    size?: CheckboxSize;

    /**
     * Additional classes applied to the wrapper.
     */
    class?: string;
  }

  let {
    checked = $bindable(false),
    indeterminate = $bindable(false),
    invalid = false,
    disabled = false,
    size = 'md',
    class: className = '',
    ...rest
  }: Props = $props();

  const sizeClasses = {
    sm: 'size-4 rounded',
    md: 'size-5 rounded-md',
    lg: 'size-6 rounded-md'
  } satisfies Record<CheckboxSize, string>;

  const iconSizeClasses = {
    sm: 'size-3',
    md: 'size-3.5',
    lg: 'size-4'
  } satisfies Record<CheckboxSize, string>;

  const baseClasses = [
    'pointer-events-none inline-flex items-center justify-center',
    'border text-slate-800',
    'transition-all duration-150 ease-out',
    'peer-focus-visible:ring-2 peer-focus-visible:ring-offset-2',
    'peer-disabled:opacity-50'
  ].join(' ');

  const secondaryClasses = [
    // Matches secondary.outline when unchecked
    'border-slate-800 bg-white',
    'peer-hover:bg-emerald-100',
    'peer-focus-visible:ring-emerald-50 peer-focus-visible:bg-emerald-200',

    // Matches secondary.solid when checked or indeterminate
    'peer-checked:border-slate-800 peer-checked:bg-emerald-400',
    'peer-indeterminate:border-slate-800 peer-indeterminate:bg-emerald-400',

    // Checked / indeterminate interaction state
    'peer-checked:peer-hover:bg-emerald-500',
    'peer-indeterminate:peer-hover:bg-emerald-500',
    'peer-checked:peer-focus-visible:bg-emerald-400',
    'peer-indeterminate:peer-focus-visible:bg-emerald-400'
  ].join(' ');

  const invalidClasses = [
    'border-red-600 bg-white text-red-700',
    'peer-hover:bg-red-100',
    'peer-focus-visible:ring-red-50 peer-focus-visible:bg-red-200',
    'peer-checked:border-red-600 peer-checked:bg-red-700 peer-checked:text-white',
    'peer-indeterminate:border-red-600 peer-indeterminate:bg-red-700 peer-indeterminate:text-white',
    'peer-checked:peer-hover:bg-red-600',
    'peer-indeterminate:peer-hover:bg-red-600'
  ].join(' ');

  const disabledClasses = [
    'peer-disabled:pointer-events-none',
    'peer-disabled:border-slate-300',
    'peer-disabled:bg-slate-100',
    'peer-disabled:text-slate-500'
  ].join(' ');
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
  <input
    {...rest}
    type="checkbox"
    bind:checked
    bind:indeterminate
    {disabled}
    aria-invalid={invalid || undefined}
    class="peer absolute inset-0 z-10 m-0 size-full cursor-pointer appearance-none opacity-0 disabled:cursor-not-allowed"
  />

  <span
    aria-hidden="true"
    class={[
      baseClasses,
      invalid ? invalidClasses : secondaryClasses,
      disabledClasses,
      sizeClasses[size]
    ]
      .filter(Boolean)
      .join(' ')}
  >
    {#if indeterminate}
      <svg
        class={iconSizeClasses[size]}
        viewBox="0 0 20 20"
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        stroke-linecap="round"
      >
        <path d="M4 10h12" />
      </svg>
    {:else if checked}
      <svg
        class={iconSizeClasses[size]}
        viewBox="0 0 20 20"
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M4 10.5 8 14l8-8" />
      </svg>
    {/if}
  </span>
</span>
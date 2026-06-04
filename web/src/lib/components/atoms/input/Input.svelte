<script lang="ts">
  import type { HTMLInputAttributes } from 'svelte/elements';

  import type {
    InputIntent,
    InputSize,
    InputType,
    InputShape
  } from './input.types';

  import {
    baseInputClasses,
    inputIntentClasses,
    inputShapeClasses,
    inputSizeClasses
  } from './input.variants';

  type Props = {
    id?: string;
    name?: string;
    type?: InputType;
    value?: string;
    placeholder?: string;

    size?: InputSize;
    intent?: InputIntent;
    shape?: InputShape;

    disabled?: boolean;
    readonly?: boolean;
    required?: boolean;

    autocomplete?: HTMLInputAttributes['autocomplete'];
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
    type = 'text',
    value = $bindable(''),
    placeholder,

    size = 'md',
    intent = 'default',
    shape = 'default',

    disabled = false,
    readonly = false,
    required = false,

    autocomplete,
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
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<input
  {id}
  {name}
  {type}
  bind:value
  {placeholder}
  class={classes}
  {disabled}
  readonly={readonly}
  {required}
  autocomplete={autocomplete}
  aria-label={ariaLabel}
  aria-invalid={intent === 'error' ? 'true' : undefined}
  aria-describedby={ariaDescribedby}
  oninput={oninput}
  onchange={onchange}
  onblur={onblur}
  onfocus={onfocus}
/>

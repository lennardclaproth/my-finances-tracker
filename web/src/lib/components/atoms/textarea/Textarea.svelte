<script lang="ts">
  // Field styling tokens are shared from the Input atom (data, not a component import).
  import {
    baseInputClasses,
    inputIntentClasses,
    inputShapeClasses
  } from '$lib/components/atoms/input/input.variants';

  import type {
    TextareaIntent,
    TextareaResize,
    TextareaShape,
    TextareaSize
  } from './textarea.types';

  import { textareaResizeClasses, textareaSizeClasses } from './textarea.variants';

  type Props = {
    id?: string;
    name?: string;
    value?: string;
    placeholder?: string;
    rows?: number;
    size?: TextareaSize;
    intent?: TextareaIntent;
    shape?: TextareaShape;
    resize?: TextareaResize;
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
    placeholder,
    rows = 4,
    size = 'md',
    intent = 'default',
    shape = 'default',
    resize = 'vertical',
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
    textareaSizeClasses[size],
    inputShapeClasses[shape],
    inputIntentClasses[intent],
    textareaResizeClasses[resize],
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<textarea
  {id}
  {name}
  bind:value
  {placeholder}
  {rows}
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
></textarea>

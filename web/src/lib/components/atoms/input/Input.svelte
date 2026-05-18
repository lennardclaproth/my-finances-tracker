<script lang="ts">
  import Icon from '$lib/components/atoms/icon/Icon.svelte';
  import type { HTMLInputAttributes } from 'svelte/elements';

  import type {
    InputIntent,
    InputSize,
    InputType,
    InputShape,
  } from './input.types';

  import {
    baseInputClasses,
    inputIconContainerSizeClasses,
    inputIconIntentClasses,
    inputIconPaddingClasses,
    inputIconSizeClasses,
    inputIntentClasses,
    inputShapeClasses,
    inputSizeClasses,
    inputValidationIconIntentClasses,
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

    leftIcon?: string;
    rightIcon?: string;

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

    leftIcon,
    rightIcon,

    autocomplete,
    ariaLabel,
    ariaDescribedby,

    class: className = '',

    oninput,
    onchange,
    onblur,
    onfocus
  }: Props = $props();

  const validationIcon = $derived(
    intent === 'success'
      ? 'heroicons:check-circle'
      : intent === 'error'
        ? 'heroicons:exclamation-circle'
        : undefined
  );

  const resolvedRightIcon = $derived(rightIcon || validationIcon);

  const isValidationIcon = $derived(!rightIcon && !!validationIcon);

  const classes = $derived([
    baseInputClasses,
    inputSizeClasses[size],
    inputShapeClasses[shape],
    inputIntentClasses[intent],
    leftIcon ? inputIconPaddingClasses[size].left : '',
    resolvedRightIcon ? inputIconPaddingClasses[size].right : '',
    className
  ]
    .filter(Boolean)
    .join(' '));

  const leftIconClasses = $derived([
    'pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center',
    inputIconContainerSizeClasses[size],
    inputIconIntentClasses.default
  ].join(' '));

  const rightIconClasses = $derived([
    'pointer-events-none absolute inset-y-0 right-0 flex items-center justify-center',
    inputIconContainerSizeClasses[size],
    isValidationIcon
      ? inputValidationIconIntentClasses[intent]
      : inputIconIntentClasses.default
  ].join(' '));
</script>

<div class="group relative w-full">
  {#if leftIcon}
    <span class={leftIconClasses} aria-hidden="true">
      <Icon icon={leftIcon} size={inputIconSizeClasses[size]} />
    </span>
  {/if}

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

  {#if resolvedRightIcon}
    <span class={rightIconClasses} aria-hidden="true">
      <Icon icon={resolvedRightIcon} size={inputIconSizeClasses[size]} />
    </span>
  {/if}
</div>
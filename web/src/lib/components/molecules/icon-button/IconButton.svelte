<script lang="ts">
  import Button from '$lib/components/atoms/button/Button.svelte';
  import Icon from '$lib/components/atoms/icon/Icon.svelte';

  import type {
    ButtonIntent,
    ButtonSize,
    ButtonType,
    ButtonVariant
  } from '$lib/components/atoms/button/button.types';

  type Props = {
    icon: string;
    ariaLabel: string;
    intent?: ButtonIntent;
    variant?: ButtonVariant;
    size?: ButtonSize;
    type?: ButtonType;
    disabled?: boolean;
    loading?: boolean;
    pressed?: boolean;
    class?: string;
    onclick?: (event: MouseEvent) => void;
  };

  let {
    icon,
    ariaLabel,
    intent = 'secondary',
    variant = 'ghost',
    size = 'md',
    type = 'button',
    disabled = false,
    loading = false,
    pressed = false,
    class: className = '',
    onclick
  }: Props = $props();

  // Square sizing: the Button atom fixes the height per size, so we only add the matching width.
  const widthClasses = {
    sm: 'w-8',
    md: 'w-10',
    lg: 'w-12'
  } satisfies Record<ButtonSize, string>;

  const classes = $derived([
    widthClasses[size],
    pressed ? 'ring-2 ring-offset-2' : '',
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<Button
  {intent}
  {variant}
  {size}
  {type}
  shape="pill"
  {disabled}
  {loading}
  {pressed}
  {ariaLabel}
  class={classes}
  {onclick}
>
  {#if !loading}
    <Icon {icon} {size} />
  {/if}
</Button>

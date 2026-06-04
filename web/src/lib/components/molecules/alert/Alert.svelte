<script lang="ts">
  import type { Snippet } from 'svelte';

  import Icon from '$lib/components/atoms/icon/Icon.svelte';
  import IconButton from '$lib/components/molecules/icon-button/IconButton.svelte';

  import type { AlertIntent } from './alert.types';

  import { alertContainerClasses, alertIconColorClasses, alertIcons } from './alert.variants';

  type Props = {
    intent?: AlertIntent;
    title?: string;
    dismissible?: boolean;
    onDismiss?: () => void;
    class?: string;
    /** Alert body / message. */
    children?: Snippet;
  };

  let {
    intent = 'info',
    title,
    dismissible = false,
    onDismiss,
    class: className = '',
    children
  }: Props = $props();

  const classes = $derived([
    'flex items-start gap-3 rounded-lg border p-3',
    alertContainerClasses[intent],
    className
  ]
    .filter(Boolean)
    .join(' '));

  // Errors/warnings interrupt (assertive); info/success announce politely.
  const role = $derived(intent === 'error' || intent === 'warning' ? 'alert' : 'status');
</script>

<div class={classes} {role}>
  <span class={['mt-0.5 shrink-0', alertIconColorClasses[intent]].join(' ')}>
    <Icon icon={alertIcons[intent]} size="md" />
  </span>

  <div class="min-w-0 flex-1">
    {#if title}
      <p class="text-sm font-semibold">{title}</p>
    {/if}
    {#if children}
      <div class="text-sm">{@render children()}</div>
    {/if}
  </div>

  {#if dismissible}
    <IconButton
      icon="heroicons:x-mark"
      ariaLabel="Dismiss"
      size="sm"
      variant="ghost"
      intent="secondary"
      class="-mt-1 -mr-1 shrink-0"
      onclick={onDismiss}
    />
  {/if}
</div>

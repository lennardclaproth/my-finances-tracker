<script lang="ts">
  import Icon from '@iconify/svelte';

  type Props = {
    icon: string;
    size?: 'sm' | 'md' | 'lg';
    class?: string;
    ariaHidden?: boolean;
  };

  let {
    icon,
    size = 'md',
    class: className = '',
    ariaHidden = true
  }: Props = $props();

  const sizeClasses = {
    sm: 'size-3',
    md: 'size-4',
    lg: 'size-5'
  };

  const classes = $derived([
    'shrink-0',
    sizeClasses[size],
    className
  ].filter(Boolean).join(' '));
</script>

<Icon
  {icon}
  class={classes}
  aria-hidden={ariaHidden}
/>
<script lang="ts">
  import type { Snippet } from 'svelte';

  import type { LinkSize, LinkTone, LinkUnderline } from './link.types';

  import {
    linkBaseClasses,
    linkSizeClasses,
    linkToneClasses,
    linkUnderlineClasses
  } from './link.variants';

  type Props = {
    href: string;
    tone?: LinkTone;
    size?: LinkSize;
    underline?: LinkUnderline;
    /** Open in a new tab with safe rel attributes. */
    external?: boolean;
    class?: string;
    onclick?: (event: MouseEvent) => void;
    children?: Snippet;
  };

  let {
    href,
    tone = 'default',
    size = 'md',
    underline = 'hover',
    external = false,
    class: className = '',
    onclick,
    children
  }: Props = $props();

  const classes = $derived([
    linkBaseClasses,
    linkToneClasses[tone],
    linkSizeClasses[size],
    linkUnderlineClasses[underline],
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<a
  {href}
  class={classes}
  target={external ? '_blank' : undefined}
  rel={external ? 'noopener noreferrer' : undefined}
  {onclick}
>
  {@render children?.()}
</a>

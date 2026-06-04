<script lang="ts">
  import type { AvatarShape, AvatarSize } from './avatar.types';

  import { avatarShapeClasses, avatarSizeClasses } from './avatar.variants';

  type Props = {
    /** Image source. When absent or it fails to load, initials are shown. */
    src?: string;
    /** Alternative text for the image / accessible name for the initials fallback. */
    alt?: string;
    /** Full name used to derive initials when `initials` is not given. */
    name?: string;
    /** Explicit initials override. */
    initials?: string;
    size?: AvatarSize;
    shape?: AvatarShape;
    class?: string;
  };

  let {
    src,
    alt = '',
    name,
    initials,
    size = 'md',
    shape = 'circle',
    class: className = ''
  }: Props = $props();

  let failed = $state(false);

  const label = $derived(alt || name || '');

  const computedInitials = $derived((initials ?? deriveInitials(name ?? alt)).slice(0, 2).toUpperCase());

  function deriveInitials(value: string): string {
    const parts = value.trim().split(/\s+/).filter(Boolean);
    if (parts.length === 0) return '';
    if (parts.length === 1) return parts[0].slice(0, 2);
    return parts[0][0] + parts[parts.length - 1][0];
  }

  const showImage = $derived(!!src && !failed);

  const classes = $derived([
    'inline-flex shrink-0 select-none items-center justify-center overflow-hidden bg-slate-200 font-medium text-slate-700',
    avatarSizeClasses[size],
    avatarShapeClasses[shape],
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<span
  class={classes}
  role={!showImage && label ? 'img' : undefined}
  aria-label={!showImage && label ? label : undefined}
>
  {#if showImage}
    <img {src} {alt} class="h-full w-full object-cover" onerror={() => (failed = true)} />
  {:else}
    <span aria-hidden={label ? 'true' : undefined}>{computedInitials}</span>
  {/if}
</span>

<script lang="ts">
  import type { FileInputSize } from './file-input.types';

  import { fileInputSizeClasses } from './file-input.variants';

  type Props = {
    id?: string;
    name?: string;
    /** Comma-separated list of accepted types, e.g. '.csv,.xml'. */
    accept?: string;
    multiple?: boolean;
    disabled?: boolean;
    size?: FileInputSize;
    /** Selected files. Supports bind:files. */
    files?: FileList | null;
    ariaLabel?: string;
    class?: string;
    onchange?: (event: Event) => void;
  };

  let {
    id,
    name,
    accept,
    multiple = false,
    disabled = false,
    size = 'md',
    files = $bindable(null),
    ariaLabel,
    class: className = '',
    onchange
  }: Props = $props();

  const baseClasses = [
    'block w-full text-slate-600',
    'file:mr-3 file:cursor-pointer file:rounded-md file:border-0 file:font-medium',
    'file:bg-emerald-400 file:text-slate-800 hover:file:bg-emerald-500',
    'disabled:cursor-not-allowed disabled:opacity-50'
  ].join(' ');

  const classes = $derived([baseClasses, fileInputSizeClasses[size], className]
    .filter(Boolean)
    .join(' '));
</script>

<input
  type="file"
  bind:files
  {id}
  {name}
  {accept}
  {multiple}
  {disabled}
  aria-label={ariaLabel}
  class={classes}
  {onchange}
/>

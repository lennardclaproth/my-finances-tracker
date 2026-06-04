<script lang="ts">
  import Icon from '$lib/components/atoms/icon/Icon.svelte';

  type Props = {
    /** Comma-separated accepted types, e.g. '.csv,.xml'. */
    accept?: string;
    multiple?: boolean;
    disabled?: boolean;
    /** Selected files. Supports bind:files. */
    files?: FileList | null;
    label?: string;
    hint?: string;
    class?: string;
    onchange?: (event: Event) => void;
  };

  let {
    accept,
    multiple = false,
    disabled = false,
    files = $bindable(null),
    label = 'Drag & drop files here, or click to browse',
    hint,
    class: className = '',
    onchange
  }: Props = $props();

  let dragging = $state(false);

  function handleDragOver(event: DragEvent) {
    event.preventDefault();
    if (!disabled) dragging = true;
  }

  function handleDragLeave() {
    dragging = false;
  }

  function handleDrop(event: DragEvent) {
    event.preventDefault();
    dragging = false;
    if (disabled) return;
    if (event.dataTransfer?.files && event.dataTransfer.files.length > 0) {
      files = event.dataTransfer.files;
    }
  }

  const classes = $derived([
    'flex flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed p-6 text-center transition-colors',
    disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer',
    dragging
      ? 'border-emerald-400 bg-emerald-50'
      : 'border-slate-300 bg-slate-50 hover:border-slate-400',
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<label class={classes} ondragover={handleDragOver} ondragleave={handleDragLeave} ondrop={handleDrop}>
  <span class="text-slate-400">
    <Icon icon="heroicons:arrow-up-tray" size="lg" />
  </span>

  <span class="text-sm font-medium text-slate-700">{label}</span>

  {#if hint}
    <span class="text-xs text-slate-500">{hint}</span>
  {/if}

  <input type="file" bind:files {accept} {multiple} {disabled} class="sr-only" {onchange} />
</label>

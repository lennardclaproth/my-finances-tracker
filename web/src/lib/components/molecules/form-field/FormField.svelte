<script lang="ts">
  import type { Snippet } from 'svelte';

  import Label from '$lib/components/atoms/label/Label.svelte';
  import Text from '$lib/components/atoms/typography/Text.svelte';
  import type { LabelSize } from '$lib/components/atoms/label/label.types';

  type FieldContext = {
    id: string | undefined;
    describedby: string | undefined;
    invalid: boolean;
  };

  type Props = {
    label?: string;
    /** Id wired to the control; also derives the hint/error ids for aria-describedby. */
    id?: string;
    hint?: string;
    error?: string;
    required?: boolean;
    size?: LabelSize;
    class?: string;
    /** Renders the control, receiving the field context to spread onto it. */
    children?: Snippet<[FieldContext]>;
  };

  let {
    label,
    id,
    hint,
    error,
    required = false,
    size = 'md',
    class: className = '',
    children
  }: Props = $props();

  const invalid = $derived(!!error);
  const hintId = $derived(id ? `${id}-hint` : undefined);
  const errorId = $derived(id ? `${id}-error` : undefined);
  const describedby = $derived(error ? errorId : hint ? hintId : undefined);

  const context = $derived({ id, describedby, invalid });
</script>

<div class={['flex flex-col gap-1.5', className].filter(Boolean).join(' ')}>
  {#if label}
    <Label for={id} {size} {required}>{label}</Label>
  {/if}

  {@render children?.(context)}

  {#if error}
    <div id={errorId} role="alert">
      <Text as="span" size="sm" tone="danger">{error}</Text>
    </div>
  {:else if hint}
    <div id={hintId}>
      <Text as="span" size="sm" tone="muted">{hint}</Text>
    </div>
  {/if}
</div>

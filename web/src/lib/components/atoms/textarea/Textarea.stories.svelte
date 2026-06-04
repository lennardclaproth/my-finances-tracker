<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Textarea from './Textarea.svelte';

  import {
    textareaIntents,
    textareaResizes,
    textareaSizes
  } from './textarea.types';

  const { Story } = defineMeta({
    title: 'Atoms/Textarea',
    component: Textarea,
    tags: ['autodocs'],
    args: { ariaLabel: 'Notes' },
    argTypes: {
      value: { control: 'text' },
      placeholder: { control: 'text' },
      rows: { control: 'number' },
      size: { control: 'select', options: textareaSizes },
      intent: { control: 'select', options: textareaIntents },
      resize: { control: 'select', options: textareaResizes },
      disabled: { control: 'boolean' },
      readonly: { control: 'boolean' },
      required: { control: 'boolean' }
    }
  });
</script>

<Story
  name="Playground"
  args={{
    value: '',
    placeholder: 'Add a note for this transaction…',
    rows: 4,
    size: 'md',
    intent: 'default',
    resize: 'vertical',
    disabled: false,
    readonly: false,
    required: false
  }}
/>

<Story name="Sizes" asChild>
  <div class="grid max-w-md gap-4">
    {#each textareaSizes as size}
      <Textarea {size} ariaLabel={`${size} textarea`} placeholder={`${size} textarea`} />
    {/each}
  </div>
</Story>

<Story name="Intents" asChild>
  <div class="grid max-w-md gap-4">
    {#each textareaIntents as intent}
      <Textarea {intent} ariaLabel={`${intent} textarea`} placeholder={`${intent} textarea`} />
    {/each}
  </div>
</Story>

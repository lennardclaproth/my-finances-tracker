<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import FileInput from './FileInput.svelte';

  import { fileInputSizes } from './file-input.types';

  const { Story } = defineMeta({
    title: 'Atoms/FileInput',
    component: FileInput,
    tags: ['autodocs'],
    args: { ariaLabel: 'Upload statement' },
    argTypes: {
      accept: { control: 'text' },
      multiple: { control: 'boolean' },
      size: { control: 'select', options: fileInputSizes },
      disabled: { control: 'boolean' }
    }
  });
</script>

<Story
  name="Playground"
  args={{ accept: '.csv,.xml', multiple: false, size: 'md', disabled: false }}
/>

<Story name="Statement Import" asChild>
  <div class="max-w-md">
    <FileInput accept=".csv,.xml" multiple ariaLabel="Import bank statements" />
  </div>
</Story>

<Story name="Sizes" asChild>
  <div class="flex max-w-md flex-col gap-4">
    {#each fileInputSizes as size}
      <FileInput {size} ariaLabel={`${size} file input`} />
    {/each}
  </div>
</Story>

<Story name="Disabled" asChild>
  <div class="max-w-md">
    <FileInput disabled ariaLabel="Disabled file input" />
  </div>
</Story>

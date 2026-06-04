<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Dropzone from './Dropzone.svelte';

  const { Story } = defineMeta({
    title: 'Molecules/Dropzone',
    component: Dropzone,
    tags: ['autodocs'],
    argTypes: {
      accept: { control: 'text' },
      multiple: { control: 'boolean' },
      disabled: { control: 'boolean' },
      label: { control: 'text' },
      hint: { control: 'text' }
    }
  });
</script>

<Story
  name="Playground"
  args={{
    accept: '.csv,.xml',
    multiple: true,
    disabled: false,
    label: 'Drag & drop files here, or click to browse',
    hint: 'CSV or XML, up to 10MB'
  }}
/>

<Story name="Statement Import" asChild>
  <div class="max-w-md">
    <Dropzone
      accept=".csv,.xml"
      multiple
      label="Drop bank statements here"
      hint="DEGIRO, ING, N26 or Brand New Day exports"
    />
  </div>
</Story>

<Story name="Disabled" asChild>
  <div class="max-w-md">
    <Dropzone disabled label="Uploads are disabled" />
  </div>
</Story>

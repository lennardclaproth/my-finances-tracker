<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import Checkbox from './Checkbox.svelte';
  import { checkboxSizes } from './checkbox.types';

  const { Story } = defineMeta({
    title: 'Atoms/Checkbox',
    component: Checkbox,
    tags: ['autodocs'],
    argTypes: {
      checked: {
        control: 'boolean'
      },
      indeterminate: {
        control: 'boolean'
      },
      invalid: {
        control: 'boolean'
      },
      disabled: {
        control: 'boolean'
      },
      size: {
        control: 'select',
        options: checkboxSizes
      }
    }
  });
</script>

<script lang="ts">
  let checked = $state(false);
  let indeterminate = $state(true);
</script>

<Story
  name="Playground"
  args={{
    checked: false,
    indeterminate: false,
    invalid: false,
    disabled: false,
    size: 'md',
    'aria-label': 'Checkbox'
  }}
/>

<Story
  name="Checked"
  args={{
    checked: true,
    size: 'md',
    'aria-label': 'Checked checkbox'
  }}
/>

<Story
  name="Indeterminate"
  args={{
    indeterminate: true,
    size: 'md',
    'aria-label': 'Partially selected checkbox'
  }}
/>

<Story
  name="Invalid"
  args={{
    invalid: true,
    size: 'md',
    'aria-label': 'Invalid checkbox'
  }}
/>

<Story
  name="Disabled"
  args={{
    disabled: true,
    size: 'md',
    'aria-label': 'Disabled checkbox'
  }}
/>

<Story name="Interactive States">
  {#snippet template()}
    <div class="flex flex-col gap-4">
      <label class="inline-flex cursor-pointer items-center gap-3">
        <Checkbox bind:checked bind:indeterminate />

        <span class="text-sm text-slate-700">
          Select all items
        </span>
      </label>

      <div class="flex gap-2 text-sm text-slate-500">
        <span>Checked: {checked ? 'true' : 'false'}</span>
        <span>·</span>
        <span>Indeterminate: {indeterminate ? 'true' : 'false'}</span>
      </div>

      <button
        type="button"
        class="w-fit rounded-md border border-slate-300 px-3 py-1.5 text-sm text-slate-700 hover:bg-slate-50"
        onclick={() => {
          checked = false;
          indeterminate = true;
        }}
      >
        Reset to indeterminate
      </button>
    </div>
  {/snippet}
</Story>
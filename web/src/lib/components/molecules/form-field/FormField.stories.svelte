<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';

  import FormField from './FormField.svelte';
  import Input from '$lib/components/atoms/input/Input.svelte';
  import Select from '$lib/components/atoms/select/Select.svelte';
  import CurrencyInput from '$lib/components/atoms/currency-input/CurrencyInput.svelte';

  const { Story } = defineMeta({
    title: 'Molecules/FormField',
    component: FormField,
    tags: ['autodocs']
  });
</script>

<Story name="With Hint" asChild>
  <div class="max-w-md">
    <FormField id="email" label="Email address" hint="We'll never share it." required>
      {#snippet children({ id, describedby, invalid })}
        <Input
          {id}
          type="email"
          ariaDescribedby={describedby}
          intent={invalid ? 'error' : 'default'}
          placeholder="you@example.com"
        />
      {/snippet}
    </FormField>
  </div>
</Story>

<Story name="With Error" asChild>
  <div class="max-w-md">
    <FormField id="amount" label="Amount" error="Amount is required.">
      {#snippet children({ id, describedby, invalid })}
        <CurrencyInput
          {id}
          ariaDescribedby={describedby}
          intent={invalid ? 'error' : 'default'}
        />
      {/snippet}
    </FormField>
  </div>
</Story>

<Story name="With Select" asChild>
  <div class="max-w-md">
    <FormField id="account" label="Account">
      {#snippet children({ id, describedby })}
        <Select
          {id}
          ariaDescribedby={describedby}
          placeholder="Select an account"
          options={[
            { value: 'checking', label: 'Checking' },
            { value: 'savings', label: 'Savings' }
          ]}
        />
      {/snippet}
    </FormField>
  </div>
</Story>

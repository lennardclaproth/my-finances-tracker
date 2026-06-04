<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import type { ComponentProps } from 'svelte';

  import Alert from './Alert.svelte';

  import { alertIntents } from './alert.types';

  type AlertProps = ComponentProps<typeof Alert>;

  const { Story } = defineMeta({
    title: 'Molecules/Alert',
    component: Alert,
    tags: ['autodocs'],
    argTypes: {
      intent: { control: 'select', options: alertIntents },
      title: { control: 'text' },
      dismissible: { control: 'boolean' }
    }
  });
</script>

{#snippet playground(args: AlertProps)}
  <Alert {...args}>Your import completed with 3 skipped rows.</Alert>
{/snippet}

<Story
  name="Playground"
  args={{ intent: 'info', title: 'Heads up', dismissible: true }}
  template={playground}
/>

<Story name="Intents" asChild>
  <div class="flex max-w-lg flex-col gap-3">
    {#each alertIntents as intent}
      <Alert {intent} title={`${intent} alert`}>
        This is a {intent} message describing what happened.
      </Alert>
    {/each}
  </div>
</Story>

<Story name="Dismissible" asChild>
  <div class="max-w-lg">
    <Alert intent="success" title="Statement imported" dismissible onDismiss={() => {}}>
      42 transactions were added to your Checking account.
    </Alert>
  </div>
</Story>

<Story name="Without Title" asChild>
  <div class="max-w-lg">
    <Alert intent="error">Could not parse the uploaded file. Check the format and try again.</Alert>
  </div>
</Story>

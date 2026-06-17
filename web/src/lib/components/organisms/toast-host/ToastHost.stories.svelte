<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import ToastHost from './ToastHost.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/ToastHost',
		component: ToastHost,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	import { onMount } from 'svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { toast } from '$lib/stores/toast.svelte';

	// Isolate each story from the shared singleton.
	onMount(() => {
		toast.clear();
		return () => toast.clear();
	});

	function stack() {
		toast.info('Import scheduled', { title: 'Background job' });
		toast.success('Transaction saved');
		toast.warning('Balance is low');
		toast.error('Sync failed — retrying');
	}
</script>

<Story name="Tones" asChild>
	<div class="flex flex-wrap gap-2 p-2">
		<Button intent="info" onclick={() => toast.info('Heads up — something happened.')}>Info</Button>
		<Button intent="success" onclick={() => toast.success('Saved successfully.')}>Success</Button>
		<Button intent="warning" onclick={() => toast.warning('Double-check this.')}>Warning</Button>
		<Button intent="error" onclick={() => toast.error('That did not work.')}>Error</Button>
	</div>
	<ToastHost />
</Story>

<Story name="Stacked" asChild>
	<div class="flex flex-wrap gap-2 p-2">
		<Button onclick={stack}>Trigger 4 toasts</Button>
		<Button variant="ghost" intent="secondary" onclick={() => toast.clear()}>Clear all</Button>
	</div>
	<ToastHost />
</Story>

<Story name="Auto-dismiss vs sticky" asChild>
	<div class="flex flex-wrap gap-2 p-2">
		<Button onclick={() => toast.success('Auto-dismisses in 4.5s')}>Auto-dismiss</Button>
		<Button
			intent="warning"
			onclick={() => toast.warning('Stays until dismissed', { duration: 0, title: 'Sticky' })}
		>
			Sticky (duration 0)
		</Button>
	</div>
	<ToastHost />
</Story>

<Story name="With title + message" asChild>
	<div class="flex flex-wrap gap-2 p-2">
		<Button
			intent="info"
			onclick={() =>
				toast.fromStatus('import.completed', 'Imported 128 transactions from ING.', {
					title: 'Import complete'
				})}
		>
			From status
		</Button>
	</div>
	<ToastHost />
</Story>

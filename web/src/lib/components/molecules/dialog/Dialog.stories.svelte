<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Dialog from './Dialog.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { dialogSizes, type DialogSize } from './dialog.types';

	const { Story } = defineMeta({
		title: 'Molecules/Dialog',
		component: Dialog,
		tags: ['autodocs'],
		argTypes: {
			size: { control: 'select', options: dialogSizes },
			dismissible: { control: 'boolean' },
			closeOnBackdrop: { control: 'boolean' }
		}
	});
</script>

<script lang="ts">
	let open = $state(false);
	let openWithFooter = $state(false);
	let openSize = $state<DialogSize | null>(null);
</script>

<Story name="Playground" asChild>
	<div>
		<Button onclick={() => (open = true)}>Open dialog</Button>
		<Dialog bind:open title="Dialog title">
			<p>
				This is a native <code>&lt;dialog&gt;</code>. Press <kbd>Esc</kbd>, click the backdrop, or
				use the close button to dismiss it. Focus is trapped while it is open.
			</p>
		</Dialog>
	</div>
</Story>

<Story name="With footer actions" asChild>
	<div>
		<Button onclick={() => (openWithFooter = true)}>Delete item…</Button>
		<Dialog bind:open={openWithFooter} size="sm" title="Delete transaction?">
			<p>This action cannot be undone. The transaction will be permanently removed.</p>
			{#snippet footer()}
				<Button variant="ghost" intent="secondary" onclick={() => (openWithFooter = false)}>
					Cancel
				</Button>
				<Button intent="error" onclick={() => (openWithFooter = false)}>Delete</Button>
			{/snippet}
		</Dialog>
	</div>
</Story>

<Story name="Sizes" asChild>
	<div class="flex flex-wrap gap-2">
		{#each dialogSizes as size (size)}
			<Button variant="outline" onclick={() => (openSize = size)}>{size}</Button>
		{/each}

		<Dialog
			open={openSize !== null}
			size={openSize ?? 'md'}
			title={`Size: ${openSize ?? ''}`}
			onClose={() => (openSize = null)}
		>
			<p>The dialog grows with the <code>size</code> prop ({openSize}).</p>
		</Dialog>
	</div>
</Story>

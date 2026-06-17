<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import Drawer from './Drawer.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/Drawer',
		component: Drawer,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	let open = $state(false);
</script>

<Story name="Right slide-in" asChild>
	<div>
		<Button onclick={() => (open = true)}>Open drawer</Button>
		<Drawer bind:open title="Details">
			<p class="text-sm text-slate-700">
				A right-hand slide-in panel. Press <kbd>Esc</kbd>, click the backdrop, or the close button
				to dismiss it.
			</p>
			{#snippet footer()}
				<Button variant="ghost" intent="secondary" onclick={() => (open = false)}>Close</Button>
				<Button onclick={() => (open = false)}>Save</Button>
			{/snippet}
		</Drawer>
	</div>
</Story>

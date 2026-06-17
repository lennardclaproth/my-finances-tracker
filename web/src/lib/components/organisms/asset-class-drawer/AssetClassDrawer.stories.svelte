<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import AssetClassDrawer from './AssetClassDrawer.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import { assetClassDetails } from '$lib/data/fixtures/assets';

	const stocks = assetClassDetails['cls-stocks'];

	const { Story } = defineMeta({
		title: 'Organisms/AssetClassDrawer',
		component: AssetClassDrawer,
		tags: ['autodocs']
	});
</script>

<script lang="ts">
	let open = $state(false);
	let openLoading = $state(false);
</script>

<Story name="With details" asChild>
	<div>
		<Button onclick={() => (open = true)}>Open asset class</Button>
		<AssetClassDrawer bind:open details={stocks} />
	</div>
</Story>

<Story name="Loading" asChild>
	<div>
		<Button onclick={() => (openLoading = true)}>Open (loading)</Button>
		<AssetClassDrawer bind:open={openLoading} loading />
	</div>
</Story>

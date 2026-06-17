<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import FooterBar from './FooterBar.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';

	const { Story } = defineMeta({
		title: 'Organisms/FooterBar',
		component: FooterBar,
		tags: ['autodocs']
	});
</script>

<Story name="No selection" asChild>
	<div class="rounded-xl border border-slate-200">
		<FooterBar total={132} limit={25} offset={0} />
	</div>
</Story>

<Story name="With selection + bulk actions" asChild>
	<div class="rounded-xl border border-slate-200">
		<FooterBar total={132} limit={25} offset={0} selectedCount={3}>
			{#snippet actions()}
				<Button size="sm" variant="outline">Tag</Button>
				<Button size="sm" variant="outline" intent="error">Ignore</Button>
			{/snippet}
		</FooterBar>
	</div>
</Story>

<Story name="Last page (next disabled)" asChild>
	<div class="rounded-xl border border-slate-200">
		<FooterBar total={132} limit={50} offset={100} />
	</div>
</Story>

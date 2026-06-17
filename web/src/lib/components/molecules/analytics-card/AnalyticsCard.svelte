<script lang="ts">
	import type { Snippet } from 'svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Text from '$lib/components/atoms/typography/Text.svelte';

	type Props = {
		/** Optional muted label shown above the content (e.g. a chart caption). */
		title?: string;
		/** Extra classes forwarded to the underlying Panel (e.g. grid column spans). */
		class?: string;
		children: Snippet;
	};

	let { title, class: className = '', children }: Props = $props();
</script>

<Panel shape="xl" shadow="sm" bordered padding="sm" class={className}>
	{#if title}
		<Text as="p" size="xs" tone="muted" weight="medium" class="mb-1">{title}</Text>
	{/if}
	{@render children()}
</Panel>

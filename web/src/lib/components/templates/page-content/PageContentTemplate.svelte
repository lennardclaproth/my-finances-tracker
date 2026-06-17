<script lang="ts">
	import type { Snippet } from 'svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { zClasses } from '$lib/styles/z-index';

	type Props = {
		/** Analytics section above the content (charts / KPIs); shrink-0. */
		analytics?: Snippet;
		/** Primary content; placed inside a flex-1 panel that owns its own scrolling. */
		children: Snippet;
		showFab?: boolean;
		fabIcon?: string;
		fabLabel?: string;
		onFabClick?: () => void;
		/** Render the content inside a panel surface (set false to lay out raw content). */
		panel?: boolean;
		class?: string;
	};

	let {
		analytics,
		children,
		showFab = false,
		fabIcon = 'heroicons:plus',
		fabLabel = 'Create',
		onFabClick,
		panel = true,
		class: className = ''
	}: Props = $props();
</script>

<div
	class={['relative flex h-full min-h-0 flex-col gap-3 px-4 pt-3 pb-4', className]
		.filter(Boolean)
		.join(' ')}
>
	{#if analytics}
		<div class="shrink-0">{@render analytics()}</div>
	{/if}

	{#if panel}
		<Panel
			shape="xl"
			shadow="sm"
			padding="none"
			class="flex min-h-0 flex-1 flex-col overflow-hidden"
		>
			{@render children()}
		</Panel>
	{:else}
		<div class="flex min-h-0 flex-1 flex-col overflow-hidden">
			{@render children()}
		</div>
	{/if}

	{#if showFab}
		<button
			type="button"
			aria-label={fabLabel}
			class={[
				'fixed right-6 bottom-6 inline-flex size-14 items-center justify-center rounded-full border border-amber-200 bg-slate-600 text-amber-200 shadow-lg',
				'transition-all duration-150 ease-out hover:bg-slate-500 active:scale-95',
				'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:ring-offset-2 focus-visible:outline-none',
				zClasses.fab
			].join(' ')}
			onclick={onFabClick}
		>
			<Icon icon={fabIcon} size="lg" />
		</button>
	{/if}
</div>

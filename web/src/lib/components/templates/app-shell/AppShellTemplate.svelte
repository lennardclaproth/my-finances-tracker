<script lang="ts">
	import type { Snippet } from 'svelte';
	import ToastHost from '$lib/components/organisms/toast-host/ToastHost.svelte';

	type Props = {
		/** Sticky top region (typically the TopNavbar). */
		top?: Snippet;
		/** Main scrollable area. */
		children: Snippet;
		/** Render the single app-level toast host. */
		withToastHost?: boolean;
		class?: string;
	};

	let { top, children, withToastHost = true, class: className = '' }: Props = $props();
</script>

<div
	class={['flex h-screen min-h-0 flex-col bg-taupe-100 text-slate-800', className]
		.filter(Boolean)
		.join(' ')}
>
	{#if top}
		<div class="shrink-0">{@render top()}</div>
	{/if}

	<main class="min-h-0 flex-1 overflow-hidden">
		{@render children()}
	</main>

	{#if withToastHost}
		<ToastHost />
	{/if}
</div>

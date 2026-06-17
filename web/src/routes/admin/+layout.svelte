<script lang="ts">
	import { goto } from '$app/navigation';
	import { adminMode } from '$lib/stores/admin.svelte';

	let { children } = $props();

	// Client-side admin guard (DESIGN_PLAN §3.5): admin mode is a client toggle, so non-admins are
	// bounced to the app root.
	$effect(() => {
		if (!adminMode.enabled) {
			// eslint-disable-next-line svelte/no-navigation-without-resolve
			goto('/');
		}
	});
</script>

{#if adminMode.enabled}
	{@render children()}
{:else}
	<p class="p-4 text-sm text-slate-500">Admin access required.</p>
{/if}

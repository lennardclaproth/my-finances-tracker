<script lang="ts">
	import { flip } from 'svelte/animate';
	import { fly } from 'svelte/transition';
	import Alert from '$lib/components/molecules/alert/Alert.svelte';
	import { zClasses } from '$lib/styles/z-index';
	import { toast } from '$lib/stores/toast.svelte';

	type Props = {
		class?: string;
	};

	let { class: className = '' }: Props = $props();
</script>

<div
	class={[
		'pointer-events-none fixed top-4 right-4 flex w-full max-w-sm flex-col gap-2',
		zClasses.modal,
		className
	]
		.filter(Boolean)
		.join(' ')}
	role="region"
	aria-label="Notifications"
>
	{#each toast.items as item (item.id)}
		<div
			class="pointer-events-auto"
			in:fly={{ x: 16, y: -8, duration: 200 }}
			out:fly={{ x: 16, duration: 150 }}
			animate:flip={{ duration: 200 }}
		>
			<Alert
				intent={item.intent}
				title={item.title}
				dismissible={item.dismissible}
				onDismiss={() => toast.dismiss(item.id)}
				class="shadow-lg"
			>
				{item.message}
			</Alert>
		</div>
	{/each}
</div>

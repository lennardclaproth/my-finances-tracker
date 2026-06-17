<script lang="ts">
	import type { Snippet } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { zClasses } from '$lib/styles/z-index';

	type Props = {
		/** Two-way bindable open state. */
		open?: boolean;
		title?: string;
		side?: 'right' | 'left';
		/** Tailwind max-width utility for the panel. */
		width?: string;
		dismissible?: boolean;
		closeOnBackdrop?: boolean;
		onClose?: () => void;
		header?: Snippet;
		footer?: Snippet;
		children: Snippet;
		class?: string;
	};

	let {
		open = $bindable(false),
		title,
		side = 'right',
		width = 'max-w-md',
		dismissible = true,
		closeOnBackdrop = true,
		onClose,
		header,
		footer,
		children,
		class: className = ''
	}: Props = $props();

	function close() {
		if (open === false) return;
		open = false;
		onClose?.();
	}

	// Close on Escape while open.
	$effect(() => {
		if (!open) return;
		function onKeydown(event: KeyboardEvent) {
			if (event.key === 'Escape' && dismissible) close();
		}
		document.addEventListener('keydown', onKeydown, true);
		return () => document.removeEventListener('keydown', onKeydown, true);
	});
</script>

{#if open}
	<div
		class={[
			'fixed inset-0 flex',
			side === 'right' ? 'justify-end' : 'justify-start',
			zClasses.modal
		].join(' ')}
	>
		<button
			type="button"
			aria-label="Close"
			class="absolute inset-0 bg-slate-900/50"
			transition:fade={{ duration: 200 }}
			onclick={() => closeOnBackdrop && close()}
		></button>

		<div
			role="dialog"
			aria-modal="true"
			aria-label={title}
			class={['relative flex h-full w-full flex-col bg-white shadow-2xl', width, className]
				.filter(Boolean)
				.join(' ')}
			transition:fly={{ x: side === 'right' ? 360 : -360, duration: 200 }}
		>
			{#if title || header || dismissible}
				<header class="flex items-start justify-between gap-4 border-b border-slate-200 px-5 py-4">
					<div class="min-w-0 flex-1">
						{#if title}
							<h2 class="text-lg">{title}</h2>
						{/if}
						{#if header}
							{@render header()}
						{/if}
					</div>
					{#if dismissible}
						<button
							type="button"
							aria-label="Close"
							class="-mt-1 -mr-1 inline-flex size-8 shrink-0 items-center justify-center rounded-full text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-800 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none"
							onclick={close}
						>
							<Icon icon="heroicons:x-mark" size="md" />
						</button>
					{/if}
				</header>
			{/if}

			<div class="min-h-0 flex-1 overflow-y-auto px-5 py-4">
				{@render children()}
			</div>

			{#if footer}
				<footer class="flex items-center justify-end gap-2 border-t border-slate-200 px-5 py-4">
					{@render footer()}
				</footer>
			{/if}
		</div>
	</div>
{/if}

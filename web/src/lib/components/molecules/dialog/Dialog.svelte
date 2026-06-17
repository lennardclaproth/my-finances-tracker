<script lang="ts">
	import type { Snippet } from 'svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import type { DialogSize } from './dialog.types';
	import { dialogSizeClasses } from './dialog.types';

	type Props = {
		/** Two-way bindable open state. Uses the native top-layer (`showModal`). */
		open?: boolean;
		size?: DialogSize;
		/** Optional simple title rendered in the header (use the `header` snippet for richer headers). */
		title?: string;
		/** Show the built-in close (×) button. */
		dismissible?: boolean;
		/** Close when the backdrop is clicked. */
		closeOnBackdrop?: boolean;
		ariaLabel?: string;
		onClose?: () => void;
		class?: string;
		header?: Snippet;
		footer?: Snippet;
		children: Snippet;
	};

	let {
		open = $bindable(false),
		size = 'md',
		title,
		dismissible = true,
		closeOnBackdrop = true,
		ariaLabel,
		onClose,
		class: className = '',
		header,
		footer,
		children
	}: Props = $props();

	let dialogEl = $state<HTMLDialogElement | null>(null);

	function setOpen(next: boolean) {
		if (open === next) return;
		open = next;
		if (!next) onClose?.();
	}

	// Drive the native dialog from the bindable `open` state.
	$effect(() => {
		const el = dialogEl;
		if (!el) return;
		if (open && !el.open) el.showModal();
		else if (!open && el.open) el.close();
	});

	function handleClose() {
		// Fires for Esc / native close / programmatic close — keep `open` in sync.
		setOpen(false);
	}

	function handleClick(event: MouseEvent) {
		// With padding moved to the inner wrapper, a click whose target is the <dialog> itself is a
		// backdrop click.
		if (closeOnBackdrop && event.target === dialogEl) setOpen(false);
	}

	const dialogClasses = $derived(
		[
			'w-full rounded-2xl bg-white p-0 text-slate-800 shadow-2xl',
			'backdrop:bg-slate-900/50',
			dialogSizeClasses[size],
			className
		]
			.filter(Boolean)
			.join(' ')
	);

	const showHeader = $derived(Boolean(title || header || dismissible));
</script>

<dialog
	bind:this={dialogEl}
	class={dialogClasses}
	aria-label={ariaLabel ?? title}
	onclose={handleClose}
	onclick={handleClick}
>
	<div class="flex max-h-[85vh] flex-col">
		{#if showHeader}
			<header class="flex items-start justify-between gap-4 px-5 pt-5">
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
						onclick={() => setOpen(false)}
					>
						<Icon icon="heroicons:x-mark" size="md" />
					</button>
				{/if}
			</header>
		{/if}

		<div class="min-h-0 flex-1 overflow-y-auto px-5 py-4 text-sm">
			{@render children()}
		</div>

		{#if footer}
			<footer class="flex items-center justify-end gap-2 border-t border-slate-200 px-5 py-4">
				{@render footer()}
			</footer>
		{/if}
	</div>
</dialog>

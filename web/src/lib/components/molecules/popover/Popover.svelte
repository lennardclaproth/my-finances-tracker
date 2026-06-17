<script lang="ts">
	import type { Snippet } from 'svelte';
	import { zClasses } from '$lib/styles/z-index';
	import type { PopoverApi, PopoverPlacement } from './popover.types';

	type Props = {
		/** Two-way bindable open state (controlled or uncontrolled). */
		open?: boolean;
		placement?: PopoverPlacement;
		/** Gap in px between the trigger and the panel. */
		offset?: number;
		/** Append the panel to <body> to escape clipping/transform ancestors. */
		portal?: boolean;
		/** Make the panel the same width as the trigger. */
		matchWidth?: boolean;
		closeOnOutsideClick?: boolean;
		closeOnEscape?: boolean;
		onOpenChange?: (open: boolean) => void;
		/** Extra classes for the floating panel. */
		class?: string;
		/** Trigger markup; receives the control API (on a native control wire `aria-expanded={api.open}`). */
		trigger: Snippet<[PopoverApi]>;
		/** Panel content. */
		children: Snippet;
	};

	let {
		open = $bindable(false),
		placement = 'bottom-start',
		offset = 8,
		portal = false,
		matchWidth = false,
		closeOnOutsideClick = true,
		closeOnEscape = true,
		onOpenChange,
		class: className = '',
		trigger,
		children
	}: Props = $props();

	let anchorEl = $state<HTMLElement | null>(null);
	let panelEl = $state<HTMLElement | null>(null);
	let coords = $state<{ top: number; left: number; width?: number }>({ top: 0, left: 0 });
	let ready = $state(false);

	function setOpen(next: boolean) {
		if (open === next) return;
		open = next;
		onOpenChange?.(next);
	}

	const api: PopoverApi = {
		get open() {
			return open;
		},
		toggle: () => setOpen(!open),
		show: () => setOpen(true),
		close: () => setOpen(false)
	};

	function computePosition() {
		if (!anchorEl || !panelEl) return;
		const a = anchorEl.getBoundingClientRect();
		const p = panelEl.getBoundingClientRect();
		const vw = window.innerWidth;
		const vh = window.innerHeight;
		const gap = offset;

		const [side, align = 'center'] = placement.split('-') as [
			'top' | 'bottom' | 'left' | 'right',
			'start' | 'end' | 'center' | undefined
		];

		let top: number;
		let left: number;

		if (side === 'bottom' || side === 'top') {
			top = side === 'bottom' ? a.bottom + gap : a.top - p.height - gap;
			if (align === 'start') left = a.left;
			else if (align === 'end') left = a.right - p.width;
			else left = a.left + a.width / 2 - p.width / 2;
			// Flip vertically when the preferred side overflows and the opposite side fits.
			if (side === 'bottom' && top + p.height > vh - 8 && a.top - p.height - gap > 8) {
				top = a.top - p.height - gap;
			} else if (side === 'top' && top < 8 && a.bottom + gap + p.height < vh - 8) {
				top = a.bottom + gap;
			}
		} else {
			left = side === 'right' ? a.right + gap : a.left - p.width - gap;
			if (align === 'start') top = a.top;
			else if (align === 'end') top = a.bottom - p.height;
			else top = a.top + a.height / 2 - p.height / 2;
			if (side === 'right' && left + p.width > vw - 8 && a.left - p.width - gap > 8) {
				left = a.left - p.width - gap;
			} else if (side === 'left' && left < 8 && a.right + gap + p.width < vw - 8) {
				left = a.right + gap;
			}
		}

		left = Math.min(Math.max(8, left), Math.max(8, vw - p.width - 8));
		top = Math.min(Math.max(8, top), Math.max(8, vh - p.height - 8));

		coords = { top, left, width: matchWidth ? a.width : undefined };
		ready = true;
	}

	// Position the panel while open and keep it pinned on scroll/resize.
	$effect(() => {
		if (!open) {
			ready = false;
			return;
		}
		if (!anchorEl || !panelEl) return;
		computePosition();
		const reposition = () => computePosition();
		window.addEventListener('scroll', reposition, true);
		window.addEventListener('resize', reposition);
		return () => {
			window.removeEventListener('scroll', reposition, true);
			window.removeEventListener('resize', reposition);
		};
	});

	// Dismiss on outside pointerdown / Escape.
	$effect(() => {
		if (!open) return;
		function onPointerDown(event: PointerEvent) {
			const target = event.target as Node;
			if (anchorEl?.contains(target) || panelEl?.contains(target)) return;
			if (closeOnOutsideClick) setOpen(false);
		}
		function onKeydown(event: KeyboardEvent) {
			if (event.key === 'Escape' && closeOnEscape) setOpen(false);
		}
		document.addEventListener('pointerdown', onPointerDown, true);
		document.addEventListener('keydown', onKeydown, true);
		return () => {
			document.removeEventListener('pointerdown', onPointerDown, true);
			document.removeEventListener('keydown', onKeydown, true);
		};
	});

	// Optionally relocate the panel to <body>.
	function portalAction(node: HTMLElement, enabled: boolean) {
		if (enabled) document.body.appendChild(node);
		return {
			destroy() {
				if (node.parentNode) node.parentNode.removeChild(node);
			}
		};
	}

	const panelClasses = $derived(
		[
			'rounded-xl border border-slate-300 bg-white shadow-md',
			zClasses.popover,
			ready ? '' : 'invisible',
			className
		]
			.filter(Boolean)
			.join(' ')
	);
</script>

<span bind:this={anchorEl} class="inline-flex">
	{@render trigger(api)}
</span>

{#if open}
	<div
		bind:this={panelEl}
		use:portalAction={portal}
		class={panelClasses}
		style={`position: fixed; top: ${coords.top}px; left: ${coords.left}px;${coords.width ? ` width: ${coords.width}px;` : ''}`}
	>
		{@render children()}
	</div>
{/if}

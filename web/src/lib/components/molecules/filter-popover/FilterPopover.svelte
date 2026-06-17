<script lang="ts">
	import type { Snippet } from 'svelte';
	import Popover from '$lib/components/molecules/popover/Popover.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';

	type Props = {
		/** Bindable open state of the popover. */
		open?: boolean;
		/** Drives the active (filtered) trigger styling — set when the filter has a value. */
		active?: boolean;
		/** Accessible label for the trigger button. */
		label: string;
		/** Iconify id for the trigger (defaults to a funnel). */
		icon?: string;
		/** Optional count shown as a badge on the trigger (e.g. number of selected options). */
		count?: number;
		/** Render the Clear / Apply footer. */
		showFooter?: boolean;
		applyLabel?: string;
		clearLabel?: string;
		/** Extra classes for the panel body wrapper. */
		panelClass?: string;
		onApply?: () => void;
		onClear?: () => void;
		children: Snippet;
	};

	let {
		open = $bindable(false),
		active = false,
		label,
		icon = 'heroicons:funnel',
		count,
		showFooter = true,
		applyLabel = 'Apply',
		clearLabel = 'Clear',
		panelClass = '',
		onApply,
		onClear,
		children
	}: Props = $props();

	function triggerClasses(highlighted: boolean): string {
		return [
			'relative inline-flex size-9 items-center justify-center rounded-full transition-colors',
			'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
			highlighted
				? 'bg-amber-200 text-slate-800 hover:bg-amber-300'
				: 'bg-transparent text-slate-500 hover:bg-slate-100'
		].join(' ');
	}

	function apply() {
		onApply?.();
		open = false;
	}

	function clear() {
		onClear?.();
		open = false;
	}
</script>

<Popover bind:open placement="bottom-start">
	{#snippet trigger(api)}
		<button
			type="button"
			aria-label={label}
			aria-expanded={api.open}
			class={triggerClasses(active || api.open)}
			onclick={api.toggle}
		>
			<Icon {icon} size="md" />
			{#if count}
				<span
					class="absolute -top-1 -right-1 inline-flex h-4 min-w-4 items-center justify-center rounded-full bg-slate-600 px-1 text-[10px] leading-none font-semibold text-amber-200"
				>
					{count}
				</span>
			{/if}
		</button>
	{/snippet}

	<div class={['p-3', panelClass].filter(Boolean).join(' ')}>
		{@render children()}

		{#if showFooter}
			<div class="mt-3 flex justify-end gap-2 border-t border-slate-200 pt-3">
				<Button size="sm" variant="ghost" intent="secondary" onclick={clear}>{clearLabel}</Button>
				<Button size="sm" onclick={apply}>{applyLabel}</Button>
			</div>
		{/if}
	</div>
</Popover>

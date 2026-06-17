<script lang="ts">
	import type { TabItem, TabsSize } from './tabs.types';
	import {
		tabBaseClasses,
		tabSizeClasses,
		tabStateClasses,
		tabsTrackClasses
	} from './tabs.variants';

	type Props = {
		tabs: TabItem[];
		/** Two-way bindable selected value. Defaults to the first tab. */
		value?: string;
		size?: TabsSize;
		ariaLabel?: string;
		onChange?: (value: string) => void;
		class?: string;
	};

	let {
		tabs,
		value = $bindable(tabs[0]?.value ?? ''),
		size = 'md',
		ariaLabel,
		onChange,
		class: className = ''
	}: Props = $props();

	function select(next: string) {
		if (value === next) return;
		value = next;
		onChange?.(next);
	}

	// Roving arrow-key navigation across enabled tabs.
	function onKeydown(event: KeyboardEvent) {
		if (event.key !== 'ArrowRight' && event.key !== 'ArrowLeft') return;
		const enabled = tabs.filter((t) => !t.disabled);
		const currentIndex = enabled.findIndex((t) => t.value === value);
		if (currentIndex === -1) return;
		event.preventDefault();
		const delta = event.key === 'ArrowRight' ? 1 : -1;
		const next = enabled[(currentIndex + delta + enabled.length) % enabled.length];
		select(next.value);
	}

	const trackClasses = $derived([tabsTrackClasses, className].filter(Boolean).join(' '));
</script>

<div role="tablist" aria-label={ariaLabel} class={trackClasses}>
	{#each tabs as tab (tab.value)}
		{@const selected = tab.value === value}
		<button
			type="button"
			role="tab"
			aria-selected={selected}
			tabindex={selected ? 0 : -1}
			disabled={tab.disabled}
			class={[
				tabBaseClasses,
				tabSizeClasses[size],
				selected ? tabStateClasses.selected : tabStateClasses.unselected
			].join(' ')}
			onclick={() => select(tab.value)}
			onkeydown={onKeydown}
		>
			{tab.label}
		</button>
	{/each}
</div>

<script lang="ts">
	import { fly } from 'svelte/transition';
	import Popover from '$lib/components/molecules/popover/Popover.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import type { PopoverPlacement } from '$lib/components/molecules/popover/popover.types';
	import type { MenuItem } from './menu.types';

	type Props = {
		items: MenuItem[];
		label?: string;
		/** Iconify id for the trigger's trailing affordance (rotates when open). */
		icon?: string;
		placement?: PopoverPlacement;
		class?: string;
	};

	let {
		items,
		label = 'Actions',
		icon = 'heroicons:chevron-down',
		placement = 'bottom-end',
		class: className = ''
	}: Props = $props();

	let open = $state(false);

	function choose(item: MenuItem) {
		if (item.disabled) return;
		item.onSelect?.();
		open = false;
	}
</script>

<Popover bind:open {placement}>
	{#snippet trigger(api)}
		<button
			type="button"
			aria-haspopup="menu"
			aria-expanded={api.open}
			class={[
				'inline-flex h-9 items-center gap-1.5 rounded-lg border border-slate-300 bg-white px-3 text-sm text-slate-700 transition-colors',
				'hover:bg-slate-50 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
				className
			].join(' ')}
			onclick={api.toggle}
		>
			<span>{label}</span>
			<Icon
				{icon}
				size="sm"
				class={[
					'text-slate-400 transition-transform duration-200',
					api.open ? 'rotate-180' : ''
				].join(' ')}
			/>
		</button>
	{/snippet}

	<ul role="menu" class="min-w-44 py-1" transition:fly={{ y: -4, duration: 150 }}>
		{#each items as item, index (item.label)}
			{#if item.divider}
				<li role="separator" class="my-1 border-t border-slate-200"></li>
			{/if}
			<li role="none" in:fly={{ y: -4, duration: 150, delay: index * 20 }}>
				<button
					type="button"
					role="menuitem"
					disabled={item.disabled}
					class={[
						'flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm transition-colors',
						'disabled:pointer-events-none disabled:opacity-50',
						item.intent === 'danger'
							? 'text-red-700 hover:bg-red-50'
							: 'text-slate-700 hover:bg-slate-50'
					].join(' ')}
					onclick={() => choose(item)}
				>
					{#if item.icon}
						<Icon icon={item.icon} size="sm" class="text-slate-400" />
					{/if}
					{item.label}
				</button>
			</li>
		{/each}
	</ul>
</Popover>

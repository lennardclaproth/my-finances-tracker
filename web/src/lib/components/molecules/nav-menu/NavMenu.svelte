<script lang="ts">
	import { fly } from 'svelte/transition';
	import { goto } from '$app/navigation';
	import Popover from '$lib/components/molecules/popover/Popover.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import type { PopoverPlacement } from '$lib/components/molecules/popover/popover.types';
	import type { NavItem } from './nav-menu.types';

	type Props = {
		/** Routes to list in the menu. */
		items: NavItem[];
		/** Current route path; the matching item is highlighted as active. */
		currentPath?: string;
		/** Iconify id for the trigger button (defaults to a hamburger). */
		icon?: string;
		placement?: PopoverPlacement;
		ariaLabel?: string;
		class?: string;
	};

	let {
		items,
		currentPath = '',
		icon = 'heroicons:bars-3',
		placement = 'bottom-start',
		ariaLabel = 'Open navigation menu',
		class: className = ''
	}: Props = $props();

	let open = $state(false);

	function isActive(href: string): boolean {
		return currentPath === href || currentPath.startsWith(`${href}/`);
	}

	function choose(item: NavItem) {
		open = false;
		if (!isActive(item.href)) void goto(item.href);
	}
</script>

<Popover bind:open {placement}>
	{#snippet trigger(api)}
		<button
			type="button"
			aria-haspopup="menu"
			aria-expanded={api.open}
			aria-label={ariaLabel}
			class={[
				'inline-flex size-9 items-center justify-center rounded-lg text-slate-700 transition-colors',
				'hover:bg-slate-100 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
				className
			]
				.filter(Boolean)
				.join(' ')}
			onclick={api.toggle}
		>
			<Icon {icon} size="md" />
		</button>
	{/snippet}

	<ul role="menu" class="min-w-48 py-1" transition:fly={{ y: -4, duration: 150 }}>
		{#each items as item, index (item.href)}
			{#if item.divider}
				<li role="separator" class="my-1 border-t border-slate-200"></li>
			{/if}
			<li role="none" in:fly={{ y: -4, duration: 150, delay: index * 20 }}>
				<button
					type="button"
					role="menuitem"
					aria-current={isActive(item.href) ? 'page' : undefined}
					class={[
						'flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm transition-colors',
						isActive(item.href)
							? 'bg-amber-100 font-medium text-slate-900'
							: 'text-slate-700 hover:bg-slate-50'
					].join(' ')}
					onclick={() => choose(item)}
				>
					{#if item.icon}
						<Icon
							icon={item.icon}
							size="sm"
							class={isActive(item.href) ? 'text-amber-700' : 'text-slate-500'}
						/>
					{/if}
					{item.label}
				</button>
			</li>
		{/each}
	</ul>
</Popover>

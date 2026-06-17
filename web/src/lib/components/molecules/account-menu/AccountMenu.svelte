<script lang="ts">
	import { fly } from 'svelte/transition';
	import Popover from '$lib/components/molecules/popover/Popover.svelte';
	import Avatar from '$lib/components/atoms/avatar/Avatar.svelte';
	import Switch from '$lib/components/atoms/switch/Switch.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	type Props = {
		name: string;
		email?: string;
		/** Avatar image; falls back to initials derived from `name`. */
		src?: string;
		/** Bindable admin-mode flag, surfaced as a toggle in the menu. */
		adminMode?: boolean;
		showAdminToggle?: boolean;
		items?: MenuItem[];
		onAdminToggle?: (value: boolean) => void;
		class?: string;
	};

	let {
		name,
		email,
		src,
		adminMode = $bindable(false),
		showAdminToggle = true,
		items = [],
		onAdminToggle,
		class: className = ''
	}: Props = $props();

	let open = $state(false);

	function choose(item: MenuItem) {
		if (item.disabled) return;
		item.onSelect?.();
		open = false;
	}

	function toggleAdmin() {
		adminMode = !adminMode;
		onAdminToggle?.(adminMode);
	}
</script>

<Popover bind:open placement="bottom-end">
	{#snippet trigger(api)}
		<button
			type="button"
			aria-haspopup="menu"
			aria-expanded={api.open}
			aria-label="Account menu"
			class={[
				'inline-flex items-center gap-2 rounded-full p-0.5 transition-colors',
				'hover:bg-slate-100 focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
				className
			].join(' ')}
			onclick={api.toggle}
		>
			<Avatar {name} {src} size="sm" />
		</button>
	{/snippet}

	<div class="w-60 py-1" transition:fly={{ y: -4, duration: 150 }}>
		<div class="flex items-center gap-3 px-3 py-2">
			<Avatar {name} {src} size="md" />
			<div class="min-w-0">
				<p class="truncate text-sm font-medium text-slate-900">{name}</p>
				{#if email}
					<p class="truncate text-xs text-slate-500">{email}</p>
				{/if}
			</div>
		</div>

		{#if showAdminToggle}
			<div class="my-1 border-t border-slate-200"></div>
			<div class="flex items-center justify-between gap-3 px-3 py-1.5 text-sm text-slate-700">
				<span class="flex items-center gap-2">
					<Icon icon="heroicons:shield-check" size="sm" class="text-slate-400" />
					Admin mode
				</span>
				<Switch checked={adminMode} aria-label="Admin mode" onchange={toggleAdmin} />
			</div>
		{/if}

		{#if items.length > 0}
			<div class="my-1 border-t border-slate-200"></div>
			<ul role="menu">
				{#each items as item (item.label)}
					<li role="none">
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
		{/if}
	</div>
</Popover>

<script lang="ts">
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import Link from '$lib/components/atoms/link/Link.svelte';
	import type { BreadcrumbItem } from './breadcrumb.types';

	type Props = {
		items: BreadcrumbItem[];
		/** Iconify id for the separator between segments. */
		separatorIcon?: string;
		ariaLabel?: string;
		class?: string;
	};

	let {
		items,
		separatorIcon = 'heroicons:chevron-right',
		ariaLabel = 'Breadcrumb',
		class: className = ''
	}: Props = $props();

	const navClasses = $derived(['min-w-0', className].filter(Boolean).join(' '));
</script>

<nav aria-label={ariaLabel} class={navClasses}>
	<ol class="flex flex-wrap items-center gap-1.5 text-sm">
		{#each items as item, index (index)}
			{@const last = index === items.length - 1}
			<li class="flex items-center gap-1.5">
				{#if item.href && !last}
					<Link href={item.href} tone="muted" size="sm" underline="hover">{item.label}</Link>
				{:else}
					<span
						class={last ? 'font-medium text-slate-800' : 'text-slate-500'}
						aria-current={last ? 'page' : undefined}
					>
						{item.label}
					</span>
				{/if}

				{#if !last}
					<Icon icon={separatorIcon} size="sm" class="text-slate-300" />
				{/if}
			</li>
		{/each}
	</ol>
</nav>

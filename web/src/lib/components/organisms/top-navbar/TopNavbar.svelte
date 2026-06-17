<script lang="ts">
	import { page } from '$app/state';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import DateRangePicker from '$lib/components/molecules/date-range-picker/DateRangePicker.svelte';
	import ActionMenu from '$lib/components/molecules/action-menu/ActionMenu.svelte';
	import AccountMenu from '$lib/components/molecules/account-menu/AccountMenu.svelte';
	import NavMenu from '$lib/components/molecules/nav-menu/NavMenu.svelte';
	import Panel from '$lib/components/atoms/panel/Panel.svelte';
	import Heading from '$lib/components/atoms/typography/Heading.svelte';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';
	import type { NavItem } from '$lib/components/molecules/nav-menu/nav-menu.types';

	type DateRange = { from: string | null; to: string | null };

	type Props = {
		title?: string;
		showSearch?: boolean;
		searchValue?: string;
		searchPlaceholder?: string;
		onSearch?: (query: string) => void;
		showDateRange?: boolean;
		dateFrom?: string | null;
		dateTo?: string | null;
		onDateChange?: (range: DateRange) => void;
		actions?: MenuItem[];
		accountName: string;
		accountEmail?: string;
		adminMode?: boolean;
		showAdminToggle?: boolean;
		accountItems?: MenuItem[];
		onAdminToggle?: (value: boolean) => void;
		class?: string;
	};

	let {
		title,
		showSearch = false,
		searchValue = $bindable(''),
		searchPlaceholder = 'Search…',
		onSearch,
		showDateRange = false,
		dateFrom = $bindable(null),
		dateTo = $bindable(null),
		onDateChange,
		actions = [],
		accountName,
		accountEmail,
		adminMode = $bindable(false),
		showAdminToggle = true,
		accountItems = [],
		onAdminToggle,
		class: className = ''
	}: Props = $props();

	// The single source of in-app navigation, surfaced through the left menu button.
	const navItems: NavItem[] = [
		{ label: 'Cashflow', href: '/cashflow', icon: 'heroicons:banknotes' },
		{ label: 'Assets', href: '/assets', icon: 'heroicons:building-library' },
		{ label: 'Portfolio', href: '/portfolio', icon: 'heroicons:chart-pie' },
		{ label: 'Listings', href: '/admin/listings', icon: 'heroicons:cog-6-tooth', divider: true }
	];
</script>

<header
	class={['flex flex-col gap-3 px-4 py-3 md:flex-row md:items-center', className]
		.filter(Boolean)
		.join(' ')}
>
	<Panel
		variant="floating"
		shape="xl"
		shadow="sm"
		padding="sm"
		class="flex items-center gap-2 md:shrink-0"
	>
		<NavMenu items={navItems} currentPath={page.url.pathname} />
		{#if title}
			<Heading level="h1" size="xl" class="leading-none">{title}</Heading>
		{/if}
	</Panel>

	<div class="flex justify-center md:flex-1">
		{#if showSearch}
			<div class="w-full md:max-w-lg">
				<SearchInput
					bind:value={searchValue}
					{onSearch}
					placeholder={searchPlaceholder}
					size="lg"
					shape="pill"
				/>
			</div>
		{/if}
	</div>

	<div class="flex flex-wrap items-center gap-2 md:shrink-0">
		{#if showDateRange}
			<DateRangePicker bind:from={dateFrom} bind:to={dateTo} size="lg" onChange={onDateChange} />
		{/if}
		{#if actions.length > 0}
			<ActionMenu items={actions} />
		{/if}
		<AccountMenu
			name={accountName}
			email={accountEmail}
			bind:adminMode
			{showAdminToggle}
			items={accountItems}
			{onAdminToggle}
		/>
	</div>
</header>

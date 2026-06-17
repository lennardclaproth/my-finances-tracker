<script lang="ts">
	import Breadcrumb from '$lib/components/molecules/breadcrumb/Breadcrumb.svelte';
	import SearchInput from '$lib/components/molecules/search-input/SearchInput.svelte';
	import DateRangePicker from '$lib/components/molecules/date-range-picker/DateRangePicker.svelte';
	import ActionMenu from '$lib/components/molecules/action-menu/ActionMenu.svelte';
	import AccountMenu from '$lib/components/molecules/account-menu/AccountMenu.svelte';
	import type { BreadcrumbItem } from '$lib/components/molecules/breadcrumb/breadcrumb.types';
	import type { MenuItem } from '$lib/components/molecules/action-menu/menu.types';

	type DateRange = { from: string | null; to: string | null };

	type Props = {
		title?: string;
		breadcrumb?: BreadcrumbItem[];
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
		breadcrumb,
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
</script>

<header
	class={[
		'flex flex-col gap-3 border-b border-slate-200 bg-white/95 px-4 py-3 backdrop-blur md:flex-row md:items-center md:justify-between',
		className
	]
		.filter(Boolean)
		.join(' ')}
>
	<div class="flex min-w-0 flex-col gap-1">
		{#if breadcrumb}
			<Breadcrumb items={breadcrumb} />
		{/if}
		{#if title}
			<h1 class="font-heading text-xl text-slate-900 md:text-2xl">{title}</h1>
		{/if}
	</div>

	<div class="flex flex-wrap items-center gap-2">
		{#if showSearch}
			<div class="w-full sm:w-56">
				<SearchInput
					bind:value={searchValue}
					{onSearch}
					placeholder={searchPlaceholder}
					size="sm"
				/>
			</div>
		{/if}
		{#if showDateRange}
			<DateRangePicker bind:from={dateFrom} bind:to={dateTo} size="sm" onChange={onDateChange} />
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

<script lang="ts">
	import Input from '$lib/components/atoms/input/Input.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import type { InputShape, InputSize } from '$lib/components/atoms/input/input.types';
	import {
		inputIconContainerSizeClasses,
		inputIconPaddingClasses,
		inputIconSizeClasses
	} from '$lib/components/molecules/icon-input/icon-input.variants';

	type Props = {
		/** Two-way bindable query value. */
		value?: string;
		placeholder?: string;
		size?: InputSize;
		shape?: InputShape;
		disabled?: boolean;
		/** Debounce window (ms) before `onSearch` fires. */
		debounceMs?: number;
		ariaLabel?: string;
		/** Debounced query callback (fires after `debounceMs` of inactivity; immediate on clear). */
		onSearch?: (value: string) => void;
		/** Raw, non-debounced input callback. */
		oninput?: (event: Event) => void;
		class?: string;
	};

	let {
		value = $bindable(''),
		placeholder = 'Search…',
		size = 'md',
		shape = 'rounded',
		disabled = false,
		debounceMs = 300,
		ariaLabel = 'Search',
		onSearch,
		oninput,
		class: className = ''
	}: Props = $props();

	let mounted = false;

	// Debounce `onSearch` on value changes; skip the initial mount so it doesn't fire for the seed value.
	$effect(() => {
		const current = value;
		if (!mounted) {
			mounted = true;
			return;
		}
		if (!onSearch) return;
		const timer = setTimeout(() => onSearch(current), debounceMs);
		return () => clearTimeout(timer);
	});

	function clear() {
		value = '';
		onSearch?.('');
	}

	const clearable = $derived(value.length > 0 && !disabled);

	const inputClasses = $derived(
		[
			inputIconPaddingClasses[size].left,
			clearable ? inputIconPaddingClasses[size].right : '',
			className
		]
			.filter(Boolean)
			.join(' ')
	);
</script>

<div class="group relative w-full">
	<span
		class={[
			'pointer-events-none absolute inset-y-0 left-0 flex items-center justify-center text-slate-500',
			inputIconContainerSizeClasses[size]
		].join(' ')}
		aria-hidden="true"
	>
		<Icon icon="heroicons:magnifying-glass" size={inputIconSizeClasses[size]} />
	</span>

	<Input
		type="search"
		bind:value
		{placeholder}
		{size}
		{shape}
		{disabled}
		{ariaLabel}
		class={inputClasses}
		{oninput}
	/>

	{#if clearable}
		<button
			type="button"
			aria-label="Clear search"
			class={[
				'absolute inset-y-0 right-0 flex items-center justify-center text-slate-500',
				'transition-colors hover:text-slate-700',
				'focus-visible:ring-2 focus-visible:ring-slate-300 focus-visible:outline-none',
				inputIconContainerSizeClasses[size]
			].join(' ')}
			onclick={clear}
		>
			<Icon icon="heroicons:x-mark" size={inputIconSizeClasses[size]} />
		</button>
	{/if}
</div>

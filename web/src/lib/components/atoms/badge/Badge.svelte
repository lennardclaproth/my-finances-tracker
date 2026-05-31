<script lang="ts">
	import type { Snippet } from 'svelte';

	import type {
		BadgeIntent,
		BadgeShape,
		BadgeSize,
		BadgeVariant
	} from './badge.types';

	import {
		baseBadgeClasses,
		badgeDotIntentClasses,
		badgeDotSizeClasses,
		badgeIntentVariantClasses,
		badgeShapeClasses,
		badgeSizeClasses
	} from './badge.variants';

	type Props = {
		children: Snippet;
		intent?: BadgeIntent;
		variant?: BadgeVariant;
		size?: BadgeSize;
		shape?: BadgeShape;
		dot?: boolean;
		class?: string;
	};

	let {
		children,
		intent = 'neutral',
		variant = 'soft',
		size = 'md',
		shape = 'pill',
		dot = false,
		class: className = ''
	}: Props = $props();
</script>

<span
	class={[
		baseBadgeClasses,
		badgeSizeClasses[size],
		badgeShapeClasses[shape],
		badgeIntentVariantClasses[intent][variant],
		className
	]}
>
	{#if dot}
		<span
			class={[
				'shrink-0 rounded-full',
				badgeDotSizeClasses[size],
				badgeDotIntentClasses[intent]
			]}
			aria-hidden="true"
		></span>
	{/if}

	{@render children()}
</span>
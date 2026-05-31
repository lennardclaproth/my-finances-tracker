<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import type { ComponentProps } from 'svelte';

	import Badge from './Badge.svelte';

	import {
		badgeIntents,
		badgeShapes,
		badgeSizes,
		badgeVariants,
		type BadgeIntent,
		type BadgeVariant
	} from './badge.types';

	type BadgeArgs = Omit<ComponentProps<typeof Badge>, 'children'> & {
		children: string;
	};

	const intentLabels: Record<BadgeIntent, string> = {
		neutral: 'Neutral',
		primary: 'Primary',
        secondary: 'Secondary',
		success: 'Success',
		warning: 'Warning',
		error: 'Error',
		info: 'Info'
	};

	const variantLabels: Record<BadgeVariant, string> = {
		soft: 'Soft',
		solid: 'Solid',
		outline: 'Outline'
	};

	const { Story } = defineMeta({
		title: 'Atoms/Badge',
		component: Badge,
		render: template,
		tags: ['autodocs'],
		argTypes: {
			children: {
				control: 'text',
				description: 'Visible badge text.'
			},
			intent: {
				control: 'select',
				options: badgeIntents,
				description: 'Defines the semantic meaning and colour treatment.'
			},
			variant: {
				control: 'select',
				options: badgeVariants,
				description: 'Defines the visual emphasis of the badge.'
			},
			size: {
				control: 'select',
				options: badgeSizes,
				description: 'Defines the internal spacing and text size.'
			},
			shape: {
				control: 'select',
				options: badgeShapes,
				description: 'Defines the corner radius of the badge.'
			},
			dot: {
				control: 'boolean',
				description: 'Adds a small semantic status indicator.'
			}
		},
		args: {
			children: 'Badge',
			intent: 'neutral',
			variant: 'soft',
			size: 'md',
			shape: 'pill',
			dot: false
		}
	});
</script>

{#snippet template({ children, ...args }: BadgeArgs)}
	<Badge {...args}>{children}</Badge>
{/snippet}

<Story name="Playground" />

<Story name="Variants" asChild>
	<div class="flex flex-col gap-8">
		{#each badgeVariants as variant}
			<section class="flex flex-col gap-3">
				<h3 class="text-sm font-semibold capitalize">{variant}</h3>

				<div class="flex flex-wrap items-center gap-3">
					{#each badgeIntents as intent}
						<Badge {variant} {intent}>
							{intentLabels[intent]}
						</Badge>
					{/each}
				</div>
			</section>
		{/each}
	</div>
</Story>

<Story name="Sizes" asChild>
	<div class="flex flex-wrap items-center gap-3">
		<Badge size="sm" intent="primary">Small</Badge>
		<Badge size="md" intent="primary">Medium</Badge>
		<Badge size="lg" intent="primary">Large</Badge>
	</div>
</Story>

<Story name="Shapes" asChild>
	<div class="flex flex-wrap items-center gap-3">
		<Badge shape="rounded" intent="primary">Rounded</Badge>
		<Badge shape="pill" intent="primary">Pill</Badge>
	</div>
</Story>

<Story name="Status Dot" asChild>
	<div class="flex flex-wrap items-center gap-3">
		<Badge dot intent="neutral">Archived</Badge>
		<Badge dot intent="primary">Selected</Badge>
		<Badge dot intent="success">Active</Badge>
		<Badge dot intent="warning">Pending</Badge>
		<Badge dot intent="error">Failed</Badge>
		<Badge dot intent="info">Scheduled</Badge>
	</div>
</Story>
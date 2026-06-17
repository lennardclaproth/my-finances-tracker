<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import type { ComponentProps } from 'svelte';

	import Popover from './Popover.svelte';
	import Button from '$lib/components/atoms/button/Button.svelte';
	import Icon from '$lib/components/atoms/icon/Icon.svelte';
	import { popoverPlacements } from './popover.types';

	type PopoverProps = ComponentProps<typeof Popover>;

	const { Story } = defineMeta({
		title: 'Molecules/Popover',
		component: Popover,
		tags: ['autodocs'],
		argTypes: {
			placement: { control: 'select', options: popoverPlacements },
			portal: { control: 'boolean' },
			matchWidth: { control: 'boolean' },
			offset: { control: 'number' }
		}
	});
</script>

{#snippet playground(args: PopoverProps)}
	<div class="flex min-h-64 items-center justify-center">
		<Popover
			placement={args.placement}
			portal={args.portal}
			matchWidth={args.matchWidth}
			offset={args.offset}
		>
			{#snippet trigger(api)}
				<Button onclick={api.toggle}>
					<Icon icon="heroicons:funnel" />
					Open popover
				</Button>
			{/snippet}
			<div class="w-64 p-4 text-sm text-slate-700">
				<p class="mb-1 font-semibold text-slate-900">Popover panel</p>
				<p>Click outside or press <kbd>Esc</kbd> to dismiss.</p>
			</div>
		</Popover>
	</div>
{/snippet}

<Story
	name="Playground"
	args={{ placement: 'bottom-start', portal: false, matchWidth: false, offset: 8 }}
	template={playground}
/>

<Story name="Placements" asChild>
	<div class="grid grid-cols-2 place-items-center gap-10 p-16 md:grid-cols-3">
		{#each ['top', 'bottom', 'left', 'right', 'bottom-start', 'bottom-end'] as const as placement (placement)}
			<Popover {placement}>
				{#snippet trigger(api)}
					<Button variant="outline" onclick={api.toggle}>{placement}</Button>
				{/snippet}
				<div class="w-44 p-3 text-sm text-slate-700">Placed <strong>{placement}</strong></div>
			</Popover>
		{/each}
	</div>
</Story>

<Story name="Match trigger width" asChild>
	<div class="flex min-h-48 items-start justify-center p-8">
		<Popover matchWidth placement="bottom-start">
			{#snippet trigger(api)}
				<Button class="w-72" onclick={api.toggle}>
					Full-width trigger
					<Icon icon="heroicons:chevron-down" />
				</Button>
			{/snippet}
			<ul class="py-1 text-sm text-slate-700">
				<li class="px-3 py-1.5 hover:bg-slate-50">Option one</li>
				<li class="px-3 py-1.5 hover:bg-slate-50">Option two</li>
				<li class="px-3 py-1.5 hover:bg-slate-50">Option three</li>
			</ul>
		</Popover>
	</div>
</Story>

<Story name="Portaled" asChild>
	<div class="h-40 overflow-hidden rounded-xl border border-slate-300 p-4">
		<p class="mb-3 text-sm text-slate-500">
			This container has <code>overflow: hidden</code>; the portaled panel still escapes it.
		</p>
		<Popover portal placement="bottom-start">
			{#snippet trigger(api)}
				<Button onclick={api.toggle}>Open (portaled)</Button>
			{/snippet}
			<div class="w-64 p-4 text-sm text-slate-700">Rendered in &lt;body&gt;, not clipped.</div>
		</Popover>
	</div>
</Story>

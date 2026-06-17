<script module lang="ts">
	import { defineMeta } from '@storybook/addon-svelte-csf';
	import { zLayers } from '$lib/styles/z-index';

	// Foundations smoke story (DESIGN_PLAN §8 acceptance): verifies the theme tokens render — taupe
	// background family, the serif heading / sans body fonts, the semantic intent palette, and the
	// documented z-index scale. Because the app loads the same app.css as Storybook, this view should
	// look identical in both.
	const { Story } = defineMeta({
		title: 'Foundations/Theme',
		tags: ['autodocs']
	});

	const taupe = [50, 100, 200, 300, 400, 500, 600, 700, 800, 900];
	const taupeClasses: Record<number, string> = {
		50: 'bg-taupe-50',
		100: 'bg-taupe-100',
		200: 'bg-taupe-200',
		300: 'bg-taupe-300',
		400: 'bg-taupe-400',
		500: 'bg-taupe-500',
		600: 'bg-taupe-600',
		700: 'bg-taupe-700',
		800: 'bg-taupe-800',
		900: 'bg-taupe-900'
	};

	const intents = [
		{ name: 'primary', cls: 'bg-slate-600 text-amber-200' },
		{ name: 'secondary', cls: 'bg-emerald-400 text-slate-800' },
		{ name: 'success', cls: 'bg-emerald-700 text-white' },
		{ name: 'warning', cls: 'bg-amber-500 text-slate-800' },
		{ name: 'error', cls: 'bg-red-700 text-white' },
		{ name: 'info', cls: 'bg-sky-700 text-white' }
	];
</script>

<Story name="Tokens" asChild>
	<div class="space-y-8 p-6">
		<section class="space-y-3">
			<h1 class="text-2xl">Typography — EB Garamond headings over Noto Sans body</h1>
			<p class="max-w-prose text-sm text-slate-700">
				This paragraph is Noto Sans (the body font). The heading above is EB Garamond (the serif
				heading font). If the heading is not serif, the local <code>@font-face</code> failed to load.
			</p>
		</section>

		<section class="space-y-2">
			<h2 class="text-lg">Taupe background family</h2>
			<div class="flex flex-wrap gap-2">
				{#each taupe as step (step)}
					<div class="flex flex-col items-center gap-1">
						<div
							class={['size-16 rounded-xl border border-slate-300', taupeClasses[step]].join(' ')}
						></div>
						<span class="text-[11px] text-slate-500">taupe-{step}</span>
					</div>
				{/each}
			</div>
		</section>

		<section class="space-y-2">
			<h2 class="text-lg">Semantic intents</h2>
			<div class="flex flex-wrap gap-2">
				{#each intents as intent (intent.name)}
					<div
						class={[
							'flex h-16 w-28 items-center justify-center rounded-xl text-sm font-medium',
							intent.cls
						].join(' ')}
					>
						{intent.name}
					</div>
				{/each}
			</div>
		</section>

		<section class="space-y-2">
			<h2 class="text-lg">Z-index scale</h2>
			<ul class="space-y-1 text-sm text-slate-700">
				{#each Object.entries(zLayers) as [name, value] (name)}
					<li><span class="font-mono text-slate-900">{value}</span> — {name}</li>
				{/each}
			</ul>
		</section>
	</div>
</Story>

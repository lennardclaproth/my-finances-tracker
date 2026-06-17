<script lang="ts">
	import Icon from '$lib/components/atoms/icon/Icon.svelte';

	// All token values are mirrored from src/app.css (@theme), src/lib/charts/theme.ts, the component
	// *.variants.ts files, and src/lib/styles/z-index.ts. This page is a living reference — when those
	// sources change, update the matching array here.
	const taupe: [string, string][] = [
		['taupe-50', '#faf8f4'],
		['taupe-100', '#f2ede4'],
		['taupe-200', '#e7dfd1'],
		['taupe-300', '#d8ccb8'],
		['taupe-400', '#bfae93'],
		['taupe-500', '#a48f70'],
		['taupe-600', '#877458'],
		['taupe-700', '#6b5b46'],
		['taupe-800', '#4c4031'],
		['taupe-900', '#2f271e']
	];

	const slate: [string, string][] = [
		['slate-50', '#f8fafc'],
		['slate-100', '#f1f5f9'],
		['slate-200', '#e2e8f0'],
		['slate-300', '#cbd5e1'],
		['slate-400', '#94a3b8'],
		['slate-500', '#64748b'],
		['slate-600', '#475569'],
		['slate-700', '#334155'],
		['slate-800', '#1e293b'],
		['slate-900', '#0f172a'],
		['slate-950', '#020617']
	];

	const amber: [string, string][] = [
		['amber-200', '#fde68a'],
		['amber-300', '#fcd34d'],
		['amber-400', '#fbbf24'],
		['amber-500', '#f59e0b'],
		['amber-600', '#d97706'],
		['amber-800', '#92400e']
	];

	const data: [string, string, string][] = [
		['positive', '#059669', 'income / gain'],
		['negative', '#dc2626', 'expense / loss'],
		['negativeStrong', '#b91c1c', 'strong loss'],
		['net', '#334155', 'net line'],
		['info', '#0284c7', 'informational'],
		['markerBorder', '#fde68a', 'hover marker'],
		['grid', '#e2e8f0', 'grid lines'],
		['zeroLine', '#94a3b8', 'zero axis'],
		['axisText', '#64748b', 'axis labels'],
		['tooltipBg', '#0f172a', 'tooltip surface']
	];

	const intents: [string, string, string, string][] = [
		['primary', '#475569', '#fde68a', 'slate-600 / amber-200'],
		['secondary', '#fde68a', '#1e293b', 'amber-200 / slate-800'],
		['success', '#047857', '#ffffff', 'emerald-700 / white'],
		['warning', '#f59e0b', '#1e293b', 'amber-500 / slate-800'],
		['error', '#b91c1c', '#ffffff', 'red-700 / white'],
		['info', '#0369a1', '#ffffff', 'sky-700 / white'],
		['neutral', '#475569', '#ffffff', 'slate-600 / white']
	];

	const donutIn = ['#059669', '#0ea5e9', '#10b981', '#38bdf8', '#34d399', '#7dd3fc', '#6ee7b7'];
	const donutOut = ['#b91c1c', '#f59e0b', '#dc2626', '#fbbf24', '#ef4444', '#fcd34d', '#f87171'];

	const zLayers: [string, number, string][] = [
		['chartOverlay', 10, 'chart drag overlay'],
		['stickyHeader', 20, 'sticky table header'],
		['popover', 30, 'popover / table footer'],
		['fab', 40, 'FAB / expanded scrim'],
		['modal', 50, 'modal scrim + toast host'],
		['filterPopover', 60, 'portaled filter popovers'],
		['asyncDropdown', 70, 'async search dropdown']
	];
</script>

<svelte:head>
	<title>Tally — Style Guide</title>
	<link rel="icon" href="/brand/favicon.svg" />
</svelte:head>

<div class="sg">
	<!-- Cover -->
	<header class="cover">
		<div class="wrap">
			<p class="kicker">Design System · v0.1</p>
			<h1>Tally<span class="dot">.</span></h1>
			<p class="lede">
				The visual language of a personal finance tracker — warm, calm, and exact. Editorial serif
				headings over a quiet sans body, set on warm taupe paper, with a single gold accent reserved
				for what matters.
			</p>
			<div class="meta">
				<span>Heading · <b>EB Garamond</b></span>
				<span>Body · <b>Noto Sans</b></span>
				<span>Surface · <b>Taupe</b></span>
				<span>Accent · <b>Amber / Gold</b></span>
			</div>
		</div>
	</header>

	<nav class="toc">
		<div class="wrap">
			<a href="#logo">Logo</a>
			<a href="#color">Color</a>
			<a href="#type">Typography</a>
			<a href="#icons">Iconography</a>
			<a href="#components">Components</a>
			<a href="#foundations">Foundations</a>
			<a href="#voice">Voice</a>
		</div>
	</nav>

	<!-- Logo -->
	<section id="logo">
		<div class="wrap">
			<div class="sec-head"><span class="num">01</span><h2>Logo &amp; mark</h2></div>
			<p class="intro">
				Four mark concepts to choose from. <b>Tally Marks</b> is the recommendation — it states what
				the product does (counting, keeping score) and is built from pure gold strokes on slate, so
				it reads at any size. The growth and coin-stack marks are name-independent; the monogram
				follows your chosen initial.
			</p>

			<div class="grid cols-4">
				<div class="card logo-cell">
					<img src="/brand/tally-mark.svg" alt="Tally marks logo" />
					<div><div class="cap">Tally Marks</div><div class="star">★ Recommended</div></div>
				</div>
				<div class="card logo-cell">
					<img src="/brand/monogram.svg" alt="Serif monogram logo" />
					<div class="cap">Serif monogram</div>
				</div>
				<div class="card logo-cell">
					<img src="/brand/growth.svg" alt="Growth bars logo" />
					<div class="cap">Growth bars</div>
				</div>
				<div class="card logo-cell">
					<img src="/brand/ledger.svg" alt="Coin stack logo" />
					<div class="cap">Coin stack</div>
				</div>
			</div>

			<h3 class="block">Horizontal lockup</h3>
			<div class="lockup-row">
				<div class="card"><img src="/brand/logo-horizontal.svg" alt="Tally horizontal logo on light" /></div>
				<div class="card dark"><img src="/brand/logo-horizontal.svg" alt="Tally horizontal logo on dark" /></div>
			</div>

			<h3 class="block">Usage</h3>
			<div class="dd">
				<div class="item do">
					<b>Do</b>
					Keep clear space around the mark equal to the height of one tally stroke. Use the mark on
					taupe, white, or slate-900. Pair gold strokes with a dark slate field only.
				</div>
				<div class="item dont">
					<b>Don't</b>
					Don't recolor the gold to a flat yellow, stretch the lockup, put the mark on a busy photo,
					or drop the dark container behind the gold strokes (they need the contrast).
				</div>
			</div>
		</div>
	</section>

	<!-- Color -->
	<section id="color">
		<div class="wrap">
			<div class="sec-head"><span class="num">02</span><h2>Color</h2></div>
			<p class="intro">
				The palette is three families plus a disciplined data set. <b>Taupe</b> is the warm surface
				you see first; <b>Slate</b> is the ink and primary UI; <b>Amber</b> is the gold accent used
				sparingly for emphasis and brand. Greens and reds are reserved for money meaning — gains and
				losses.
			</p>

			<h3 class="block">Taupe — warm surface family</h3>
			<div class="ramp">
				{#each taupe as [name, hex] (name)}
					<div class="swatch">
						<div class="chip" style="background:{hex}"></div>
						<div class="info"><div class="name">{name}</div><div class="hex">{hex}</div></div>
					</div>
				{/each}
			</div>

			<h3 class="block">Slate — ink &amp; primary UI</h3>
			<div class="ramp">
				{#each slate as [name, hex] (name)}
					<div class="swatch">
						<div class="chip" style="background:{hex}"></div>
						<div class="info"><div class="name">{name}</div><div class="hex">{hex}</div></div>
					</div>
				{/each}
			</div>

			<h3 class="block">Amber — gold accent</h3>
			<div class="ramp">
				{#each amber as [name, hex] (name)}
					<div class="swatch">
						<div class="chip" style="background:{hex}"></div>
						<div class="info"><div class="name">{name}</div><div class="hex">{hex}</div></div>
					</div>
				{/each}
			</div>

			<h3 class="block">Semantic intents</h3>
			<p class="intro" style="margin-top:0">
				Each intent is a background + foreground pairing. Note the signature <b>primary</b>:
				slate-600 with gold text, not a saturated blue.
			</p>
			<div class="grid cols-4">
				{#each intents as [name, bg, fg, pair] (name)}
					<div class="intent-chip" style="background:{bg};color:{fg}">
						<span class="label">{name}</span>
						<span class="pair">{pair}</span>
					</div>
				{/each}
			</div>

			<h3 class="block">Data &amp; finance colors</h3>
			<p class="intro" style="margin-top:0">
				Fixed meanings for charts and figures — never decorative. Source:
				<code>src/lib/charts/theme.ts</code>.
			</p>
			<div class="ramp">
				{#each data as [name, hex, note] (name)}
					<div class="swatch">
						<div class="chip" style="background:{hex}"></div>
						<div class="info">
							<div class="name">{name}</div>
							<div class="hex">{hex}</div>
							<div class="hex" style="text-transform:none">{note}</div>
						</div>
					</div>
				{/each}
			</div>
			<h3 class="block" style="font-size:14px">Donut ramps</h3>
			<div class="grid cols-2">
				<div class="card">
					<div style="font-size:12px;color:var(--muted);margin-bottom:8px">Incoming (emerald → sky)</div>
					<div class="ramp-bar">
						{#each donutIn as c (c)}<div style="flex:1;background:{c}"></div>{/each}
					</div>
				</div>
				<div class="card">
					<div style="font-size:12px;color:var(--muted);margin-bottom:8px">Outgoing (red → amber)</div>
					<div class="ramp-bar">
						{#each donutOut as c (c)}<div style="flex:1;background:{c}"></div>{/each}
					</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Typography -->
	<section id="type">
		<div class="wrap">
			<div class="sec-head"><span class="num">03</span><h2>Typography</h2></div>
			<p class="intro">
				Two typefaces, clear jobs. <b>EB Garamond</b> (serif) carries headings and large figures — it
				is where the editorial warmth lives. <b>Noto Sans</b> carries everything you read in bulk:
				labels, body, tables, controls.
			</p>

			<div class="grid cols-2">
				<div class="card">
					<div style="font-size:12px;color:var(--muted)">Heading face</div>
					<div class="heading-font" style="font-size:56px;line-height:1.1">EB Garamond</div>
					<div class="heading-font" style="font-size:20px;color:var(--slate-600)">Aa Bb Cc · 0123456789 · €1,240.50</div>
				</div>
				<div class="card">
					<div style="font-size:12px;color:var(--muted)">Body face</div>
					<div class="body-font" style="font-size:48px;line-height:1.15;font-weight:600">Noto Sans</div>
					<div class="body-font" style="font-size:18px;color:var(--slate-600)">Aa Bb Cc · 0123456789 · €1,240.50</div>
				</div>
			</div>

			<h3 class="block">Heading scale — EB Garamond</h3>
			<div class="card">
				<div class="specimen"><div class="tag"><b>2xl</b>text-4xl · 36px</div><div class="heading-font" style="font-size:36px">Net worth at a glance</div></div>
				<div class="specimen"><div class="tag"><b>xl</b>text-3xl · 30px</div><div class="heading-font" style="font-size:30px">Net worth at a glance</div></div>
				<div class="specimen"><div class="tag"><b>lg</b>text-2xl · 24px</div><div class="heading-font" style="font-size:24px">Net worth at a glance</div></div>
				<div class="specimen"><div class="tag"><b>md</b>text-xl · 20px</div><div class="heading-font" style="font-size:20px">Net worth at a glance</div></div>
				<div class="specimen"><div class="tag"><b>sm</b>text-lg · 18px</div><div class="heading-font" style="font-size:18px">Net worth at a glance</div></div>
			</div>

			<h3 class="block">Body scale — Noto Sans</h3>
			<div class="card">
				<div class="specimen"><div class="tag"><b>lg</b>text-lg · 18px</div><div class="body-font" style="font-size:18px">Every transaction, categorised and reconciled.</div></div>
				<div class="specimen"><div class="tag"><b>md</b>text-base · 16px</div><div class="body-font" style="font-size:16px">Every transaction, categorised and reconciled.</div></div>
				<div class="specimen"><div class="tag"><b>sm</b>text-sm · 14px</div><div class="body-font" style="font-size:14px">Every transaction, categorised and reconciled.</div></div>
				<div class="specimen"><div class="tag"><b>xs</b>text-xs · 12px</div><div class="body-font" style="font-size:12px">Every transaction, categorised and reconciled.</div></div>
			</div>

			<h3 class="block">Text tones</h3>
			<div class="card tone-row">
				<span style="color:var(--slate-700)">Default — slate-700</span>
				<span style="color:var(--slate-500)">Muted — slate-500</span>
				<span style="color:var(--slate-950);font-weight:600">Strong — slate-950</span>
				<span style="color:#047857;font-weight:600">Success — emerald-700</span>
				<span style="color:#b91c1c;font-weight:600">Danger — red-700</span>
			</div>
		</div>
	</section>

	<!-- Iconography -->
	<section id="icons">
		<div class="wrap">
			<div class="sec-head"><span class="num">04</span><h2>Iconography</h2></div>
			<p class="intro">
				Icons come from <b>Iconify</b> via <code>@iconify/svelte</code>, wrapped by the
				<code>Icon</code> atom. The house set is <b>Heroicons</b> (ids like
				<code>heroicons:check</code>). Icons inherit <code>currentColor</code> and a fixed
				<code>sm/md/lg</code> size; they are line-weight, never filled illustrations. Atoms receive
				icons through <code>children</code>; icon-only controls are molecules.
			</p>
			<div class="card icon-row">
				<Icon icon="heroicons:banknotes" size="lg" />
				<Icon icon="heroicons:chart-bar" size="lg" />
				<Icon icon="heroicons:arrow-trending-up" size="lg" />
				<Icon icon="heroicons:wallet" size="lg" />
				<Icon icon="heroicons:arrow-path" size="lg" />
				<Icon icon="heroicons:check-circle" size="lg" />
				<span style="font-size:13px;color:var(--muted)">Rendered live through the <code>Icon</code> atom.</span>
			</div>
		</div>
	</section>

	<!-- Components -->
	<section id="components">
		<div class="wrap">
			<div class="sec-head"><span class="num">05</span><h2>Components</h2></div>
			<p class="intro">
				A preview of the atomic-design library (<code>web/src/lib/components/</code>). These are
				static reproductions of the live <code>variants.ts</code> tokens — the running components and
				their full state matrix live in Storybook.
			</p>

			<h3 class="block">Buttons — intents</h3>
			<div class="card demo">
				<button class="btn primary">Primary</button>
				<button class="btn secondary">Secondary</button>
				<button class="btn success">Success</button>
				<button class="btn warning">Warning</button>
				<button class="btn error">Error</button>
				<button class="btn info">Info</button>
			</div>

			<h3 class="block">Buttons — variants &amp; shapes</h3>
			<div class="card demo">
				<button class="btn primary">Solid</button>
				<button class="btn outline">Outline</button>
				<button class="btn ghost">Ghost</button>
				<button class="btn primary pill">Pill</button>
				<button class="btn primary" style="height:32px;padding:0 8px;font-size:13px">Small</button>
				<button class="btn primary" style="height:48px;font-size:16px">Large</button>
			</div>

			<h3 class="block">Badges</h3>
			<div class="card demo">
				<span class="badge neutral"><span class="d" style="background:var(--slate-500)"></span>Neutral</span>
				<span class="badge primary"><span class="d" style="background:var(--amber-200)"></span>Primary</span>
				<span class="badge success">Reconciled</span>
				<span class="badge warning">Pending</span>
				<span class="badge error">Failed</span>
				<span class="badge info">Imported</span>
			</div>

			<h3 class="block">Surfaces &amp; figures</h3>
			<div class="grid cols-3">
				<div class="stat-card">
					<div class="lbl">Net worth</div>
					<div class="val">€124,580</div>
					<div class="delta">▲ 3.2% this month</div>
				</div>
				<div class="stat-card">
					<div class="lbl">Spending</div>
					<div class="val">€2,140</div>
					<div class="delta down">▼ 8.0% vs. budget</div>
				</div>
				<div class="card">
					<div style="font-size:13px;color:var(--muted);margin-bottom:8px">Money figures</div>
					<div style="font-size:20px"><span class="num-pos">+€1,250.00</span></div>
					<div style="font-size:20px"><span class="num-neg">−€430.18</span></div>
					<div style="font-size:20px;color:var(--slate-700);font-weight:600">€0.00</div>
				</div>
			</div>
		</div>
	</section>

	<!-- Foundations -->
	<section id="foundations">
		<div class="wrap">
			<div class="sec-head"><span class="num">06</span><h2>Foundations</h2></div>
			<p class="intro">The structural tokens shared across every component.</p>

			<div class="grid cols-2">
				<div class="card">
					<h3 class="block" style="margin-top:0">Radius</h3>
					<div class="demo">
						<div class="radius-box" style="border-radius:6px"></div>
						<div class="radius-box" style="border-radius:12px"></div>
						<div class="radius-box" style="border-radius:16px"></div>
						<div class="radius-box" style="width:48px;border-radius:9999px"></div>
					</div>
					<div class="foot-note">md 6px (controls) · xl 12px (panels) · 2xl 16px · full (pills, avatars)</div>
				</div>

				<div class="card">
					<h3 class="block" style="margin-top:0">Elevation &amp; motion</h3>
					<div class="demo">
						<div class="elev" style="box-shadow:none"></div>
						<div class="elev" style="box-shadow:0 1px 2px rgba(15,23,42,.06)"></div>
						<div class="elev" style="box-shadow:0 4px 10px rgba(15,23,42,.10)"></div>
					</div>
					<div class="foot-note">
						shadow none · sm · md. Transitions <code>150ms ease-out</code>; pressed controls scale to
						<code>0.98</code>.
					</div>
				</div>
			</div>

			<h3 class="block">Z-index scale</h3>
			<div class="card">
				<table class="kv">
					<thead><tr><th>Layer</th><th>Value</th><th>Use</th></tr></thead>
					<tbody>
						{#each zLayers as [name, val, use] (name)}
							<tr><td><span class="mono">{name}</span></td><td><span class="mono">{val}</span></td><td>{use}</td></tr>
						{/each}
					</tbody>
				</table>
			</div>

			<h3 class="block">Spacing</h3>
			<div class="card">
				<p style="margin:0;font-size:14px;color:var(--slate-600)">
					Tailwind's 4px base scale. Component padding lands on <code>p-3</code> (12px),
					<code>p-4</code> (16px), and <code>p-6</code> (24px); gaps prefer
					<code>gap-1</code>–<code>gap-4</code>.
				</p>
			</div>
		</div>
	</section>

	<!-- Voice -->
	<section id="voice">
		<div class="wrap">
			<div class="sec-head"><span class="num">07</span><h2>Voice &amp; principles</h2></div>
			<div class="dd">
				<div class="item do">
					<b>Calm &amp; exact</b>
					Money is sensitive. State numbers plainly, round consistently, and never use alarm where a
					quiet figure will do. Gold means "noteworthy", not "warning".
				</div>
				<div class="item do">
					<b>Warm, not corporate</b>
					Taupe paper and a serif voice make the tool feel personal — like a well-kept ledger, not a
					bank dashboard. Lead with plain language over jargon.
				</div>
			</div>
			<div class="dd" style="margin-top:16px">
				<div class="item dont">
					<b>Avoid alarm by default</b>
					Don't paint the whole screen red on a loss — red and green carry meaning only on the
					figures themselves.
				</div>
				<div class="item dont">
					<b>Avoid accent inflation</b>
					Don't use gold for more than one thing per view. If everything is highlighted, nothing is.
				</div>
			</div>
		</div>
	</section>

	<footer>
		<div class="wrap">
			<span>Tally — Design System v0.1 · <code>my-finances-tracker</code></span>
			<span>Tokens: <code>src/app.css</code> · <code>charts/theme.ts</code> · component <code>*.variants.ts</code></span>
		</div>
	</footer>
</div>

<style>
	.sg {
		--font-sans: 'Noto Sans', ui-sans-serif, system-ui, sans-serif;
		--font-heading: 'EB Garamond', ui-serif, Georgia, serif;

		--taupe-100: #f2ede4;
		--taupe-300: #d8ccb8;
		--taupe-50: #faf8f4;

		--slate-100: #f1f5f9;
		--slate-200: #e2e8f0;
		--slate-300: #cbd5e1;
		--slate-400: #94a3b8;
		--slate-500: #64748b;
		--slate-600: #475569;
		--slate-700: #334155;
		--slate-800: #1e293b;
		--slate-900: #0f172a;
		--slate-950: #020617;

		--amber-200: #fde68a;
		--amber-500: #f59e0b;
		--amber-600: #d97706;
		--amber-800: #92400e;

		--ink: var(--slate-800);
		--ink-strong: var(--slate-900);
		--muted: var(--slate-500);
		--paper: var(--taupe-100);
		--card: #ffffff;
		--hairline: var(--slate-300);
		--gold: var(--amber-200);

		background: var(--paper);
		color: var(--ink);
		font-family: var(--font-sans);
		line-height: 1.6;
		min-height: 100vh;
	}

	.sg h1,
	.sg h2,
	.sg h3 {
		font-family: var(--font-heading);
		color: var(--ink-strong);
		letter-spacing: -0.01em;
		line-height: 1.15;
		font-weight: 600;
		margin: 0;
	}

	.sg code {
		font-family: ui-monospace, 'SFMono-Regular', 'Cascadia Code', Menlo, Consolas, monospace;
		font-size: 0.82em;
	}
	.sg .mono {
		font-family: ui-monospace, 'SFMono-Regular', 'Cascadia Code', Menlo, Consolas, monospace;
		font-size: 0.82em;
	}

	.sg a {
		color: var(--slate-700);
	}

	.sg .wrap {
		max-width: 1080px;
		margin: 0 auto;
		padding: 0 24px;
	}

	/* Cover */
	.sg .cover {
		background: radial-gradient(120% 140% at 80% 0%, #1e293b 0%, var(--slate-900) 60%);
		color: var(--slate-100);
		padding: 72px 0 64px;
		border-bottom: 3px solid var(--gold);
	}
	.sg .cover .kicker {
		font-family: var(--font-sans);
		text-transform: uppercase;
		letter-spacing: 0.28em;
		font-size: 12px;
		font-weight: 600;
		color: var(--amber-200);
		margin: 0 0 18px;
	}
	.sg .cover h1 {
		color: #fff;
		font-size: clamp(44px, 8vw, 84px);
	}
	.sg .cover h1 .dot {
		color: var(--amber-500);
	}
	.sg .cover .lede {
		max-width: 56ch;
		margin: 20px 0 0;
		color: var(--slate-300);
		font-size: 18px;
	}
	.sg .cover .meta {
		margin-top: 32px;
		display: flex;
		gap: 10px 28px;
		flex-wrap: wrap;
		font-size: 13px;
		color: var(--slate-400);
	}
	.sg .cover .meta b {
		color: var(--amber-200);
		font-weight: 600;
	}

	/* Sticky local nav */
	.sg nav.toc {
		position: sticky;
		top: 0;
		z-index: 50;
		background: rgba(242, 237, 228, 0.85);
		backdrop-filter: blur(8px);
		border-bottom: 1px solid var(--taupe-300);
	}
	.sg nav.toc .wrap {
		display: flex;
		gap: 22px;
		flex-wrap: wrap;
		padding-top: 14px;
		padding-bottom: 14px;
	}
	.sg nav.toc a {
		text-decoration: none;
		color: var(--slate-600);
		font-size: 13px;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}
	.sg nav.toc a:hover {
		color: var(--ink-strong);
	}

	/* Sections */
	.sg section {
		padding: 56px 0;
		border-bottom: 1px solid var(--taupe-300);
	}
	.sg .sec-head {
		display: flex;
		align-items: baseline;
		gap: 14px;
		margin-bottom: 8px;
	}
	.sg .sec-head .num {
		font-family: var(--font-heading);
		font-style: italic;
		color: var(--amber-600);
		font-size: 20px;
	}
	.sg section h2 {
		font-size: 30px;
	}
	.sg section .intro {
		color: var(--slate-600);
		max-width: 64ch;
		margin: 6px 0 32px;
	}
	.sg h3.block {
		font-size: 18px;
		margin: 36px 0 14px;
	}
	.sg h3.block:first-of-type {
		margin-top: 8px;
	}

	/* Cards / grids */
	.sg .card {
		background: var(--card);
		border: 1px solid var(--hairline);
		border-radius: 12px;
		padding: 20px;
		box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
	}
	.sg .grid {
		display: grid;
		gap: 16px;
	}
	.sg .cols-2 {
		grid-template-columns: repeat(2, 1fr);
	}
	.sg .cols-3 {
		grid-template-columns: repeat(3, 1fr);
	}
	.sg .cols-4 {
		grid-template-columns: repeat(4, 1fr);
	}
	@media (max-width: 860px) {
		.sg .cols-3,
		.sg .cols-4 {
			grid-template-columns: repeat(2, 1fr);
		}
	}
	@media (max-width: 560px) {
		.sg .cols-2,
		.sg .cols-3,
		.sg .cols-4 {
			grid-template-columns: 1fr;
		}
	}

	/* Swatches */
	.sg .ramp {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(92px, 1fr));
		gap: 10px;
	}
	.sg .swatch {
		border-radius: 10px;
		overflow: hidden;
		border: 1px solid var(--taupe-300);
		background: #fff;
	}
	.sg .swatch .chip {
		height: 56px;
	}
	.sg .swatch .info {
		padding: 7px 9px;
		font-size: 11px;
		line-height: 1.45;
	}
	.sg .swatch .info .name {
		font-weight: 600;
		color: var(--ink-strong);
	}
	.sg .swatch .info .hex {
		color: var(--muted);
		text-transform: uppercase;
	}
	.sg .ramp-bar {
		display: flex;
		height: 28px;
		border-radius: 8px;
		overflow: hidden;
	}

	.sg .intent-chip {
		border-radius: 10px;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-height: 92px;
		justify-content: space-between;
		border: 1px solid rgba(15, 23, 42, 0.06);
	}
	.sg .intent-chip .label {
		font-weight: 600;
		font-size: 14px;
	}
	.sg .intent-chip .pair {
		font-size: 11px;
		opacity: 0.8;
	}

	/* Logo gallery */
	.sg .logo-cell {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 14px;
		padding: 28px 20px;
		text-align: center;
	}
	.sg .logo-cell.dark {
		background: var(--slate-900);
		border-color: var(--slate-800);
	}
	.sg .logo-cell img {
		width: 96px;
		height: 96px;
	}
	.sg .logo-cell .cap {
		font-size: 12px;
		color: var(--muted);
	}
	.sg .logo-cell.dark .cap {
		color: var(--slate-400);
	}
	.sg .logo-cell .star {
		font-size: 11px;
		color: var(--amber-600);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.08em;
	}
	.sg .lockup-row {
		display: flex;
		flex-wrap: wrap;
		gap: 20px;
	}
	.sg .lockup-row .card {
		flex: 1 1 320px;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 32px;
	}
	.sg .lockup-row .card.dark {
		background: var(--slate-900);
		border-color: var(--slate-800);
	}
	.sg .lockup-row img {
		max-width: 100%;
		height: 64px;
	}

	/* Type specimens */
	.sg .specimen {
		display: flex;
		align-items: baseline;
		gap: 18px;
		padding: 12px 0;
		border-bottom: 1px dashed var(--taupe-300);
	}
	.sg .specimen:last-child {
		border-bottom: 0;
	}
	.sg .specimen .tag {
		width: 150px;
		flex: none;
		font-size: 11px;
		color: var(--muted);
	}
	.sg .specimen .tag b {
		display: block;
		color: var(--ink-strong);
		font-size: 12px;
	}
	.sg .heading-font {
		font-family: var(--font-heading);
		color: var(--ink-strong);
	}
	.sg .body-font {
		font-family: var(--font-sans);
	}
	.sg .tone-row {
		display: flex;
		flex-wrap: wrap;
		gap: 18px;
	}
	.sg .tone-row span {
		font-size: 15px;
	}

	/* Iconography */
	.sg .icon-row {
		display: flex;
		flex-wrap: wrap;
		gap: 22px;
		align-items: center;
		color: var(--slate-700);
	}

	/* Component demos */
	.sg .demo {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
		align-items: center;
	}
	.sg .btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		height: 40px;
		padding: 0 14px;
		border-radius: 6px;
		font-size: 14px;
		font-weight: 500;
		border: 1px solid transparent;
		cursor: default;
		transition: all 150ms ease-out;
		font-family: var(--font-sans);
	}
	.sg .btn.primary {
		background: var(--slate-600);
		color: var(--amber-200);
		border-color: var(--amber-200);
	}
	.sg .btn.secondary {
		background: var(--amber-200);
		color: var(--slate-800);
		border-color: var(--slate-800);
	}
	.sg .btn.success {
		background: #047857;
		color: #fff;
	}
	.sg .btn.warning {
		background: var(--amber-500);
		color: var(--slate-800);
	}
	.sg .btn.error {
		background: #b91c1c;
		color: #fff;
	}
	.sg .btn.info {
		background: #0369a1;
		color: #fff;
	}
	.sg .btn.outline {
		background: transparent;
		color: var(--slate-700);
		border-color: var(--slate-700);
	}
	.sg .btn.ghost {
		background: transparent;
		color: var(--slate-700);
		border-color: transparent;
	}
	.sg .btn.pill {
		border-radius: 9999px;
	}

	.sg .badge {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 3px 10px;
		border-radius: 9999px;
		font-size: 12px;
		font-weight: 500;
		border: 1px solid transparent;
	}
	.sg .badge .d {
		width: 6px;
		height: 6px;
		border-radius: 9999px;
	}
	.sg .badge.neutral {
		background: var(--slate-100);
		color: var(--slate-700);
		border-color: var(--slate-200);
	}
	.sg .badge.primary {
		background: var(--slate-100);
		color: var(--slate-700);
		border-color: var(--amber-200);
	}
	.sg .badge.success {
		background: #d1fae5;
		color: #065f46;
		border-color: #a7f3d0;
	}
	.sg .badge.warning {
		background: #fef3c7;
		color: var(--amber-800);
		border-color: #fde68a;
	}
	.sg .badge.error {
		background: #fee2e2;
		color: #b91c1c;
		border-color: #fecaca;
	}
	.sg .badge.info {
		background: #e0f2fe;
		color: #075985;
		border-color: #bae6fd;
	}

	.sg .stat-card {
		border: 1px solid var(--hairline);
		border-radius: 12px;
		background: #fff;
		padding: 18px 20px;
		max-width: 280px;
	}
	.sg .stat-card .lbl {
		font-size: 12px;
		color: var(--muted);
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}
	.sg .stat-card .val {
		font-family: var(--font-heading);
		font-size: 30px;
		color: var(--ink-strong);
		margin: 4px 0;
	}
	.sg .stat-card .delta {
		font-size: 13px;
		font-weight: 600;
		color: #047857;
	}
	.sg .stat-card .delta.down {
		color: #b91c1c;
	}

	.sg .num-pos {
		color: #059669;
		font-weight: 600;
	}
	.sg .num-neg {
		color: #dc2626;
		font-weight: 600;
	}

	.sg .radius-box {
		width: 64px;
		height: 48px;
		background: var(--slate-200);
		border: 1px solid var(--slate-300);
	}
	.sg .elev {
		width: 80px;
		height: 48px;
		background: #fff;
		border: 1px solid var(--slate-200);
		border-radius: 10px;
	}
	.sg .foot-note {
		font-size: 12px;
		color: var(--muted);
		margin-top: 10px;
	}

	/* Do / Don't */
	.sg .dd {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 16px;
	}
	@media (max-width: 560px) {
		.sg .dd {
			grid-template-columns: 1fr;
		}
	}
	.sg .dd .item {
		border-radius: 10px;
		padding: 14px 16px;
		font-size: 14px;
	}
	.sg .dd .do {
		background: #ecfdf5;
		border: 1px solid #a7f3d0;
	}
	.sg .dd .dont {
		background: #fef2f2;
		border: 1px solid #fecaca;
	}
	.sg .dd .item b {
		display: block;
		margin-bottom: 4px;
	}
	.sg .dd .do b {
		color: #047857;
	}
	.sg .dd .dont b {
		color: #b91c1c;
	}

	.sg .kv {
		width: 100%;
		border-collapse: collapse;
		font-size: 14px;
	}
	.sg .kv th,
	.sg .kv td {
		text-align: left;
		padding: 9px 12px;
		border-bottom: 1px solid var(--taupe-300);
		vertical-align: top;
	}
	.sg .kv th {
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		color: var(--muted);
		font-weight: 600;
	}
	.sg .kv td .mono {
		color: var(--ink-strong);
	}

	.sg footer {
		padding: 40px 0 64px;
		color: var(--muted);
		font-size: 13px;
	}
	.sg footer .wrap {
		display: flex;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: 12px;
	}
</style>

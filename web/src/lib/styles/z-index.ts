/**
 * Single documented z-index scale for the portal's overlay layers (DESIGN_PLAN §2.6).
 *
 * The Vue reference stacked overlays ad-hoc (and had inversions); here the layering is one ordered
 * scale that every overlay/popover/modal/toast pulls from. Each entry pairs the numeric step with the
 * matching Tailwind utility (kept as a complete literal so the class scanner can see it).
 *
 * Lower sits beneath higher:
 *   chart drag overlay (10) < sticky table header (20) < popover / table footer (30)
 *   < FAB / expanded scrim (40) < modal scrim + toast host (50)
 *   < portaled filter popovers (60) < async search dropdown (70)
 */
export const zLayers = {
	chartOverlay: 10,
	stickyHeader: 20,
	popover: 30,
	fab: 40,
	modal: 50,
	filterPopover: 60,
	asyncDropdown: 70
} as const;

export type ZLayer = keyof typeof zLayers;

/** Tailwind `z-*` utility per layer. Complete literals so Tailwind's content scanner keeps them. */
export const zClasses = {
	chartOverlay: 'z-10',
	stickyHeader: 'z-20',
	popover: 'z-30',
	fab: 'z-40',
	modal: 'z-50',
	filterPopover: 'z-[60]',
	asyncDropdown: 'z-[70]'
} as const satisfies Record<ZLayer, string>;

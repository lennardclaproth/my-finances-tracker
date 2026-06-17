/** One navigable destination in the {@link NavMenu} roll-down. */
export interface NavItem {
	label: string;
	/** Route to navigate to on select. */
	href: string;
	/** Iconify id rendered before the label. */
	icon?: string;
	/** Render a separator above this item. */
	divider?: boolean;
}

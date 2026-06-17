/** One selectable row in a menu. */
export interface MenuItem {
	label: string;
	/** Iconify id rendered before the label. */
	icon?: string;
	onSelect?: () => void;
	/** `danger` renders the item in the error tone. */
	intent?: 'default' | 'danger';
	disabled?: boolean;
	/** Render a separator above this item. */
	divider?: boolean;
}

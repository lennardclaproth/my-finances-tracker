export const popoverPlacements = [
	'top',
	'top-start',
	'top-end',
	'bottom',
	'bottom-start',
	'bottom-end',
	'left',
	'left-start',
	'left-end',
	'right',
	'right-start',
	'right-end'
] as const;

export type PopoverPlacement = (typeof popoverPlacements)[number];

/** Control surface handed to the `trigger` snippet so it can open/close/toggle the popover. */
export interface PopoverApi {
	readonly open: boolean;
	toggle: () => void;
	show: () => void;
	close: () => void;
}

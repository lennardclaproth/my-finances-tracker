export const tabsSizes = ['sm', 'md', 'lg'] as const;

export type TabsSize = (typeof tabsSizes)[number];

/** One tab in a segmented control. */
export interface TabItem {
	value: string;
	label: string;
	disabled?: boolean;
}

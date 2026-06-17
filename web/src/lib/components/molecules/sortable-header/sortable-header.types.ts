export const sortDirections = ['asc', 'desc'] as const;

export type SortDirection = (typeof sortDirections)[number];

export const sortableHeaderAligns = ['left', 'center', 'right'] as const;

export type SortableHeaderAlign = (typeof sortableHeaderAligns)[number];

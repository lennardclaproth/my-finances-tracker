export const dividerOrientations = ['horizontal', 'vertical'] as const;

export type DividerOrientation = (typeof dividerOrientations)[number];

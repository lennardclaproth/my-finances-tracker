export const skeletonVariants = ['text', 'rect', 'circle'] as const;

export type SkeletonVariant = (typeof skeletonVariants)[number];

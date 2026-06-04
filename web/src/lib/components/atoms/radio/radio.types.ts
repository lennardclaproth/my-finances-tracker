export const radioSizes = ['sm', 'md', 'lg'] as const;

export type RadioSize = (typeof radioSizes)[number];

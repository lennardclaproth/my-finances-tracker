export const spinnerSizes = ['sm', 'md', 'lg'] as const;

export type SpinnerSize = (typeof spinnerSizes)[number];

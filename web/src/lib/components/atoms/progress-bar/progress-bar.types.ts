export const progressBarSizes = ['sm', 'md', 'lg'] as const;

export const progressBarIntents = ['primary', 'secondary', 'success', 'warning', 'error'] as const;

export type ProgressBarSize = (typeof progressBarSizes)[number];
export type ProgressBarIntent = (typeof progressBarIntents)[number];

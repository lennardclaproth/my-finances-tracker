export const switchSizes = ['sm', 'md', 'lg'] as const;

export const switchIntents = ['primary', 'secondary', 'success', 'warning', 'error'] as const;

export type SwitchSize = (typeof switchSizes)[number];
export type SwitchIntent = (typeof switchIntents)[number];

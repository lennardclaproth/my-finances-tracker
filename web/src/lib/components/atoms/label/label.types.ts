export const labelSizes = ['sm', 'md', 'lg'] as const;

export const labelTones = ['default', 'muted'] as const;

export type LabelSize = (typeof labelSizes)[number];
export type LabelTone = (typeof labelTones)[number];

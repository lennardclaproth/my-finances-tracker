export const linkTones = ['default', 'primary', 'muted'] as const;

export const linkSizes = ['sm', 'md', 'lg'] as const;

export const linkUnderlines = ['always', 'hover', 'none'] as const;

export type LinkTone = (typeof linkTones)[number];
export type LinkSize = (typeof linkSizes)[number];
export type LinkUnderline = (typeof linkUnderlines)[number];

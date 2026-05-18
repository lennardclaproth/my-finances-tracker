export const textElements = ['p', 'span', 'div'] as const;

export const textSizes = ['xs', 'sm', 'md', 'lg'] as const;

export const textTones = [
  'default',
  'muted',
  'subtle',
  'strong',
  'danger',
  'success'
] as const;

export const textWeights = [
  'normal',
  'medium',
  'semibold',
  'bold'
] as const;

export const headingLevels = ['h1', 'h2', 'h3', 'h4', 'h5', 'h6'] as const;

export const headingSizes = ['sm', 'md', 'lg', 'xl', '2xl'] as const;

export const headingTones = ['default', 'muted', 'subtle'] as const;

export const headingWeights = ['medium', 'semibold', 'bold'] as const;

export type TextElement = typeof textElements[number];
export type TextSize = typeof textSizes[number];
export type TextTone = typeof textTones[number];
export type TextWeight = typeof textWeights[number];

export type HeadingLevel = typeof headingLevels[number];
export type HeadingSize = typeof headingSizes[number];
export type HeadingTone = typeof headingTones[number];
export type HeadingWeight = typeof headingWeights[number];
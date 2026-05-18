// button.types.ts

export const buttonIntents = [
  'primary',
  'secondary',
  'warning',
  'error',
  'success',
  'info'
] as const;

export const buttonVariants = [
  'solid',
  'outline',
  'ghost'
] as const;

export const buttonSizes = [
  'sm',
  'md',
  'lg'
] as const;

export const buttonShapes = [
  'default',
  'rounded',
  'pill'
] as const;

export const iconPositions = [
  'left',
  'right'
] as const;

export type ButtonIntent = typeof buttonIntents[number];
export type ButtonVariant = typeof buttonVariants[number];
export type ButtonSize = typeof buttonSizes[number];
export type ButtonShape = typeof buttonShapes[number];
export type IconPosition = typeof iconPositions[number];

export type ButtonType = 'button' | 'submit' | 'reset';
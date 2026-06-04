export const inputSizes = ['sm', 'md', 'lg'] as const;

export const inputIntents = ['default', 'error', 'success'] as const;

export const inputShapes = ['default', 'rounded', 'pill'] as const;

export type InputSize = typeof inputSizes[number];
export type InputIntent = typeof inputIntents[number];
export type InputShape = typeof inputShapes[number];

export type InputType =
  | 'text'
  | 'email'
  | 'password'
  | 'number'
  | 'search'
  | 'tel'
  | 'url'
  | 'date'
  | 'datetime-local'
  | 'time';
import type { InputIntent, InputSize } from '$lib/components/atoms/input/input.types';

export const inputIconPaddingClasses = {
  sm: {
    left: 'pl-8',
    right: 'pr-8'
  },
  md: {
    left: 'pl-9',
    right: 'pr-9'
  },
  lg: {
    left: 'pl-10',
    right: 'pr-10'
  }
} satisfies Record<InputSize, Record<'left' | 'right', string>>;

export const inputIconContainerSizeClasses = {
  sm: 'h-8 w-8',
  md: 'h-10 w-9',
  lg: 'h-12 w-10'
} satisfies Record<InputSize, string>;

export const inputIconSizeClasses = {
  sm: 'sm',
  md: 'md',
  lg: 'md'
} as const;

export const inputIconIntentClasses = {
  default: 'text-slate-400',
  error: 'text-slate-400',
  success: 'text-slate-400'
} satisfies Record<InputIntent, string>;

export const inputValidationIconIntentClasses = {
  default: 'text-slate-400',
  error: 'text-red-400 group-focus-within:text-red-600',
  success: 'text-emerald-400 group-focus-within:text-emerald-600'
} satisfies Record<InputIntent, string>;

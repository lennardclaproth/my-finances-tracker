import type { InputIntent, InputShape, InputSize } from './input.types';

export const baseInputClasses = [
  'bg-white border-slate-300 text-slate-900',
  'block w-full',
  'border outline-none',
  'transition-colors duration-150 ease-out',
  'placeholder:text-slate-400',
  'disabled:cursor-not-allowed disabled:opacity-50',
  'read-only:cursor-default',
].join(' ');

export const inputSizeClasses = {
  sm: 'h-8 px-2 text-sm',
  md: 'h-10 px-3 text-sm',
  lg: 'h-12 px-4 text-base'
} satisfies Record<InputSize, string>;

export const inputShapeClasses = {
  default: 'rounded-md',
  rounded: 'rounded-xl',
  pill: 'rounded-full'
} satisfies Record<InputShape, string>;

export const inputIntentClasses = {
  default:
    'focus-visible:border-slate-500 focus-visible:ring-slate-200 focus-visible:ring-2',
  error:
    'border-red-600 ring-2 ring-red-200 focus:border-red-700 focus:ring-red-400',
  success:
    'border-emerald-600 ring-2 ring-emerald-200 focus:border-emerald-700 focus:ring-emerald-400'
} satisfies Record<InputIntent, string>;

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
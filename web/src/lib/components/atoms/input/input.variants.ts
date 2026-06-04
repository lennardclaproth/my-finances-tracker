import type { InputIntent, InputShape, InputSize } from './input.types';

// Base input styling. Icon/affix-related class maps live with the IconInput molecule, since icons
// are composed there rather than in this atom.

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
import type { ButtonIntent, ButtonSize, ButtonVariant, ButtonShape } from './button.types';

export const baseButtonClasses = [
  'inline-flex items-center justify-center gap-1',
  'select-none whitespace-nowrap',
  'transition-all duration-150 ease-out',
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2',
  'disabled:pointer-events-none disabled:opacity-50',
  'active:scale-[0.98]'
].join(' ');

export const buttonSizeClasses = {
  sm: 'h-8 px-2 text-sm',
  md: 'h-10 px-3 text-sm',
  lg: 'h-12 px-3 text-base'
} satisfies Record<ButtonSize, string>;

export const iconButtonSizeClasses = {
  sm: 'h-8 w-8 text-sm',
  md: 'h-10 w-10 text-sm',
  lg: 'h-12 w-12 text-base'
} satisfies Record<ButtonSize, string>;

export const iconSizeClasses = {
  sm: 'size-2',
  md: 'size-3',
  lg: 'size-4'
} satisfies Record<ButtonSize, string>;

export const buttonShapeClasses = {
  default: 'rounded-md',
  rounded: 'rounded-xl',
  pill: 'rounded-full'
} satisfies Record<ButtonShape, string>;

export const intentVariantClasses = {
  primary: {
    solid:
      'bg-slate-600 text-amber-200 hover:bg-slate-500 border-1 border-amber-200 focus:ring-slate-300 focus:ring-2 focus:bg-slate-600 active:bg-slate-600',
    outline:
      'border border-slate-700 text-slate-700 hover:bg-slate-100 focus:ring-slate-50 focus:ring-2 focus:bg-slate-200 active:bg-slate-200',
    ghost:
      'text-slate-700 hover:bg-slate-100 focus:ring-slate-50 focus:ring-2 focus:bg-slate-200 active:bg-slate-200'
  },

  secondary: {
    solid:
      'bg-emerald-400 text-slate-800 border-1 border-slate-800 hover:bg-emerald-500 focus:ring-emerald-200 focus:ring-2 focus:bg-emerald-400 active:bg-emerald-400',
    outline:
      'border border-slate-800 text-slate-800 hover:bg-emerald-100 focus:ring-emerald-50 focus:ring-2 focus:bg-emerald-200 active:bg-emerald-200',
    ghost:
      'text-slate-800 hover:bg-emerald-100 focus:ring-emerald-50 focus:ring-2 focus:bg-emerald-200 active:bg-emerald-200'
  },

  warning: {
    solid:
      'bg-amber-500 text-slate-800 hover:bg-amber-400 focus:ring-amber-300 focus:ring-2 focus:bg-amber-500 active:bg-amber-500',
    outline:
      'border border-amber-600 text-amber-800 hover:bg-amber-100 focus:ring-amber-50 focus:ring-2 focus:bg-amber-200 active:bg-amber-200',
    ghost:
      'text-amber-800 hover:bg-amber-100 focus:ring-amber-50 focus:ring-2 focus:bg-amber-200 active:bg-amber-200'
  },

  error: {
    solid:
      'bg-red-700 text-white hover:bg-red-600 focus:ring-red-300 focus:ring-2 focus:bg-red-700 active:bg-red-700',
    outline:
      'border border-red-600 text-red-700 hover:bg-red-100 focus:ring-red-50 focus:ring-2 focus:bg-red-200 active:bg-red-200',
    ghost:
      'text-red-700 hover:bg-red-50 focus:ring-red-50 focus:ring-2 focus:bg-red-200 active:bg-red-200'
  },

  success: {
    solid:
      'bg-emerald-700 text-white hover:bg-emerald-500 focus:ring-emerald-400 focus:ring-2 focus:bg-emerald-700 active:bg-emerald-700',
    outline:
      'border border-emerald-700 text-emerald-800 hover:bg-emerald-100 focus:ring-emerald-50 focus:ring-2 focus:bg-emerald-200 active:bg-emerald-200',
    ghost:
      'text-emerald-800 hover:bg-emerald-100 focus:ring-emerald-50 focus:ring-2 focus:bg-emerald-200 active:bg-emerald-200'
  },

  info: {
    solid:
      'bg-sky-700 text-white hover:bg-sky-500 focus:ring-sky-400 focus:ring-2 focus:bg-sky-700 active:bg-sky-700',
    outline:
      'border border-sky-600 text-sky-700 hover:bg-sky-100 focus:ring-sky-50 focus:ring-2 focus:bg-sky-200 active:bg-sky-200',
    ghost:
      'text-sky-700 hover:bg-sky-100 focus:ring-sky-50 focus:ring-2 focus:bg-sky-200 active:bg-sky-200'
  }
} satisfies Record<ButtonIntent, Record<ButtonVariant, string>>;
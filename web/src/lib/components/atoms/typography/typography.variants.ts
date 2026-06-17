import type {
  HeadingSize,
  HeadingTone,
  HeadingWeight,
  TextSize,
  TextTone,
  TextWeight
} from './typography.types';

export const headingBaseClasses = [
  'font-heading',
  'tracking-tight',
  'leading-tight'
].join(' ');

export const headingSizeClasses = {
  sm: 'text-lg',
  md: 'text-xl',
  lg: 'text-2xl',
  xl: 'text-3xl',
  '2xl': 'text-4xl'
} satisfies Record<HeadingSize, string>;

export const headingToneClasses = {
  default: 'text-slate-800',
  muted: 'text-slate-700',
  subtle: 'text-slate-500'
} satisfies Record<HeadingTone, string>;

export const headingWeightClasses = {
  medium: 'font-medium',
  semibold: 'font-semibold',
  bold: 'font-bold'
} satisfies Record<HeadingWeight, string>;

export const headingUppercaseClasses = [
  'uppercase',
  'tracking-wider'
].join(' ');

export const textBaseClasses = [
  'leading-relaxed'
].join(' ');

export const textSizeClasses = {
  xs: 'text-xs',
  sm: 'text-sm',
  md: 'text-base',
  lg: 'text-lg'
} satisfies Record<TextSize, string>;

export const textToneClasses = {
  default: 'text-slate-700',
  muted: 'text-slate-500',
  subtle: 'text-slate-500',
  strong: 'text-slate-950',
  danger: 'text-red-700',
  success: 'text-emerald-700'
} satisfies Record<TextTone, string>;

export const textWeightClasses = {
  normal: 'font-normal',
  medium: 'font-medium',
  semibold: 'font-semibold',
  bold: 'font-bold'
} satisfies Record<TextWeight, string>;

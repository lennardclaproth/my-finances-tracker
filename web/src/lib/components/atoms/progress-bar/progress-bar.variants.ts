import type { ProgressBarIntent, ProgressBarSize } from './progress-bar.types';

export const progressBarSizeClasses = {
  sm: 'h-1.5',
  md: 'h-2.5',
  lg: 'h-4'
} satisfies Record<ProgressBarSize, string>;

export const progressBarFillClasses = {
  primary: 'bg-slate-600',
  secondary: 'bg-emerald-400',
  success: 'bg-emerald-600',
  warning: 'bg-amber-500',
  error: 'bg-red-600'
} satisfies Record<ProgressBarIntent, string>;

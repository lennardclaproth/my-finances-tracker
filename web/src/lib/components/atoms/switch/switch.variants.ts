import type { SwitchIntent, SwitchSize } from './switch.types';

export const switchTrackSizeClasses = {
  sm: 'h-4 w-7',
  md: 'h-5 w-9',
  lg: 'h-6 w-11'
} satisfies Record<SwitchSize, string>;

export const switchThumbSizeClasses = {
  sm: 'size-3',
  md: 'size-4',
  lg: 'size-5'
} satisfies Record<SwitchSize, string>;

// Thumb travel distance when checked, per size.
export const switchThumbTranslateClasses = {
  sm: 'translate-x-3',
  md: 'translate-x-4',
  lg: 'translate-x-5'
} satisfies Record<SwitchSize, string>;

// Track colour when the switch is on, keyed by intent.
export const switchOnTrackClasses = {
  primary: 'bg-slate-600',
  secondary: 'bg-emerald-400',
  success: 'bg-emerald-600',
  warning: 'bg-amber-500',
  error: 'bg-red-600'
} satisfies Record<SwitchIntent, string>;

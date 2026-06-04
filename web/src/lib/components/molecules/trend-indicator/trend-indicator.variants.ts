import type { TrendIndicatorSize } from './trend-indicator.types';

export const trendTextSizeClasses = {
  sm: 'text-xs',
  md: 'text-sm',
  lg: 'text-base'
} satisfies Record<TrendIndicatorSize, string>;

// Maps to the Icon atom's size prop.
export const trendIconSizeClasses = {
  sm: 'sm',
  md: 'sm',
  lg: 'md'
} as const;

// Colour by direction of change.
export const trendDirectionClasses = {
  up: 'text-emerald-700',
  down: 'text-red-700',
  flat: 'text-slate-500'
} satisfies Record<'up' | 'down' | 'flat', string>;

export const trendDirectionIcons = {
  up: 'heroicons:arrow-up',
  down: 'heroicons:arrow-down',
  flat: 'heroicons:minus'
} satisfies Record<'up' | 'down' | 'flat', string>;

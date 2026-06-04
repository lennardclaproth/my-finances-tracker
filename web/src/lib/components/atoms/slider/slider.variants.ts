import type { SliderIntent } from './slider.types';

// Native range styled via the CSS accent-color, keyed by intent.
export const sliderAccentClasses = {
  primary: 'accent-slate-600',
  secondary: 'accent-emerald-500',
  success: 'accent-emerald-600',
  warning: 'accent-amber-500',
  error: 'accent-red-600'
} satisfies Record<SliderIntent, string>;

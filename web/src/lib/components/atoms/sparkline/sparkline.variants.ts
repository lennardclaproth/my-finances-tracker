import type { ResolvedSparklineTone } from './sparkline.types';

export const sparklineToneClasses = {
  positive: 'text-emerald-600',
  negative: 'text-red-600',
  neutral: 'text-slate-500'
} satisfies Record<ResolvedSparklineTone, string>;

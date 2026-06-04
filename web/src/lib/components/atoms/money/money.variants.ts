import type { MoneySize, MoneyWeight } from './money.types';

export const moneyBaseClasses = [
  'tabular-nums',
  'whitespace-nowrap'
].join(' ');

export const moneySizeClasses = {
  sm: 'text-sm',
  md: 'text-base',
  lg: 'text-lg',
  xl: 'text-2xl'
} satisfies Record<MoneySize, string>;

export const moneyWeightClasses = {
  normal: 'font-normal',
  medium: 'font-medium',
  semibold: 'font-semibold',
  bold: 'font-bold'
} satisfies Record<MoneyWeight, string>;

// Applied only when `colored` is set, keyed by the sign of the amount.
export const moneySignClasses = {
  positive: 'text-emerald-700',
  negative: 'text-red-700',
  zero: 'text-slate-500'
} satisfies Record<'positive' | 'negative' | 'zero', string>;

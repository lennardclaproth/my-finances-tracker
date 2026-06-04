import type { CurrencyInputSize } from './currency-input.types';

// Reserve room on the left for the currency symbol, per size.
export const currencySymbolPaddingClasses = {
  sm: 'pl-7',
  md: 'pl-8',
  lg: 'pl-9'
} satisfies Record<CurrencyInputSize, string>;

export const currencySymbolContainerClasses = {
  sm: 'h-8 w-7 text-sm',
  md: 'h-10 w-8 text-sm',
  lg: 'h-12 w-9 text-base'
} satisfies Record<CurrencyInputSize, string>;

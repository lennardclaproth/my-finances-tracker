import type {
  InputIntent,
  InputShape,
  InputSize
} from '$lib/components/atoms/input/input.types';

// CurrencyInput shares the Input field vocabulary (data reuse, not component composition).
export {
  inputSizes as currencyInputSizes,
  inputIntents as currencyInputIntents,
  inputShapes as currencyInputShapes
} from '$lib/components/atoms/input/input.types';

export type CurrencyInputSize = InputSize;
export type CurrencyInputIntent = InputIntent;
export type CurrencyInputShape = InputShape;

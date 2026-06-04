import type {
  InputIntent,
  InputShape,
  InputSize
} from '$lib/components/atoms/input/input.types';

// Select shares the Input field vocabulary so the two controls stay visually consistent.
// Re-exporting the option arrays is data reuse, not component composition.
export {
  inputSizes as selectSizes,
  inputIntents as selectIntents,
  inputShapes as selectShapes
} from '$lib/components/atoms/input/input.types';

export type SelectSize = InputSize;
export type SelectIntent = InputIntent;
export type SelectShape = InputShape;

export type SelectOption = {
  value: string;
  label: string;
  disabled?: boolean;
};

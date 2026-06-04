import type {
  InputIntent,
  InputShape,
  InputSize
} from '$lib/components/atoms/input/input.types';

// Textarea shares the Input field vocabulary (data reuse, not component composition).
export {
  inputSizes as textareaSizes,
  inputIntents as textareaIntents,
  inputShapes as textareaShapes
} from '$lib/components/atoms/input/input.types';

export const textareaResizes = ['none', 'vertical', 'horizontal', 'both'] as const;

export type TextareaSize = InputSize;
export type TextareaIntent = InputIntent;
export type TextareaShape = InputShape;
export type TextareaResize = (typeof textareaResizes)[number];

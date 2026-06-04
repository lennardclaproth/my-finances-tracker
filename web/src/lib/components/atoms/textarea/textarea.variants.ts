import type { TextareaResize, TextareaSize } from './textarea.types';

// Textarea height comes from `rows`, so size only controls padding and text scale.
export const textareaSizeClasses = {
  sm: 'px-2 py-1.5 text-sm',
  md: 'px-3 py-2 text-sm',
  lg: 'px-4 py-2.5 text-base'
} satisfies Record<TextareaSize, string>;

export const textareaResizeClasses = {
  none: 'resize-none',
  vertical: 'resize-y',
  horizontal: 'resize-x',
  both: 'resize'
} satisfies Record<TextareaResize, string>;

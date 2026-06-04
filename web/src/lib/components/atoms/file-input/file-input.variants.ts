import type { FileInputSize } from './file-input.types';

// Sizing applies to both the text and the native file-selector button (via `file:` utilities).
export const fileInputSizeClasses = {
  sm: 'text-xs file:px-2.5 file:py-1.5 file:text-xs',
  md: 'text-sm file:px-3 file:py-2 file:text-sm',
  lg: 'text-base file:px-4 file:py-2.5 file:text-base'
} satisfies Record<FileInputSize, string>;

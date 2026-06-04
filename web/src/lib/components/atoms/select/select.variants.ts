import type { SelectSize } from './select.types';

// Reserve room on the right for the chevron, per size.
export const selectChevronPaddingClasses = {
  sm: 'pr-8',
  md: 'pr-9',
  lg: 'pr-10'
} satisfies Record<SelectSize, string>;

export const selectChevronContainerClasses = {
  sm: 'h-8 w-8',
  md: 'h-10 w-9',
  lg: 'h-12 w-10'
} satisfies Record<SelectSize, string>;

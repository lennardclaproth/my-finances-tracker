export const checkboxSizes = ['sm', 'md', 'lg'] as const;

export type CheckboxSize = (typeof checkboxSizes)[number];
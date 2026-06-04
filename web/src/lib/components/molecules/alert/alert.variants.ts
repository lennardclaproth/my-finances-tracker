import type { AlertIntent } from './alert.types';

export const alertContainerClasses = {
  info: 'border-sky-200 bg-sky-50 text-sky-900',
  success: 'border-emerald-200 bg-emerald-50 text-emerald-900',
  warning: 'border-amber-200 bg-amber-50 text-amber-900',
  error: 'border-red-200 bg-red-50 text-red-900'
} satisfies Record<AlertIntent, string>;

export const alertIconColorClasses = {
  info: 'text-sky-600',
  success: 'text-emerald-600',
  warning: 'text-amber-600',
  error: 'text-red-600'
} satisfies Record<AlertIntent, string>;

export const alertIcons = {
  info: 'heroicons:information-circle',
  success: 'heroicons:check-circle',
  warning: 'heroicons:exclamation-triangle',
  error: 'heroicons:x-circle'
} satisfies Record<AlertIntent, string>;

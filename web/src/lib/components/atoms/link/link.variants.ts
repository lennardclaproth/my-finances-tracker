import type { LinkSize, LinkTone, LinkUnderline } from './link.types';

export const linkBaseClasses = [
  'inline-flex items-center gap-1 rounded-sm',
  'transition-colors duration-150 ease-out',
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2'
].join(' ');

export const linkToneClasses = {
  default: 'text-slate-700 hover:text-slate-900 focus-visible:ring-slate-300',
  primary: 'text-sky-700 hover:text-sky-800 focus-visible:ring-sky-200',
  muted: 'text-slate-500 hover:text-slate-700 focus-visible:ring-slate-200'
} satisfies Record<LinkTone, string>;

export const linkSizeClasses = {
  sm: 'text-sm',
  md: 'text-base',
  lg: 'text-lg'
} satisfies Record<LinkSize, string>;

export const linkUnderlineClasses = {
  always: 'underline underline-offset-2',
  hover: 'no-underline underline-offset-2 hover:underline',
  none: 'no-underline'
} satisfies Record<LinkUnderline, string>;

import type {
  PanelPadding,
  PanelShadow,
  PanelShape,
  PanelVariant
} from './panel.types';

export const basePanelClasses = [
  'relative',
  'transition-colors duration-150 ease-out'
].join(' ');

export const panelVariantClasses = {
  default: 'bg-white text-slate-800',
  muted: 'bg-taupe-50 text-slate-800',
  floating: 'bg-white text-slate-800',
  ghost: 'bg-transparent text-slate-800'
} satisfies Record<PanelVariant, string>;

export const panelBorderClasses = {
  true: 'border border-slate-300',
  false: 'border border-transparent'
} satisfies Record<'true' | 'false', string>;

export const panelPaddingClasses = {
  none: 'p-0',
  sm: 'p-3',
  md: 'p-4',
  lg: 'p-6'
} satisfies Record<PanelPadding, string>;

export const panelShapeClasses = {
  sm: 'rounded-md',
  md: 'rounded-xl',
  xl: 'rounded-2xl'
} satisfies Record<PanelShape, string>;

export const panelShadowClasses = {
  none: 'shadow-none',
  sm: 'shadow-sm',
  md: 'shadow-md'
} satisfies Record<PanelShadow, string>;

export const panelInteractiveClasses = [
  'hover:bg-slate-50',
  'focus-within:ring-2 focus-within:ring-slate-200',
  'focus-within:ring-offset-2 focus-within:ring-offset-slate-50'
].join(' ');
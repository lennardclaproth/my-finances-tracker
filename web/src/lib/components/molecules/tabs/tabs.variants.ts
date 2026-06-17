import type { TabsSize } from './tabs.types';

export const tabsTrackClasses = [
	'inline-flex items-center gap-1 rounded-xl border border-slate-200 bg-taupe-100 p-1'
].join(' ');

export const tabBaseClasses = [
	'inline-flex items-center justify-center gap-1.5 rounded-lg font-medium',
	'transition-all duration-150 ease-out',
	'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-300',
	'disabled:pointer-events-none disabled:opacity-50'
].join(' ');

export const tabSizeClasses = {
	sm: 'h-8 px-2.5 text-xs',
	md: 'h-9 px-3 text-sm',
	lg: 'h-10 px-4 text-sm'
} satisfies Record<TabsSize, string>;

export const tabStateClasses = {
	selected: 'bg-white text-slate-900 shadow-sm',
	unselected: 'text-slate-600 hover:text-slate-900'
} satisfies Record<'selected' | 'unselected', string>;

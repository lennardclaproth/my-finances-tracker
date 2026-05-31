import type {
	BadgeIntent,
	BadgeShape,
	BadgeSize,
	BadgeVariant
} from './badge.types';

export const baseBadgeClasses = [
	'inline-flex items-center justify-center gap-1.5',
	'border font-medium whitespace-nowrap',
	'transition-colors duration-150 ease-out'
].join(' ');

export const badgeSizeClasses = {
	sm: 'px-2 py-0.5 text-xs',
	md: 'px-2.5 py-1 text-xs',
	lg: 'px-3 py-1.5 text-sm'
} satisfies Record<BadgeSize, string>;

export const badgeDotSizeClasses = {
	sm: 'size-1.5',
	md: 'size-1.5',
	lg: 'size-2'
} satisfies Record<BadgeSize, string>;

export const badgeShapeClasses = {
	rounded: 'rounded-md',
	pill: 'rounded-full'
} satisfies Record<BadgeShape, string>;

export const badgeIntentVariantClasses = {
	neutral: {
		soft:
			'border-slate-200 bg-slate-100 text-slate-700',
		solid:
			'border-slate-600 bg-slate-600 text-white',
		outline:
			'border-slate-400 bg-transparent text-slate-700'
	},

	primary: {
		soft:
			'border-amber-200 bg-slate-100 text-slate-700',
		solid:
			'border-amber-200 bg-slate-600 text-amber-200',
		outline:
			'border-slate-700 bg-transparent text-slate-700'
	},

    secondary: {
        soft:
            'border-emerald-200 bg-emerald-100 text-slate-800',
        solid:
            'border-slate-800 bg-emerald-400 text-slate-800',
        outline:
            'border-slate-800 bg-transparent text-slate-800'
    },

	success: {
		soft:
			'border-emerald-200 bg-emerald-100 text-emerald-800',
		solid:
			'border-emerald-700 bg-emerald-700 text-white',
		outline:
			'border-emerald-700 bg-transparent text-emerald-800'
	},

	warning: {
		soft:
			'border-amber-200 bg-amber-100 text-amber-800',
		solid:
			'border-amber-500 bg-amber-500 text-slate-800',
		outline:
			'border-amber-600 bg-transparent text-amber-800'
	},

	error: {
		soft:
			'border-red-200 bg-red-100 text-red-700',
		solid:
			'border-red-700 bg-red-700 text-white',
		outline:
			'border-red-600 bg-transparent text-red-700'
	},

	info: {
		soft:
			'border-sky-200 bg-sky-100 text-sky-700',
		solid:
			'border-sky-700 bg-sky-700 text-white',
		outline:
			'border-sky-600 bg-transparent text-sky-700'
	}
} satisfies Record<BadgeIntent, Record<BadgeVariant, string>>;

export const badgeDotIntentClasses = {
	neutral: 'bg-slate-500',
	primary: 'bg-amber-200',
    secondary: 'bg-emerald-400',
	success: 'bg-emerald-700',
	warning: 'bg-amber-500',
	error: 'bg-red-700',
	info: 'bg-sky-700'
} satisfies Record<BadgeIntent, string>;
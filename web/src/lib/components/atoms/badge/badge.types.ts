export const badgeIntents = [
	'neutral',
	'primary',
    'secondary',
	'success',
	'warning',
	'error',
	'info'
] as const;

export type BadgeIntent = (typeof badgeIntents)[number];

export const badgeVariants = ['soft', 'solid', 'outline'] as const;

export type BadgeVariant = (typeof badgeVariants)[number];

export const badgeSizes = ['sm', 'md', 'lg'] as const;

export type BadgeSize = (typeof badgeSizes)[number];

export const badgeShapes = ['rounded', 'pill'] as const;

export type BadgeShape = (typeof badgeShapes)[number];
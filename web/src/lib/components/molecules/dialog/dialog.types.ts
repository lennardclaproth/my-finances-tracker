export const dialogSizes = ['sm', 'md', 'lg', 'xl'] as const;

export type DialogSize = (typeof dialogSizes)[number];

export const dialogSizeClasses = {
	sm: 'max-w-sm',
	md: 'max-w-md',
	lg: 'max-w-lg',
	xl: 'max-w-2xl'
} satisfies Record<DialogSize, string>;

export const avatarSizes = ['xs', 'sm', 'md', 'lg', 'xl'] as const;

export const avatarShapes = ['circle', 'rounded', 'square'] as const;

export type AvatarSize = (typeof avatarSizes)[number];
export type AvatarShape = (typeof avatarShapes)[number];

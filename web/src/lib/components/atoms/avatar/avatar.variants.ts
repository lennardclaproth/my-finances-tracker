import type { AvatarShape, AvatarSize } from './avatar.types';

export const avatarSizeClasses = {
  xs: 'size-6 text-xs',
  sm: 'size-8 text-xs',
  md: 'size-10 text-sm',
  lg: 'size-12 text-base',
  xl: 'size-16 text-lg'
} satisfies Record<AvatarSize, string>;

export const avatarShapeClasses = {
  circle: 'rounded-full',
  rounded: 'rounded-lg',
  square: 'rounded-none'
} satisfies Record<AvatarShape, string>;

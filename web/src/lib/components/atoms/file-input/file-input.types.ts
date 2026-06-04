export const fileInputSizes = ['sm', 'md', 'lg'] as const;

export type FileInputSize = (typeof fileInputSizes)[number];

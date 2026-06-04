export const trendIndicatorSizes = ['sm', 'md', 'lg'] as const;

export const trendIndicatorFormats = ['percent', 'currency', 'number'] as const;

export type TrendIndicatorSize = (typeof trendIndicatorSizes)[number];
export type TrendIndicatorFormat = (typeof trendIndicatorFormats)[number];

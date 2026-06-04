export const sparklineTones = ['auto', 'positive', 'negative', 'neutral'] as const;

export type SparklineTone = (typeof sparklineTones)[number];

// Tone after 'auto' has been resolved against the data trend.
export type ResolvedSparklineTone = Exclude<SparklineTone, 'auto'>;

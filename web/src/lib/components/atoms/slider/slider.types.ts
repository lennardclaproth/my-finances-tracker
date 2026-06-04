export const sliderIntents = ['primary', 'secondary', 'success', 'warning', 'error'] as const;

export type SliderIntent = (typeof sliderIntents)[number];

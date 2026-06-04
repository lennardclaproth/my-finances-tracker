export const alertIntents = ['info', 'success', 'warning', 'error'] as const;

export type AlertIntent = (typeof alertIntents)[number];

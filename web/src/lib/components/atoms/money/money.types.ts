export const moneySizes = ['sm', 'md', 'lg', 'xl'] as const;

export const moneyWeights = ['normal', 'medium', 'semibold', 'bold'] as const;

export const moneySignDisplays = ['auto', 'always', 'never', 'exceptZero'] as const;

export type MoneySize = (typeof moneySizes)[number];
export type MoneyWeight = (typeof moneyWeights)[number];
export type MoneySignDisplay = (typeof moneySignDisplays)[number];

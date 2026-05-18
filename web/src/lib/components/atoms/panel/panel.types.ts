export const panelVariants = ['default', 'muted', 'floating', 'ghost'] as const;

export const panelPaddings = ['none', 'sm', 'md', 'lg'] as const;

export const panelShapes = ['sm', 'md', 'xl'] as const;

export const panelShadows = ['none', 'sm', 'md'] as const;

export type PanelVariant = typeof panelVariants[number];
export type PanelPadding = typeof panelPaddings[number];
export type PanelShape = typeof panelShapes[number];
export type PanelShadow = typeof panelShadows[number];
<script lang="ts">
  import type { Snippet } from 'svelte';

  import type {
    PanelPadding,
    PanelShadow,
    PanelShape,
    PanelVariant
  } from './panel.types';

  import {
    basePanelClasses,
    panelBorderClasses,
    panelInteractiveClasses,
    panelPaddingClasses,
    panelShadowClasses,
    panelShapeClasses,
    panelVariantClasses
  } from './panel.variants';

  type Props = {
    variant?: PanelVariant;
    padding?: PanelPadding;
    shape?: PanelShape;
    shadow?: PanelShadow;
    bordered?: boolean;
    interactive?: boolean;
    class?: string;
    children?: Snippet;
  };

  let {
    variant = 'default',
    padding = 'md',
    shape = 'md',
    shadow = 'none',
    bordered = true,
    interactive = false,
    class: className = '',
    children
  }: Props = $props();

  const classes = $derived([
    basePanelClasses,
    panelVariantClasses[variant],
    panelPaddingClasses[padding],
    panelShapeClasses[shape],
    panelShadowClasses[shadow],
    panelBorderClasses[String(bordered) as 'true' | 'false'],
    interactive ? panelInteractiveClasses : '',
    className
  ]
    .filter(Boolean)
    .join(' '));
</script>

<div class={classes}>
  {@render children?.()}
</div>
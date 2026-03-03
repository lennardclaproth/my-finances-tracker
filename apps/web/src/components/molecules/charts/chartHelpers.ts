import type { Chart, Plugin } from "chart.js";

export const chartTheme = {
  text: "#3f3f46",
  muted: "#71717a",
  tooltipBg: "#0f172a",
  grid: "rgba(148, 163, 184, 0.2)",
  gridZero: "rgba(51, 65, 85, 0.55)",
  hoverGuide: "rgba(100, 116, 139, 0.45)",
};

export const HOVER_MARKER_SIZE = 8;
export const HOVER_MARKER_BORDER_WIDTH = 2;
export const HOVER_MARKER_BORDER_COLOR = "#6366f1";
export const HOVER_MARKER_FILL_COLOR = "#ffffff";

export function horizontalGridConfig() {
  return {
    display: false,
  };
}

export const horizontalGuidePlugin: Plugin = {
  id: "horizontal-guide-lines",
  beforeDatasetsDraw(chart) {
    const yScale = chart.scales.y;
    if (!yScale || !chart.chartArea) {
      return;
    }

    const { left, right } = chart.chartArea;
    const { ctx } = chart;
    for (const tick of yScale.ticks) {
      const value = Number(tick.value);
      if (Number.isNaN(value)) {
        continue;
      }
      const y = yScale.getPixelForValue(value);
      const isZero = value === 0;
      ctx.save();
      ctx.beginPath();
      ctx.strokeStyle = isZero ? chartTheme.gridZero : chartTheme.grid;
      ctx.lineWidth = isZero ? 1.75 : 1;
      ctx.setLineDash(isZero ? [] : [4, 4]);
      ctx.moveTo(left, y);
      ctx.lineTo(right, y);
      ctx.stroke();
      ctx.restore();
    }
  },
};

export function hexToRgba(hex: string, alpha: number): string {
  const normalized = hex.replace("#", "");
  const bigint = Number.parseInt(normalized, 16);
  const r = (bigint >> 16) & 255;
  const g = (bigint >> 8) & 255;
  const b = bigint & 255;
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

export function gradientFillWithAlpha(
  chartInstance: Chart,
  color: string,
  alphaTop: number,
  alphaBottom: number,
): CanvasGradient | string {
  const chartArea = chartInstance.chartArea;
  if (!chartArea) {
    return hexToRgba(color, alphaTop);
  }
  const gradient = chartInstance.ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
  gradient.addColorStop(0, hexToRgba(color, alphaTop));
  gradient.addColorStop(1, hexToRgba(color, alphaBottom));
  return gradient;
}

export const hoverGuidePlugin: Plugin = {
  id: "hover-guide-line",
  afterDatasetsDraw(chart) {
    if ((chart as Chart & { __suspendHoverGuide?: boolean }).__suspendHoverGuide) {
      return;
    }
    const tooltip = chart.tooltip;
    if (!tooltip) {
      return;
    }
    const activeElements = tooltip.getActiveElements();
    if (!activeElements || activeElements.length === 0) {
      return;
    }
    const x = activeElements[0].element.x;
    const { left, right, top, bottom } = chart.chartArea;
    if (x < left || x > right) {
      return;
    }

    const { ctx } = chart;
    ctx.save();
    ctx.beginPath();
    ctx.setLineDash([4, 4]);
    ctx.lineWidth = 1;
    ctx.strokeStyle = chartTheme.hoverGuide;
    ctx.moveTo(x, top);
    ctx.lineTo(x, bottom);
    ctx.stroke();
    ctx.restore();
  },
};

export function toCanvasX(chart: Chart, canvas: HTMLCanvasElement, clientX: number): number {
  const rect = canvas.getBoundingClientRect();
  const xFromCanvas = clientX - rect.left;
  return Math.max(chart.chartArea.left, Math.min(chart.chartArea.right, xFromCanvas));
}

export interface DragIndicatorStyles {
  marker: Record<string, string>;
  band: Record<string, string>;
}

export function selectionIndicatorStyles(
  anchorX: number,
  targetX: number,
  anchorY: number,
  bandTop: number,
  bandHeight: number,
): DragIndicatorStyles {
  const diameter = HOVER_MARKER_SIZE;
  const half = diameter / 2;
  const left = Math.min(anchorX, targetX);
  const width = Math.max(0, Math.abs(targetX - anchorX));

  return {
    marker: {
      left: `${anchorX - half}px`,
      top: `${anchorY - half}px`,
      width: `${diameter}px`,
      height: `${diameter}px`,
    },
    band: {
      left: `${left}px`,
      top: `${bandTop}px`,
      width: `${width}px`,
      height: `${bandHeight}px`,
    },
  };
}

export function selectionLineStyle(x: number, top: number, height: number): Record<string, string> {
  return {
    left: `${x}px`,
    top: `${top}px`,
    height: `${height}px`,
  };
}

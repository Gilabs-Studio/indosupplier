/**
 * Color utility functions for converting CSS template variables to hex/rgb values.
 * Used in area form and other components that need to send colors to the API.
 */

/**
 * Mapping of CSS template variable names to their hex color values.
 * These correspond to the --color-* definitions in globals.css
 */
const COLOR_TEMPLATE_MAP: Record<string, string> = {
  "var(--color-primary)": "#E27D18",           // 30 81% 49% - Orange Primary
  "var(--color-secondary)": "#E5E7EB",         // 220 14% 92%
  "var(--color-destructive)": "#EF4444",       // 0 84.2% 60.2% - Red
  "var(--color-accent)": "#E5E7EB",            // 220 14% 92%
  "var(--color-success)": "#22C55E",           // 142 76% 36% - Green
  "var(--color-warning)": "#F59E0B",           // 38 92% 50% - Amber
  "var(--color-muted)": "#D1D5DB",             // 220 14% 92%
  "var(--color-info)": "#0EA5E9",              // 187 100% 44% - Cyan (approximation)
  "var(--color-purple)": "#A855F7",            // 262 83% 58% - Purple (light theme)
  "var(--color-cyan)": "#06B6D4",              // 187 100% 44% - Cyan
  "var(--color-rose)": "#F43F5E",              // 340 82% 57% - Rose
};

/**
 * Converts a CSS color (template variable or hex/rgb) to a normalized hex value.
 * If the color is a template variable, converts it to hex.
 * If already hex/rgb, returns as-is (trimmed).
 *
 * @param color - The CSS color value (e.g., "var(--color-primary)" or "#FF0000")
 * @returns Normalized color value (hex string)
 */
export function normalizeColorForAPI(color: string | null | undefined): string {
  if (!color) return "#E27D18"; // Default to primary orange

  const trimmed = color.trim();

  // Check if it's a template variable
  if (trimmed in COLOR_TEMPLATE_MAP) {
    return COLOR_TEMPLATE_MAP[trimmed];
  }

  // If it's already a valid color format (hex, rgb, etc.), return as-is
  if (isValidCSSColor(trimmed)) {
    return trimmed;
  }

  // Fallback to primary orange
  return "#E27D18";
}

/**
 * Validates if a string is a valid CSS color format.
 */
function isValidCSSColor(color: string): boolean {
  // Hex color: #RGB, #RRGGBB, #RGBA, #RRGGBBAA
  if (/^#([0-9A-Fa-f]{3}){1,2}([0-9A-Fa-f]{2})?$/.test(color)) {
    return true;
  }

  // rgb/rgba: rgb(255, 0, 0) or rgba(255, 0, 0, 0.5)
  if (/^rgba?\s*\(\s*\d+\s*,\s*\d+\s*,\s*\d+(\s*,\s*[\d.]+)?\s*\)$/.test(color)) {
    return true;
  }

  // hsl/hsla: hsl(0, 100%, 50%) or hsla(0, 100%, 50%, 0.5)
  if (/^hsla?\s*\(\s*\d+\s*,\s*\d+%\s*,\s*\d+%(\s*,\s*[\d.]+)?\s*\)$/.test(color)) {
    return true;
  }

  // CSS named colors (basic check)
  const namedColors = [
    "red", "green", "blue", "black", "white", "gray", "yellow", "orange",
    "purple", "pink", "cyan", "brown", "lime", "navy", "teal", "olive",
  ];
  if (namedColors.includes(color.toLowerCase())) {
    return true;
  }

  return false;
}

/**
 * Get all available template colors for display.
 */
export function getTemplateColors(): Array<{ template: string; hex: string }> {
  return Object.entries(COLOR_TEMPLATE_MAP).map(([template, hex]) => ({
    template,
    hex,
  }));
}

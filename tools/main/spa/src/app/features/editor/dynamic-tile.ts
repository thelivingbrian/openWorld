/**
 * Dynamic tile token parser for the design workspace editor.
 *
 * Mirrors the runtime token logic in server/main/assets/canvas.js so that
 * animated prototypes and fragments preview their behavior while editing.
 *
 * Supported tokens:
 *   cycle(colorA,colorB)    – fill color oscillates between two named colors
 *   cycle-b(colorA,colorB)  – border color oscillates between two named colors
 *   rainbow                 – fill cycles through a full HSL hue sweep
 *   rainbow-b               – border cycles through a full HSL hue sweep
 *   water                   – fill shimmers between blue and sky-blue
 *   sparkle                 – brightness shimmer (approximates the canvas glint dots)
 *
 * Per-cell phase offsets are derived deterministically from (y, x) coordinates
 * so the animation looks stable while panning or editing the blueprint.
 *
 * Editor vs runtime differences:
 *   - sparkle uses a CSS `filter: brightness()` shimmer rather than drawing
 *     individual glint circles (canvas-only).
 *   - All other tokens produce identical colour values to the runtime.
 *   - Unknown / invalid tokens are ignored and produce an empty style object,
 *     so old or custom class strings remain safe and non-breaking.
 */

/** Named color map – must stay in sync with COLOR_MAP in canvas.js. */
export const DYNAMIC_COLOR_MAP: Record<string, string> = {
  invisible: 'rgba(0, 0, 0, 0)',
  day: 'rgb(170, 255, 255)',
  night: 'rgb(18, 5, 0)',
  twilight: 'rgb(255, 106, 92)',
  green: 'rgb(32, 255, 60)',
  'dark-green': 'rgb(32, 155, 60)',
  lime: 'rgb(150, 255, 60)',
  yellow: 'rgb(255, 255, 60)',
  orange: 'rgb(255, 150, 60)',
  red: 'rgb(255, 32, 60)',
  'dark-red': 'rgb(139, 0, 0)',
  blue: 'rgb(72, 52, 238)',
  fuchsia: 'rgb(253, 52, 172)',
  'dim-fuchsia': 'rgb(89, 18, 60)',
  pink: 'rgb(253, 182, 215)',
  salmon: 'rgb(255, 165, 145)',
  'half-gray': 'rgb(120, 120, 129)',
  'dark-blue': 'rgb(32, 15, 65)',
  sand: 'rgb(250, 245, 175)',
  grass: 'rgb(210, 250, 180)',
  ice: 'rgb(225, 240, 255)',
  wall: 'rgb(100, 100, 100)',
  white: 'rgb(255, 255, 255)',
  'light-gray': 'rgb(210, 210, 210)',
  cinnamon: 'rgb(210, 125, 45)',
  nude: 'rgb(242, 210, 189)',
  chocolate: 'rgb(123, 63, 0)',
  tan: 'rgb(210, 180, 140)',
  copper: 'rgb(184, 115, 51)',
  'dark-gray': 'rgb(60, 60, 60)',
  purple: 'rgb(128, 0, 128)',
  teal: 'rgb(0, 128, 128)',
  turquoise: 'rgb(64, 224, 208)',
  lavender: 'rgb(218, 203, 250)',
  olive: 'rgb(128, 128, 0)',
  gold: 'rgb(255, 215, 0)',
  navy: 'rgb(0, 0, 128)',
  peach: 'rgb(255, 218, 185)',
  burgundy: 'rgb(128, 0, 32)',
  'dark-grass': 'rgb(160, 200, 130)',
  'sky-blue': 'rgb(48, 152, 240)',
  'dim-sky-blue': 'rgb(17, 53, 84)',
  black: 'rgb(0, 0, 0)',
  'med-gray': 'rgb(145, 145, 145)',
  'dark-lavender': 'rgb(172, 152, 219)',
};

/** Returns true when the class string contains at least one dynamic token. */
export function hasDynamicToken(classes: string | undefined): boolean {
  if (!classes) return false;
  return (
    classes.includes('sparkle') ||
    classes.includes('rainbow') ||
    classes.includes('water') ||
    classes.includes('cycle(') ||
    classes.includes('cycle-b(')
  );
}

/**
 * Removes dynamic tokens from a class string so the remaining tokens are
 * valid CSS class names that can be applied directly to a DOM element.
 * This prevents browser warnings about unknown class names like "cycle(blue,red)".
 */
export function stripDynamicTokens(classes: string | undefined): string {
  if (!classes) return '';
  return classes
    .split(/\s+/)
    .filter((token) => !isSingleDynamicToken(token))
    .join(' ')
    .trim();
}

/**
 * Computes an Angular-compatible inline style object for a class string at the
 * given time and world coordinates.
 *
 * Returns an empty object when there are no dynamic tokens – callers can apply
 * the result unconditionally via `[ngStyle]` without extra checks.
 */
export function computeDynamicStyle(
  classes: string | undefined,
  timeMs: number,
  y: number,
  x: number,
): Record<string, string> {
  if (!hasDynamicToken(classes)) {
    return {};
  }

  const tokens = (classes ?? '').split(/\s+/);
  const style: Record<string, string> = {};

  for (const token of tokens) {
    const cycleMatch = token.match(/^cycle\(([^,]+),([^)]+)\)$/);
    if (cycleMatch) {
      const colorA = DYNAMIC_COLOR_MAP[cycleMatch[1].trim()];
      const colorB = DYNAMIC_COLOR_MAP[cycleMatch[2].trim()];
      if (colorA && colorB) {
        const mix = 0.5 + 0.5 * Math.sin(timeMs / 420 + (y + x) * 0.3);
        style['backgroundColor'] = mixColor(colorA, colorB, mix);
      }
      continue;
    }

    const cycleBorderMatch = token.match(/^cycle-b\(([^,]+),([^)]+)\)$/);
    if (cycleBorderMatch) {
      const colorA = DYNAMIC_COLOR_MAP[cycleBorderMatch[1].trim()];
      const colorB = DYNAMIC_COLOR_MAP[cycleBorderMatch[2].trim()];
      if (colorA && colorB) {
        const mix = 0.5 + 0.5 * Math.sin(timeMs / 420 + (y + x) * 0.3);
        style['borderColor'] = mixColor(colorA, colorB, mix);
        style['borderStyle'] = 'solid';
        if (!style['borderWidth']) {
          style['borderWidth'] = '1px';
        }
      }
      continue;
    }

    if (token === 'rainbow') {
      style['backgroundColor'] = rainbowColorAt(timeMs, (y + x) * 11);
      continue;
    }

    if (token === 'rainbow-b') {
      style['borderColor'] = rainbowColorAt(timeMs, (y + x) * 11);
      style['borderStyle'] = 'solid';
      if (!style['borderWidth']) {
        style['borderWidth'] = '1px';
      }
      continue;
    }

    if (token === 'water') {
      const wave = 0.5 + 0.5 * Math.sin(timeMs / 560 + x * 0.7 + y * 0.35);
      const blueColor = DYNAMIC_COLOR_MAP['blue'];
      const skyBlueColor = DYNAMIC_COLOR_MAP['sky-blue'];
      if (blueColor && skyBlueColor) {
        style['backgroundColor'] = mixColor(blueColor, skyBlueColor, wave);
      }
      continue;
    }

    if (token === 'sparkle') {
      // Canvas runtime draws glint dots; in the editor we approximate with a
      // brightness shimmer. Phase offset keeps adjacent cells out of sync.
      const brightness = 1.0 + 0.25 * Math.abs(Math.sin(timeMs / 300 + (y * 7 + x) * 0.5));
      style['filter'] = `brightness(${brightness.toFixed(3)})`;
      continue;
    }
  }

  return style;
}

// ── Internal helpers ────────────────────────────────────────────────────────

function isSingleDynamicToken(token: string): boolean {
  return (
    token === 'sparkle' ||
    token === 'rainbow' ||
    token === 'rainbow-b' ||
    token === 'water' ||
    token.startsWith('cycle(') ||
    token.startsWith('cycle-b(')
  );
}

function mixColor(colorA: string, colorB: string, amount: number): string {
  const a = parseRgbColor(colorA);
  const b = parseRgbColor(colorB);
  if (!a || !b) return colorA;

  const t = Math.max(0, Math.min(1, amount));
  const r = Math.round(a.r * (1 - t) + b.r * t);
  const g = Math.round(a.g * (1 - t) + b.g * t);
  const bCh = Math.round(a.b * (1 - t) + b.b * t);

  return `rgb(${r}, ${g}, ${bCh})`;
}

function parseRgbColor(color: string): { r: number; g: number; b: number } | null {
  const match = color.match(/^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/);
  if (!match) return null;
  return { r: Number(match[1]), g: Number(match[2]), b: Number(match[3]) };
}

function rainbowColorAt(timeMs: number, phase: number): string {
  const hue = ((timeMs / 14) + phase) % 360;
  return hslToRgbString(hue, 0.85, 0.6);
}

function hslToRgbString(h: number, s: number, l: number): string {
  const hue = ((h % 360) + 360) % 360;
  const c = (1 - Math.abs(2 * l - 1)) * s;
  const x = c * (1 - Math.abs(((hue / 60) % 2) - 1));
  const m = l - c / 2;

  let rPrime = 0;
  let gPrime = 0;
  let bPrime = 0;

  if (hue < 60) {
    rPrime = c;
    gPrime = x;
  } else if (hue < 120) {
    rPrime = x;
    gPrime = c;
  } else if (hue < 180) {
    gPrime = c;
    bPrime = x;
  } else if (hue < 240) {
    gPrime = x;
    bPrime = c;
  } else if (hue < 300) {
    rPrime = x;
    bPrime = c;
  } else {
    rPrime = c;
    bPrime = x;
  }

  const r = Math.round((rPrime + m) * 255);
  const g = Math.round((gPrime + m) * 255);
  const b = Math.round((bPrime + m) * 255);

  return `rgb(${r}, ${g}, ${b})`;
}

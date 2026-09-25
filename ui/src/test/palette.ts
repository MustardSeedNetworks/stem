/**
 * palette.ts — read the theme's colour tokens and measure them.
 *
 * The token values live in three CSS files; the tests that hold them to a
 * contrast or separation bar read those files directly rather than restating
 * the hex values, so a token edit is measured the moment it lands.
 */
import { readFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

export const src = resolve(dirname(fileURLToPath(import.meta.url)), '..');

export const SURFACES = ['base', 'raised', 'hover', 'sunken', 'deep'] as const;

export type Mode = 'light' | 'dark';
export type Rgb = [number, number, number];

function blocks(css: string, selector: string): string[] {
  const out: string[] = [];
  const re = new RegExp(`(^|\\n)${selector.replace('.', '\\.')}\\s*\\{`, 'g');
  for (const m of css.matchAll(re)) {
    let depth = 1;
    let i = m.index + m[0].length;
    const start = i;
    while (depth > 0 && i < css.length) {
      if (css[i] === '{') depth++;
      if (css[i] === '}') depth--;
      i++;
    }
    out.push(css.slice(start, i - 1));
  }
  return out;
}

/** Every `--color-*` hex token in effect in `mode`, keyed without the prefix. */
export function palette(mode: Mode): Map<string, string> {
  const css = ['theme/msn-shared.css', 'theme/product-stem.css', 'index.css'].map((file) =>
    readFileSync(join(src, file), 'utf8'),
  );
  // Every :root first, then every .dark, as the cascade applies them.
  const selectors = mode === 'light' ? [':root'] : [':root', '.dark'];
  const tokens = new Map<string, string>();
  for (const body of selectors.flatMap((selector) =>
    css.flatMap((file) => blocks(file, selector)),
  )) {
    for (const [, name = '', hex = ''] of body.matchAll(
      /--color-([\w-]+):\s*(#[0-9a-f]{6})\s*;/gi,
    )) {
      tokens.set(name, hex.toLowerCase());
    }
  }
  return tokens;
}

export function rgb(hex: string): Rgb {
  return [1, 3, 5].map((i) => Number.parseInt(hex.slice(i, i + 2), 16)) as Rgb;
}

/** sRGB channel (0-255) to linear light (0-1). */
export function linear(v: number): number {
  const s = v / 255;
  return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
}

function luminance([r, g, b]: Rgb): number {
  return 0.2126 * linear(r) + 0.7152 * linear(g) + 0.0722 * linear(b);
}

/** WCAG 2 contrast ratio. */
export function contrast(a: Rgb, b: Rgb): number {
  const la = luminance(a);
  const lb = luminance(b);
  return (Math.max(la, lb) + 0.05) / (Math.min(la, lb) + 0.05);
}

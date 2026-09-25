/**
 * moduleColors.test.ts — the six module accents stay apart for colour-blind
 * operators (UI-STEM-10, stem#1269).
 *
 * The accents mark a module's page icon, its dot and left border in the module
 * lists, and its chart series. The old red / orange / olive band collapsed to
 * one colour under deuteranopia (ΔE 0.3 between measure and certify in dark),
 * so the six are now Okabe–Ito hues re-cut per theme, and this file holds them
 * to it: every pair stays apart under normal, protan and deutan vision, and
 * every accent is visible on every surface.
 *
 * Accents are graphics, never text (product-stem.css), so they meet the 3:1
 * non-text bar (WCAG 1.4.11); the badge text beside them is the label tokens,
 * held to 4.5:1 here. The last test keeps the accents off text.
 */
import { readdirSync, readFileSync } from 'node:fs';
import { join, relative } from 'node:path';
import { describe, expect, it } from 'vitest';
import { contrast, linear, palette, rgb, SURFACES, src } from '../test/palette';

const MODULES = ['reflector', 'benchmark', 'servicetest', 'trafficgen', 'measure', 'certify'];

// Machado, Oliveira & Fernandes (2009), severity 1.0, applied in linear RGB.
const VISION = {
  normal: [
    [1, 0, 0],
    [0, 1, 0],
    [0, 0, 1],
  ],
  protanopia: [
    [0.152286, 1.052583, -0.204868],
    [0.114503, 0.786281, 0.099216],
    [-0.003882, -0.048116, 1.051998],
  ],
  deuteranopia: [
    [0.367322, 0.860646, -0.227968],
    [0.280085, 0.672501, 0.047413],
    [-0.01182, 0.04294, 0.968881],
  ],
} as const;

// OKLab distance x100. About 2 is a just-noticeable difference; 10 keeps two
// small dots or icons apart at a glance. The shipped set measures 10.3 at its
// closest (light, measure vs certify under deuteranopia).
const MIN_DISTANCE = 10;

type Vec = [number, number, number];

function oklab([r, g, b]: Vec): Vec {
  const l = Math.cbrt(0.4122214708 * r + 0.5363325363 * g + 0.0514459929 * b);
  const m = Math.cbrt(0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b);
  const s = Math.cbrt(0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b);
  return [
    0.2104542553 * l + 0.793617785 * m - 0.0040720468 * s,
    1.9779984951 * l - 2.428592205 * m + 0.4505937099 * s,
    0.0259040371 * l + 0.7827717662 * m - 0.808675766 * s,
  ];
}

function seen(hex: string, matrix: (typeof VISION)[keyof typeof VISION]): Vec {
  const lin = rgb(hex).map(linear);
  const clamp = (v: number) => Math.min(Math.max(v, 0), 1);
  return oklab(
    matrix.map((row) => clamp(row.reduce((sum, k, i) => sum + k * (lin[i] ?? 0), 0))) as Vec,
  );
}

function distance(a: Vec, b: Vec): number {
  return 100 * Math.hypot(a[0] - b[0], a[1] - b[1], a[2] - b[2]);
}

describe('module accents', () => {
  for (const mode of ['light', 'dark'] as const) {
    const tokens = palette(mode);
    const accents = MODULES.map((name) => {
      const hex = tokens.get(`module-${name}`);
      if (!hex) throw new Error(`--color-module-${name} is not defined in ${mode} mode`);
      return { name, hex };
    });

    it.each(Object.keys(VISION))(`${mode}: every pair stays apart under %s`, (vision) => {
      const matrix = VISION[vision as keyof typeof VISION];
      const close = accents.flatMap((a, i) =>
        accents.slice(i + 1).flatMap((b) => {
          const d = distance(seen(a.hex, matrix), seen(b.hex, matrix));
          return d < MIN_DISTANCE ? [`${a.name}/${b.name}: ${d.toFixed(1)}`] : [];
        }),
      );
      expect(close).toEqual([]);
    });

    it(`${mode}: every accent is 3:1 on every surface`, () => {
      const failing = accents.flatMap(({ name, hex }) =>
        SURFACES.flatMap((surface) => {
          const ground = tokens.get(`surface-${surface}`);
          if (!ground) return [`surface-${surface} missing`];
          const ratio = contrast(rgb(hex), rgb(ground));
          return ratio < 3 ? [`${name} on ${surface}: ${ratio.toFixed(2)}`] : [];
        }),
      );
      expect(failing).toEqual([]);
    });

    it.each(['text-primary', 'text-muted'])(
      `${mode}: badge label %s is 4.5:1 on every surface`,
      (label) => {
        const text = tokens.get(label);
        expect(text).toBeDefined();
        const failing = SURFACES.flatMap((surface) => {
          const ground = tokens.get(`surface-${surface}`);
          if (!(text && ground)) return [`surface-${surface} missing`];
          const ratio = contrast(rgb(text), rgb(ground));
          return ratio < 4.5 ? [`${surface}: ${ratio.toFixed(2)}`] : [];
        });
        expect(failing).toEqual([]);
      },
    );
  }

  it('colours no text with an accent', () => {
    // A page icon takes its accent through `iconColorClass`; any other
    // `text-module-*` would paint words in a colour measured only to 3:1.
    const walk = (dir: string): string[] =>
      readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
        const path = join(dir, entry.name);
        if (entry.isDirectory()) return walk(path);
        return /\.(tsx?|css)$/.test(entry.name) && !/\.test\.tsx?$/.test(entry.name) ? [path] : [];
      });
    const offenders = walk(src).flatMap((file) =>
      readFileSync(file, 'utf8')
        .split('\n')
        .flatMap((line, i) =>
          /(?<![\w-])text-module-/.test(line) && !/\biconColorClass:/.test(line)
            ? [`${relative(src, file)}:${i + 1}`]
            : [],
        ),
    );
    expect(offenders).toEqual([]);
  });
});

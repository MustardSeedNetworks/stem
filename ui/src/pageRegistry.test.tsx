/**
 * Guards the registry's locale contract: every route resolves real copy
 * in both locales, and every test module wears the eyebrow — a page has one
 * only when its locale namespace declares it.
 */
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { cleanup, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { MODULE_FEATURES } from './constants/moduleFeatures';
import i18n from './i18n';
import { useNavGroups } from './navGroups';
import { usePages } from './pageRegistry';

describe('page registry translations', () => {
  afterEach(async () => {
    cleanup();
    await i18n.changeLanguage('en');
  });

  it.each(['en', 'es'])('resolves page metadata in %s', async (language) => {
    await i18n.changeLanguage(language);
    const { result, unmount } = renderHook(() => usePages());

    for (const page of result.current) {
      // An unresolved key falls back to the key itself, e.g. "reflector.title".
      expect(page.label, `${page.path} label`).not.toContain('.label');
      expect(page.title, `${page.path} title`).not.toContain('.title');
      expect(page.description, `${page.path} description`).not.toContain('.description');
    }
    unmount();
  });

  // The six test modules are the reflector plus the five licensed ones, so a
  // module added to MODULE_FEATURES without its eyebrow copy fails here.
  const modulePaths = ['/reflector', ...Object.keys(MODULE_FEATURES)];

  it.each([
    ['en', 'Test module'],
    ['es', 'Módulo de prueba'],
  ])(
    'gives every test module an eyebrow in %s, and nothing else one',
    async (language, eyebrow) => {
      await i18n.changeLanguage(language);
      const { result } = renderHook(() => usePages());

      for (const page of result.current) {
        expect(page.eyebrow, `${page.path} eyebrow`).toBe(
          modulePaths.includes(page.path) ? eyebrow : undefined,
        );
      }
      expect(result.current.filter((page) => page.eyebrow).length).toBe(modulePaths.length);
    },
  );
});

describe('rail <-> header label agreement', () => {
  afterEach(async () => {
    cleanup();
    await i18n.changeLanguage('en');
  });

  it.each(['en', 'es'])(
    'labels a route the same in the rail and the registry in %s',
    async (language) => {
      await i18n.changeLanguage(language);
      const pages = renderHook(() => usePages()).result.current;
      const rail = renderHook(() => useNavGroups()).result.current;

      const railLabel = new Map(
        rail.flatMap((group) => group.items.map((item) => [item.path, item.label] as const)),
      );
      for (const page of pages) {
        expect(railLabel.get(page.path), `${page.path} rail label`).toBe(page.label);
      }
    },
  );

  it('translates the rail entries that are ordinary words, not glossary terms', async () => {
    await i18n.changeLanguage('es');
    const rail = renderHook(() => useNavGroups()).result.current;
    const byPath = new Map(
      rail.flatMap((group) => group.items.map((item) => [item.path, item.label] as const)),
    );

    // History and Security are ordinary words and translate; the module names
    // are glossary terms and must read the same in every locale.
    expect(byPath.get('/history')).toBe('Historial');
    expect(byPath.get('/account/security')).toBe('Seguridad');
    expect(byPath.get('/reflector')).toBe('Reflector');
    expect(byPath.get('/tests/benchmark')).toBe('Benchmark');
  });
});

describe('usePages', () => {
  it('lands on the reflector', () => {
    const { result } = renderHook(() => usePages());

    expect(result.current[0]?.path).toBe('/reflector');
  });

  it('declares every path exactly once', () => {
    const { result } = renderHook(() => usePages());
    const paths = result.current.map((p) => p.path);

    expect(new Set(paths).size).toBe(paths.length);
  });

  it('gives every page an icon and a component', () => {
    const { result } = renderHook(() => usePages());

    for (const page of result.current) {
      expect(page.icon, `${page.path} icon`).toBeTruthy();
      expect(page.component, `${page.path} component`).toBeDefined();
    }
  });
});

describe('usePages — lazy routes', () => {
  // Path -> the module and named export the registry lazy-loads. Kept
  // explicit so a renamed export fails here rather than on the operator's
  // first navigation; the coverage assertion below stops it going stale.
  const lazyRoutes: Record<string, () => Promise<Record<string, unknown>>> = {
    '/tests/benchmark': () => import('./pages/BenchmarkPage'),
    '/tests/servicetest': () => import('./pages/ServiceTestPage'),
    '/tests/trafficgen': () => import('./pages/TrafficGenPage'),
    '/tests/measure': () => import('./pages/MeasurePage'),
    '/tests/certify': () => import('./pages/CertifyPage'),
    '/history': () => import('./pages/HistoryPage'),
    '/account/security': () => import('./pages/account/security/SecurityPage'),
  };
  const exportNames: Record<string, string> = {
    '/tests/benchmark': 'BenchmarkPage',
    '/tests/servicetest': 'ServiceTestPage',
    '/tests/trafficgen': 'TrafficGenPage',
    '/tests/measure': 'MeasurePage',
    '/tests/certify': 'CertifyPage',
    '/history': 'HistoryPage',
    '/account/security': 'SecurityPage',
  };

  it('covers every non-eager route', () => {
    const { result } = renderHook(() => usePages());
    const deferred = result.current.map((p) => p.path).filter((path) => path !== '/reflector');

    expect(deferred.sort()).toEqual(Object.keys(lazyRoutes).sort());
  });

  it.each(Object.keys(lazyRoutes))('%s still exports the component it lazy-loads', async (path) => {
    const mod = await lazyRoutes[path]();

    expect(mod[exportNames[path]], `${path}: ${exportNames[path]}`).toBeTypeOf('function');
  });
});

/**
 * CI's phone-width job visits a hand-listed set of routes: the reusable
 * workflow cannot read this registry. A page added here but not there would
 * ship without ever being checked at 390px, with the job still green.
 */
describe('pageRegistry <-> phone-width routes', () => {
  it('checks every registered page at phone width', () => {
    const ci = readFileSync(resolve(import.meta.dirname, '../../.github/workflows/ci.yml'), 'utf8');
    const job = ci.slice(ci.indexOf('\n  phone-width:\n'));
    const routes = /^ {6}routes: '(.+)'$/m.exec(job)?.[1];
    expect(routes, 'phone-width job has no routes input').toBeDefined();

    const byPath = (a: string, b: string) => a.localeCompare(b);
    const { result: pages } = renderHook(() => usePages());
    expect((JSON.parse(routes ?? '[]') as string[]).sort(byPath)).toEqual(
      pages.current.map((page) => page.path).sort(byPath),
    );
  });
});

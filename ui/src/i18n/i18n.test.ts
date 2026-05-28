/**
 * i18n init tests — smoke-test the resource loading + language
 * detection wiring so a botched namespace addition or a malformed
 * locale JSON gets caught in CI rather than in production. Mirrors
 * the niac-go suite (Phase 6) for cross-product consistency.
 */

import i18n from 'i18next';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { defaultNs, languages, namespaces } from './index';

describe('i18n configuration', () => {
  it('declares the expected set of supported languages', () => {
    const codes = languages.map((l) => l.code);
    expect(codes).toEqual(['en', 'es']);
  });

  it('declares the expected set of namespaces', () => {
    expect([...namespaces]).toEqual([
      'common',
      'errors',
      'modules',
      'recovery',
      'security',
      'settings',
      'setup',
      'cli',
      'params',
      'help',
    ]);
  });

  it('uses common as the default namespace', () => {
    expect(defaultNs).toBe('common');
  });

  it('loads EN resources for every declared namespace', () => {
    for (const ns of namespaces) {
      const bundle = i18n.getResourceBundle('en', ns);
      expect(bundle, `EN bundle missing for ${ns}`).toBeTruthy();
      expect(Object.keys(bundle).length, `EN bundle empty for ${ns}`).toBeGreaterThan(0);
    }
  });

  it('loads ES resources for every declared namespace', () => {
    for (const ns of namespaces) {
      const bundle = i18n.getResourceBundle('es', ns);
      expect(bundle, `ES bundle missing for ${ns}`).toBeTruthy();
      expect(Object.keys(bundle).length, `ES bundle empty for ${ns}`).toBeGreaterThan(0);
    }
  });
});

describe('i18n <html lang> sync', () => {
  const originalLanguage = i18n.language;
  const originalDocLang = document.documentElement.lang;

  afterEach(async () => {
    await i18n.changeLanguage(originalLanguage);
    document.documentElement.lang = originalDocLang;
  });

  it('updates document.documentElement.lang when language changes', async () => {
    await i18n.changeLanguage('es');
    expect(document.documentElement.lang).toBe('es');

    await i18n.changeLanguage('en');
    expect(document.documentElement.lang).toBe('en');
  });
});

describe('i18n key resolution', () => {
  beforeEach(async () => {
    await i18n.changeLanguage('en');
  });

  it('resolves a top-level common key', () => {
    // stem's common.json structure ships labels under buttons.* —
    // pick one that's known to exist and is reasonably stable.
    const result = i18n.t('common:buttons.save', { defaultValue: '__missing__' });
    expect(result).not.toBe('__missing__');
  });

  it('resolves Phase-7 plural keys', () => {
    expect(i18n.t('common:plurals.testCount', { count: 5 })).toBe('Found 5 tests');
    expect(i18n.t('common:plurals.testCount', { count: 1 })).toBe('Found 1 test');
  });

  it('flips translations when language changes', async () => {
    const en = i18n.t('common:plurals.testCount', { count: 1 });
    await i18n.changeLanguage('es');
    const es = i18n.t('common:plurals.testCount', { count: 1 });
    expect(es).not.toBe(en);
    expect(es).toMatch(/prueba/i);
  });
});

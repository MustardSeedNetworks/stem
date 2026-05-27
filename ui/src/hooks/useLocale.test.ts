/**
 * useLocale tests — covers the BCP-47 normalization that every Intl.*
 * call site in the codebase relies on. Mirrors the niac-go suite
 * (Phase 6) for cross-product consistency.
 *
 * NOTE: tests don't exercise the React-Context propagation path here
 * because vitest's jsdom + Vite's module-resolution behaviour produces
 * separate i18next instances when imported directly vs via react-
 * i18next, so useTranslation()'s subscription doesn't see changes
 * made to the global singleton. Instead, the hook is tested by
 * invoking the function directly with a mocked i18n instance — proves
 * the BCP-47 mapping logic, which is the unit under test.
 */

import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: vi.fn(),
}));

import { useTranslation } from 'react-i18next';
import { useLocale } from './useLocale';

const mockedUseTranslation = vi.mocked(useTranslation);

describe('useLocale', () => {
  function setLanguage(language: string) {
    mockedUseTranslation.mockReturnValue({
      // biome-ignore lint/suspicious/noExplicitAny: minimal test stub
      i18n: { language } as any,
      // biome-ignore lint/suspicious/noExplicitAny: minimal test stub
      t: ((key: string) => key) as any,
      ready: true,
    });
  }

  it('returns en-US when i18next language is en', () => {
    setLanguage('en');
    const { result } = renderHook(() => useLocale());
    expect(result.current).toBe('en-US');
  });

  it('returns es-ES when i18next language is es', () => {
    setLanguage('es');
    const { result } = renderHook(() => useLocale());
    expect(result.current).toBe('es-ES');
  });

  it('returns the raw code for unknown languages (still a valid BCP-47)', () => {
    setLanguage('fr');
    const { result } = renderHook(() => useLocale());
    expect(result.current).toBe('fr');
  });
});

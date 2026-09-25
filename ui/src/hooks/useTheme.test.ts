/**
 * The theme a fresh profile opens in is the OS's (fleet decision 2026-09-15,
 * UI-STEM-4); only an explicit toggle overrides it.
 */
import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { initThemeFromStorage, useTheme } from './useTheme';

function mockOsScheme(scheme: 'light' | 'dark'): void {
  vi.stubGlobal(
    'matchMedia',
    vi.fn((query: string) => ({
      matches: query === '(prefers-color-scheme: dark)' && scheme === 'dark',
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    })),
  );
}

describe('useTheme', () => {
  beforeEach(() => {
    window.localStorage.clear();
    document.documentElement.classList.remove('dark');
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it.each(['light', 'dark'] as const)(
    'follows an OS set to %s when nothing is stored',
    (scheme) => {
      mockOsScheme(scheme);

      initThemeFromStorage();
      const { result } = renderHook(() => useTheme());

      expect(result.current.theme).toBe('system');
      expect(result.current.effectiveTheme).toBe(scheme);
      expect(document.documentElement.classList.contains('dark')).toBe(scheme === 'dark');
    },
  );

  it('keeps an explicit choice over the OS, and persists it', () => {
    mockOsScheme('dark');
    const { result } = renderHook(() => useTheme());

    act(() => result.current.toggleTheme());

    expect(result.current.effectiveTheme).toBe('light');
    expect(window.localStorage.getItem('stem-theme')).toBe('light');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });
});

/**
 * Theme Management Hook
 *
 * Manages application theme with support for light, dark, and system modes.
 *
 * Features:
 * - Light and dark theme modes
 * - Automatic system theme detection via prefers-color-scheme
 * - Persistent theme storage in localStorage (key: 'stem-theme')
 * - Automatic system theme change detection while in 'system' mode
 * - Theme toggling functionality
 *
 * The state is one module-level store (useSyncExternalStore), so the rail toggle
 * and the Settings control share it.
 *
 * The theme is applied by adding/removing the 'dark' class on the document
 * root element, which Tailwind CSS and our index.css use to switch
 * between :root (light) and .dark (dark) custom properties.
 *
 * With no stored preference the theme is 'system', so a fresh profile
 * follows the OS (fleet decision 2026-09-15: OS default plus a persisted
 * toggle). The toggle stores an explicit 'light' or 'dark'.
 *
 * Usage:
 * ```typescript
 * const { theme, isDark, toggleTheme, setTheme } = useTheme();
 *
 * <button onClick={toggleTheme}>Toggle Theme</button>
 * <button onClick={() => setTheme('system')}>Use System Theme</button>
 * ```
 */

import { useCallback, useSyncExternalStore } from 'react';

/** Theme mode options */
export type Theme = 'light' | 'dark' | 'system';

/** localStorage key for theme persistence */
const STORAGE_KEY = 'stem-theme';

/**
 * Detects the system's preferred color scheme.
 *
 * @returns 'dark' if system prefers dark mode, 'light' otherwise.
 *          Defaults to 'dark' if detection is unavailable.
 */
function getSystemTheme(): 'light' | 'dark' {
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return 'dark';
}

/**
 * Applies the theme to the document root element.
 * Resolves 'system' theme to actual light/dark preference.
 */
function applyTheme(theme: Theme): void {
  if (typeof document === 'undefined') {
    return;
  }
  const root: HTMLElement = document.documentElement;
  const effectiveTheme: 'light' | 'dark' = theme === 'system' ? getSystemTheme() : theme;

  if (effectiveTheme === 'dark') {
    root.classList.add('dark');
  } else {
    root.classList.remove('dark');
  }
}

/**
 * Read the stored theme from localStorage. Defaults to 'system' when no
 * preference is stored.
 */
function readStoredTheme(): Theme {
  if (typeof window === 'undefined') {
    return 'system';
  }
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY) as Theme | null;
    if (stored === 'light' || stored === 'dark' || stored === 'system') {
      return stored;
    }
  } catch {
    // localStorage may be unavailable
  }
  return 'system';
}

// One store for every consumer. The rail's quick toggle and the Settings
// control both read and write it, so they cannot disagree about the theme.
const listeners = new Set<() => void>();
let current: Theme | null = null;

function snapshot(): Theme {
  current ??= readStoredTheme();
  return current;
}

function emit(): void {
  for (const listener of listeners) {
    listener();
  }
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  const media = window.matchMedia('(prefers-color-scheme: dark)');
  // While the theme is 'system' the OS is the source: re-apply and notify.
  const onSystemChange = (): void => {
    if (snapshot() === 'system') {
      applyTheme('system');
      listener();
    }
  };
  media.addEventListener('change', onSystemChange);
  return (): void => {
    listeners.delete(listener);
    media.removeEventListener('change', onSystemChange);
  };
}

function setStoredTheme(next: Theme): void {
  current = next;
  try {
    window.localStorage.setItem(STORAGE_KEY, next);
  } catch {
    // Ignore storage failures
  }
  applyTheme(next);
  emit();
}

/**
 * Custom hook for managing application theme.
 *
 * @returns Theme state and control functions
 */
export function useTheme(): {
  theme: Theme;
  effectiveTheme: 'light' | 'dark';
  setTheme: (newTheme: Theme) => void;
  toggleTheme: () => void;
  isDark: boolean;
} {
  const theme = useSyncExternalStore(subscribe, snapshot, () => 'system' as const);
  const systemTheme = useSyncExternalStore(subscribe, getSystemTheme, () => 'dark' as const);
  const effectiveTheme = theme === 'system' ? systemTheme : theme;

  const toggleTheme = useCallback((): void => {
    setStoredTheme(effectiveTheme === 'dark' ? 'light' : 'dark');
  }, [effectiveTheme]);

  return {
    theme,
    effectiveTheme,
    setTheme: setStoredTheme,
    toggleTheme,
    isDark: effectiveTheme === 'dark',
  };
}

/**
 * Apply the initial theme synchronously, before React mounts. This avoids
 * a flash of the wrong theme at boot.
 *
 * Call from main.tsx before `createRoot(...).render(...)`.
 */
export function initThemeFromStorage(): void {
  current = readStoredTheme();
  applyTheme(current);
}

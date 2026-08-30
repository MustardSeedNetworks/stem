/**
 * ModuleSettingsContext tests.
 *
 * This context decides which modules a stem will run and which tests inside
 * them, and it persists that choice across sessions. The behaviours that matter
 * are the ones where a wrong answer silently changes what a test run does:
 * a toggle leaking across modules, `getAllEnabledTests` returning tests from a
 * disabled module, or a corrupt stored value taking the app down at boot.
 */

import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ModuleSettingsProvider, useModuleSettings } from './ModuleSettingsContext';

const STORAGE_KEY = 'stem-module-settings';

function wrapper({ children }: { children: ReactNode }) {
  return <ModuleSettingsProvider>{children}</ModuleSettingsProvider>;
}

function renderSettings() {
  return renderHook(() => useModuleSettings(), { wrapper });
}

function stored() {
  const raw = window.localStorage.getItem(STORAGE_KEY);
  return raw ? JSON.parse(raw) : null;
}

beforeEach(() => {
  window.localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('initial state', () => {
  it('starts from the built-in modules when nothing is stored', () => {
    const { result } = renderSettings();

    expect(result.current.modules.length).toBeGreaterThan(0);
    expect(result.current.modules.map((m) => m.name)).toContain('reflector');
  });

  it('gives every module an idle status and no results', () => {
    const { result } = renderSettings();

    for (const mod of result.current.modules) {
      expect(result.current.moduleStatuses[mod.name]).toEqual({
        status: 'idle',
        currentTest: null,
      });
      expect(result.current.moduleResults[mod.name]).toBeNull();
    }
  });

  it('persists the modules on mount', () => {
    renderSettings();

    expect(stored()).not.toBeNull();
  });

  it('restores a stored configuration over the defaults', () => {
    const saved = [
      {
        name: 'reflector',
        displayName: 'Reflector',
        description: 'stored',
        color: 'x',
        standard: 'Loopback',
        enabled: false,
        autoStart: true,
        tests: [{ id: 'reflect', name: 'Reflect', description: 'd', enabled: false }],
      },
    ];
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(saved));

    const { result } = renderSettings();

    expect(result.current.modules).toHaveLength(1);
    expect(result.current.modules[0].enabled).toBe(false);
    expect(result.current.modules[0].autoStart).toBe(true);
  });

  it('falls back to the defaults when the stored value is not JSON', () => {
    // Operator settings survive across releases; a value an older build wrote
    // must not stop the app booting.
    window.localStorage.setItem(STORAGE_KEY, '{ not json');

    const { result } = renderSettings();

    expect(result.current.modules.map((m) => m.name)).toContain('reflector');
  });
});

describe('toggling', () => {
  it('enables and disables one module without touching the others', () => {
    const { result } = renderSettings();
    const others = result.current.modules
      .filter((m) => m.name !== 'reflector')
      .map((m) => m.enabled);

    act(() => result.current.toggleModule('reflector', false));

    const after = result.current.modules;
    expect(after.find((m) => m.name === 'reflector')?.enabled).toBe(false);
    expect(after.filter((m) => m.name !== 'reflector').map((m) => m.enabled)).toEqual(others);
  });

  it('sets autostart independently of enabled', () => {
    const { result } = renderSettings();

    act(() => result.current.toggleAutoStart('reflector', true));

    const mod = result.current.modules.find((m) => m.name === 'reflector');
    expect(mod?.autoStart).toBe(true);
    expect(mod?.enabled).toBe(true);
  });

  it('toggles one test without touching its siblings or other modules', () => {
    const { result } = renderSettings();
    const reflectorSiblingsBefore = result.current.modules
      .find((m) => m.name === 'reflector')
      ?.tests.filter((t) => t.id !== 'reflect')
      .map((t) => t.enabled);
    const benchmarkBefore = JSON.stringify(
      result.current.modules.find((m) => m.name === 'benchmark')?.tests,
    );

    act(() => result.current.toggleTest('reflector', 'reflect', false));

    const reflector = result.current.modules.find((m) => m.name === 'reflector');
    expect(reflector?.tests.find((t) => t.id === 'reflect')?.enabled).toBe(false);
    expect(reflector?.tests.filter((t) => t.id !== 'reflect').map((t) => t.enabled)).toEqual(
      reflectorSiblingsBefore,
    );
    // Another module's tests must be byte-for-byte untouched.
    expect(JSON.stringify(result.current.modules.find((m) => m.name === 'benchmark')?.tests)).toBe(
      benchmarkBefore,
    );
  });

  it('toggles a shared test id only in the named module', () => {
    // No two built-in modules share a test id today, so the assertion above
    // cannot actually catch a missing `mod.name === moduleName` check — the
    // map would rewrite every module and change nothing observable. A stored
    // configuration is arbitrary JSON and can share ids, which is what makes
    // the guard testable at all.
    window.localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify([
        {
          name: 'alpha',
          displayName: 'Alpha',
          description: '',
          color: 'x',
          standard: 's',
          enabled: true,
          autoStart: false,
          tests: [{ id: 'shared', name: 'Shared', description: '', enabled: true }],
        },
        {
          name: 'beta',
          displayName: 'Beta',
          description: '',
          color: 'x',
          standard: 's',
          enabled: true,
          autoStart: false,
          tests: [{ id: 'shared', name: 'Shared', description: '', enabled: true }],
        },
      ]),
    );

    const { result } = renderSettings();

    act(() => result.current.toggleTest('alpha', 'shared', false));

    expect(result.current.modules.find((m) => m.name === 'alpha')?.tests[0].enabled).toBe(false);
    expect(result.current.modules.find((m) => m.name === 'beta')?.tests[0].enabled).toBe(true);
  });

  it('ignores a module name that does not exist', () => {
    const { result } = renderSettings();
    const before = JSON.stringify(result.current.modules);

    act(() => result.current.toggleModule('no-such-module', false));

    expect(JSON.stringify(result.current.modules)).toBe(before);
  });

  it('persists a toggle', () => {
    const { result } = renderSettings();

    act(() => result.current.toggleModule('reflector', false));

    const saved = stored() as Array<{ name: string; enabled: boolean }>;
    expect(saved.find((m) => m.name === 'reflector')?.enabled).toBe(false);
  });
});

describe('enabled-test queries', () => {
  it('returns only the enabled tests of a module', () => {
    const { result } = renderSettings();

    act(() => result.current.toggleTest('reflector', 'echo', false));

    const ids = result.current.getEnabledTests('reflector').map((t) => t.id);
    expect(ids).toContain('reflect');
    expect(ids).not.toContain('echo');
  });

  it('returns an empty list for an unknown module', () => {
    expect(renderSettings().result.current.getEnabledTests('no-such-module')).toEqual([]);
  });

  it('omits every test of a disabled module from getAllEnabledTests', () => {
    // The one that actually changes what a run does: a disabled module must
    // contribute nothing, however its individual tests are set.
    const { result } = renderSettings();
    expect(result.current.getAllEnabledTests().some((e) => e.module === 'reflector')).toBe(true);

    act(() => result.current.toggleModule('reflector', false));

    expect(result.current.getAllEnabledTests().some((e) => e.module === 'reflector')).toBe(false);
  });

  it('omits a disabled test of an enabled module', () => {
    const { result } = renderSettings();

    act(() => result.current.toggleTest('reflector', 'echo', false));

    const entries = result.current.getAllEnabledTests();
    expect(entries.some((e) => e.module === 'reflector' && e.test.id === 'echo')).toBe(false);
    expect(entries.some((e) => e.module === 'reflector' && e.test.id === 'reflect')).toBe(true);
  });
});

describe('status and results', () => {
  it('records a module status without disturbing the others', () => {
    const { result } = renderSettings();

    act(() =>
      result.current.setModuleStatus('reflector', { status: 'running', currentTest: 'reflect' }),
    );

    expect(result.current.moduleStatuses.reflector).toEqual({
      status: 'running',
      currentTest: 'reflect',
    });
    expect(result.current.moduleStatuses.benchmark).toEqual({ status: 'idle', currentTest: null });
  });

  it('sets, updates and clears results for one module', () => {
    const { result } = renderSettings();
    const results = { summary: 'first' } as never;

    act(() => result.current.setModuleResults('reflector', results));
    expect(result.current.moduleResults.reflector).toEqual(results);

    act(() =>
      result.current.updateModuleResults(
        'reflector',
        (prev) =>
          ({
            ...(prev as object),
            summary: 'second',
          }) as never,
      ),
    );
    expect(result.current.moduleResults.reflector).toEqual({ summary: 'second' });

    act(() => result.current.clearModuleResults('reflector'));
    expect(result.current.moduleResults.reflector).toBeNull();
  });

  it('hands the updater null when the module has no results yet', () => {
    const { result } = renderSettings();
    const seen: Array<unknown> = [];

    act(() =>
      result.current.updateModuleResults('reflector', (prev) => {
        seen.push(prev);
        return null;
      }),
    );

    expect(seen).toEqual([null]);
  });
});

describe('resetToDefaults', () => {
  it('restores the built-in configuration and drops the stored copy', () => {
    const { result } = renderSettings();
    act(() => result.current.toggleModule('reflector', false));
    expect(stored()).not.toBeNull();

    act(() => result.current.resetToDefaults());

    expect(result.current.modules.find((m) => m.name === 'reflector')?.enabled).toBe(true);
    // The effect re-persists immediately after the reset, so the stored copy
    // has to be the defaults rather than merely absent.
    const saved = stored() as Array<{ name: string; enabled: boolean }>;
    expect(saved.find((m) => m.name === 'reflector')?.enabled).toBe(true);
  });
});

describe('useModuleSettings outside a provider', () => {
  it('throws rather than handing back a silent default', () => {
    expect(() => renderHook(() => useModuleSettings())).toThrow(
      'useModuleSettings must be used within a ModuleSettingsProvider',
    );
  });
});

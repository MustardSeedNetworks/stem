/**
 * ModuleSettingsContext.stability.test.tsx — pins the invariant the memos in
 * this file exist to hold, so removing them is a proof rather than a hope.
 *
 * The invariant is *identity*, not render count: the context value and every
 * action on it keep the same reference across a re-render that changed none of
 * the state they derive from. That is what makes them safe to put in a
 * dependency array or hand to a memoised child. (Render count is a different
 * claim, and not one these memos ever made — a consumer rendered as a plain
 * child re-renders with its parent whatever the context value does.)
 *
 * The React Compiler is supposed to do the same job. This test is how we find
 * out, because it fails loudly if the identities start churning.
 */
import { act, render, screen } from '@testing-library/react';
import { type ReactElement, useState } from 'react';
import { describe, expect, it } from 'vitest';
import { ModuleSettingsProvider, useModuleSettings } from './ModuleSettingsContext';

let lastValue: unknown = null;

function Consumer(): ReactElement {
  const value = useModuleSettings();
  lastValue = value;
  return <span data-testid="consumer">{value.modules.length}</span>;
}

let bumpParent: (() => void) | null = null;

function Parent(): ReactElement {
  const [tick, setTick] = useState(0);
  bumpParent = () => setTick((n) => n + 1);
  return (
    <ModuleSettingsProvider>
      <span data-testid="tick">{tick}</span>
      <Consumer />
    </ModuleSettingsProvider>
  );
}

describe('ModuleSettingsContext — value identity', () => {
  it('hands consumers the same value object across an unrelated re-render', () => {
    render(<Parent />);
    const valueAtMount = lastValue;

    act(() => {
      bumpParent?.();
    });

    expect(screen.getByTestId('tick').textContent).toBe('1');
    expect(lastValue).toBe(valueAtMount);
  });

  it('keeps every action identity stable across an unrelated re-render', () => {
    render(<Parent />);
    const before = lastValue as Record<string, unknown>;
    const actions = [
      'toggleModule',
      'toggleAutoStart',
      'toggleTest',
      'setModuleStatus',
      'setModuleResults',
      'updateModuleResults',
      'clearModuleResults',
      'getEnabledTests',
      'getAllEnabledTests',
      'resetToDefaults',
    ];
    const identities = actions.map((name) => before[name]);

    act(() => {
      bumpParent?.();
    });

    const after = lastValue as Record<string, unknown>;
    actions.forEach((name, index) => {
      expect(after[name], `${name} changed identity`).toBe(identities[index]);
    });
  });
});

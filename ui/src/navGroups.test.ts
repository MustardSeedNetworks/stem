/**
 * navGroups <-> pageRegistry parity guard.
 */

import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { useNavGroups } from './navGroups';
import { usePages } from './pageRegistry';

describe('navGroups <-> pageRegistry parity', () => {
  const pages = renderHook(() => usePages()).result.current;
  const navGroups = renderHook(() => useNavGroups()).result.current;

  const navPaths = new Set(navGroups.flatMap((group) => group.items.map((item) => item.path)));
  const routePaths = new Set(pages.map((page) => page.path));

  it('exposes every routable page in the sidebar', () => {
    const missing = pages.map((page) => page.path).filter((path) => !navPaths.has(path));
    expect(missing, `pages missing from navGroups: ${missing.join(', ')}`).toEqual([]);
  });
  it('has no sidebar entries pointing at a non-existent route', () => {
    const orphaned = [...navPaths].filter((path) => !routePaths.has(path));
    expect(orphaned, `navGroups entries without a page: ${orphaned.join(', ')}`).toEqual([]);
  });
});

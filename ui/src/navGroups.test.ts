import { describe, expect, it } from 'vitest';
import { navGroups } from './navGroups';
import { pages } from './pageRegistry';

describe('navGroups <-> pageRegistry parity', () => {
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

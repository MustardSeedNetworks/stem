/**
 * @fileoverview useOpenHelp — binds the help drawer to a route when it opens.
 * @copyright 2025 Mustard Seed Networks. All rights reserved.
 * @license Proprietary
 */

import { useCallback } from 'react';
import { matchPath } from 'react-router';
import { usePages } from './pageRegistry';
import { useShellStore } from './stores/shell-store';

/**
 * Resolve a pathname to the route table entry that serves it.
 */
export function routeForPathname(paths: readonly string[], pathname: string): string | null {
  return paths.find((path) => matchPath(path, pathname)) ?? null;
}

/**
 * useOpenHelp returns the opener every help button shares.
 *
 * It reads the address bar rather than the committed route on purpose.
 * react-router commits a navigation in a transition, so a help click that
 * follows a nav click can be rendered while the committed route is still the
 * page the user is leaving; binding the drawer to that route made it open and
 * then be discarded by the route change already in flight, and nothing
 * re-opened it (#1305). The URL is the route the user is on their way to.
 */
export function useOpenHelp(): () => void {
  const pages = usePages();
  const setHelpRoute = useShellStore((s) => s.setHelpRoute);
  return useCallback(() => {
    setHelpRoute(
      routeForPathname(
        pages.map((page) => page.path),
        window.location.pathname,
      ),
    );
  }, [pages, setHelpRoute]);
}

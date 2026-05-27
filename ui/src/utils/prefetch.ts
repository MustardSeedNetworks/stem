/**
 * Route prefetch — hover-triggered API data warm-up.
 *
 * Called by the shared Sidebar on nav-item hover. Stem-specific route
 * map is intentionally minimal today; populate ROUTE_PREFETCH_MAP with
 * any stem routes whose first paint is bottlenecked by an API round-trip.
 *
 * Safe to call with any path — unknown paths no-op.
 */

type PrefetchFn = () => Promise<unknown>;

const ROUTE_PREFETCH_MAP: Record<string, PrefetchFn[]> = {
  // Populate when stem has slow-to-paint routes worth pre-warming.
};

const prefetched = new Set<string>();

/**
 * Prefetch API data for a route during browser idle time.
 * Each route is only prefetched once per session to avoid redundant requests.
 */
export function prefetchRoute(path: string): void {
  if (prefetched.has(path)) return;

  const fetchers = ROUTE_PREFETCH_MAP[path];
  if (!fetchers) return;

  prefetched.add(path);

  const run = (): void => {
    for (const fn of fetchers) {
      fn().catch(() => {
        // Silently ignore prefetch failures
      });
    }
  };

  if ('requestIdleCallback' in window) {
    window.requestIdleCallback(run, { timeout: 2000 });
  } else {
    setTimeout(run, 100);
  }
}

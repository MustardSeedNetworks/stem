import type { FC } from 'react';

/**
 * Suspense fallback for lazy-loaded routed pages.
 */
export const PageLoader: FC = () => (
  <div className="flex-center min-h-[400px]">
    <div className="flex flex-col items-center gap-default">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-brand-primary border-t-transparent" />
      <p className="text-sm text-text-muted">Loading...</p>
    </div>
  </div>
);

/**
 * LicenseContext — the active licence, read once and shared.
 *
 * Stem knew its entitlement in exactly one place: the licence panel inside the
 * settings drawer, which fetched `/api/v1/license` for itself. So a Pro-gated
 * module could not tell an operator it was gated until the run came back 402
 * (UI-STEM-15). The status is fetched once here and read by every module gate.
 *
 * The shape is the generated wire type, not a hand-written copy: the panel's
 * own copy was missing `platform`, `licenseKey` and `message` and made
 * `expiresAt` required where the DTO omits a zero time. A hand-typed DTO is
 * compared to nothing, which is how seed's gates read `undefined` for a
 * release (seed#2688).
 */

import { createContext, type ReactNode, useCallback, useContext, useEffect, useState } from 'react';
import { authFetch } from '../stores/auth-store';
import type { LicenseStatus } from '../types/generated/license-status';

export const LICENSE_ENDPOINT = '/api/v1/license';

interface LicenseContextValue {
  status: LicenseStatus | null;
  /** True until the first fetch settles, however it settles. */
  loading: boolean;
  refresh: () => Promise<void>;
  /** True iff the active licence grants the feature. */
  hasFeature: (feature: string) => boolean;
}

const LicenseContext = createContext<LicenseContextValue | null>(null);

export function LicenseProvider({ children }: { children: ReactNode }): React.ReactElement {
  const [status, setStatus] = useState<LicenseStatus | null>(null);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async (): Promise<void> => {
    try {
      const response = await authFetch(LICENSE_ENDPOINT);
      if (response.ok) {
        setStatus((await response.json()) as LicenseStatus);
      }
    } catch {
      // An unreachable daemon is not an entitlement: leave the last known
      // status alone and let the gate treat unknown as ungated, so a network
      // blip cannot put a paying operator behind a pitch.
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const hasFeature = useCallback(
    (feature: string): boolean => status?.features?.includes(feature) === true,
    [status],
  );

  return (
    <LicenseContext.Provider value={{ status, loading, refresh, hasFeature }}>
      {children}
    </LicenseContext.Provider>
  );
}

export function useLicense(): LicenseContextValue {
  const ctx = useContext(LicenseContext);
  if (!ctx) {
    throw new Error('useLicense() must be called inside <LicenseProvider>');
  }
  return ctx;
}

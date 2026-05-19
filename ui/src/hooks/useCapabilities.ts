import { useEffect, useState } from 'react';
import { logWarn } from '../utils/logger';

/**
 * Capability descriptor returned by the backend's unauthenticated
 * /api/v1/capabilities endpoint. When `supported` is false, `reason`
 * carries a short operator-facing explanation the UI surfaces in a
 * platform-guard banner — see ReflectorPage for the canonical
 * consumer.
 */
export interface CapabilityInfo {
  supported: boolean;
  reason?: string;
}

/**
 * Shape of the JSON returned by /api/v1/capabilities. Mirrors the Go
 * CapabilitiesResponse in internal/api/handlers_capabilities.go.
 */
export interface Capabilities {
  reflector: CapabilityInfo;
  testMaster: CapabilityInfo;
}

/**
 * Default to "everything supported". Used while the request is in
 * flight or if the endpoint is unavailable (older builds, network
 * error). This keeps the UI optimistic — we only show the platform
 * guard once the server has explicitly told us a capability is off.
 */
const FALLBACK: Capabilities = {
  reflector: { supported: true },
  testMaster: { supported: true },
};

function isCapabilityInfo(value: unknown): value is CapabilityInfo {
  if (typeof value !== 'object' || value === null) {
    return false;
  }
  const candidate = value as Record<string, unknown>;
  if (typeof candidate.supported !== 'boolean') {
    return false;
  }
  if ('reason' in candidate && typeof candidate.reason !== 'string') {
    return false;
  }
  return true;
}

function isCapabilities(value: unknown): value is Capabilities {
  if (typeof value !== 'object' || value === null) {
    return false;
  }
  const candidate = value as Record<string, unknown>;
  return isCapabilityInfo(candidate.reflector) && isCapabilityInfo(candidate.testMaster);
}

/**
 * Fetches capability metadata from /api/v1/capabilities once on
 * mount. Falls back to "everything supported" so callers don't need a
 * loading state — the banner only appears once the server has
 * explicitly returned `supported: false`.
 *
 * Mirrors the surface of {@link import('./useBuildVersion').useBuildVersion}.
 */
export function useCapabilities(): Capabilities {
  const [data, setData] = useState<Capabilities>(FALLBACK);
  useEffect(() => {
    let cancelled = false;
    fetch('/api/v1/capabilities', { headers: { Accept: 'application/json' } })
      .then(async (response) => {
        if (!response.ok) {
          throw new Error(`status ${response.status}`);
        }
        return response.json();
      })
      .then((body) => {
        if (cancelled) {
          return;
        }
        if (isCapabilities(body)) {
          setData(body);
        } else {
          logWarn('Unexpected /api/v1/capabilities payload shape', {
            component: 'useCapabilities',
          });
        }
      })
      .catch((err: unknown) => {
        if (cancelled) {
          return;
        }
        logWarn('Failed to fetch /api/v1/capabilities', {
          component: 'useCapabilities',
          additionalData: { error: err instanceof Error ? err.message : String(err) },
        });
      });
    return () => {
      cancelled = true;
    };
  }, []);
  return data;
}

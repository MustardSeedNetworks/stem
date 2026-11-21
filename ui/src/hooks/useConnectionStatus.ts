/**
 * useConnectionStatus — whether the Stem daemon answers.
 *
 * Polls the unauthenticated /__version endpoint. It is independent of the SSE
 * stream, which only reports once a subscription has been opened and says
 * nothing about a daemon that went away between events. A failed check is
 * retried sooner than a healthy one is re-confirmed, so the banner clears
 * promptly once the daemon is back.
 *
 * Mirrors niac's hooks/useConnectionStatus.ts.
 */
import { useEffect, useState } from 'react';

export type ConnectionState = 'connected' | 'disconnected' | 'checking';

export const HEALTHY_INTERVAL_MS = 15_000;
export const RETRY_INTERVAL_MS = 5_000;
const TIMEOUT_MS = 5_000;

export function useConnectionStatus(): ConnectionState {
  const [status, setStatus] = useState<ConnectionState>('checking');

  useEffect(() => {
    let disposed = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    const controller = new AbortController();

    const check = async (): Promise<void> => {
      const attempt = new AbortController();
      const deadline = setTimeout(() => attempt.abort(), TIMEOUT_MS);
      const onDispose = (): void => attempt.abort();
      controller.signal.addEventListener('abort', onDispose);
      let reachable = false;
      try {
        const response = await fetch('/__version', {
          cache: 'no-store',
          headers: { Accept: 'application/json' },
          signal: attempt.signal,
        });
        reachable = response.ok;
      } catch {
        reachable = false;
      } finally {
        clearTimeout(deadline);
        controller.signal.removeEventListener('abort', onDispose);
      }
      if (disposed) {
        return;
      }
      setStatus(reachable ? 'connected' : 'disconnected');
      timer = setTimeout(
        () => {
          void check();
        },
        reachable ? HEALTHY_INTERVAL_MS : RETRY_INTERVAL_MS,
      );
    };

    void check();
    return () => {
      disposed = true;
      clearTimeout(timer);
      controller.abort();
    };
  }, []);

  return status;
}

/**
 * useSse — long-lived subscription to the stem SSE channel (#296).
 *
 * Opens an EventSource against /api/v1/events, validates every frame
 * via the valibot schemas in `@/schemas/sse`, and dispatches typed
 * payloads to caller-supplied handlers. Invalid frames are dropped
 * and logged via console.warn — the subscriber never sees an
 * unexpected shape.
 *
 * EventSource handles reconnect itself with exponential backoff;
 * we don't add another layer. The hook just owns the lifecycle.
 *
 * Mirrors seed's ui/src/hooks/useSse.ts in shape (separate stem
 * codebase, similar API).
 */
import { useEffect, useRef, useState } from 'react';

import {
  parseSseEnvelope,
  parseSseModeChanged,
  parseSseReflectorStats,
  parseSseTestProgress,
  type SseModeChanged,
  type SseReflectorStats,
  type SseTestProgress,
} from '@/schemas/sse';

export const SSE_ENDPOINT = '/api/v1/events';

export type SseConnectionStatus =
  | 'idle' // Pre-connection or hook disabled
  | 'connecting' // EventSource is opening
  | 'connected' // Live and receiving frames
  | 'reconnecting' // Browser is auto-reconnecting after a drop
  | 'closed'; // Closed by caller or fatal error

export interface UseSseOptions {
  /** When false, the hook does not open a connection. Useful for
   *  gating behind authentication or feature flags. */
  enabled?: boolean;
  /** Called for every mode_changed frame. */
  onModeChanged?: (frame: SseModeChanged) => void;
  /** Called for every reflector_stats frame (~1Hz when reflector is active). */
  onReflectorStats?: (frame: SseReflectorStats) => void;
  /** Called for every test_progress frame. */
  onTestProgress?: (frame: SseTestProgress) => void;
}

export interface UseSseResult {
  status: SseConnectionStatus;
}

/**
 * useSse opens the SSE stream and dispatches typed frames to the
 * handlers passed in `options`. Returns a `status` for the UI to
 * render connection state if it wants to (e.g. a header pip).
 */
export function useSse(options: UseSseOptions = {}): UseSseResult {
  const { enabled = true, onModeChanged, onReflectorStats, onTestProgress } = options;
  const [status, setStatus] = useState<SseConnectionStatus>('idle');

  // Refs let the EventSource's listeners see the *latest* callbacks
  // without us having to tear down/recreate the connection every time
  // a caller passes a new function.
  const onModeChangedRef = useRef(onModeChanged);
  const onReflectorStatsRef = useRef(onReflectorStats);
  const onTestProgressRef = useRef(onTestProgress);

  useEffect(() => {
    onModeChangedRef.current = onModeChanged;
  }, [onModeChanged]);
  useEffect(() => {
    onReflectorStatsRef.current = onReflectorStats;
  }, [onReflectorStats]);
  useEffect(() => {
    onTestProgressRef.current = onTestProgress;
  }, [onTestProgress]);

  useEffect(() => {
    if (!enabled) {
      setStatus('idle');
      return undefined;
    }

    setStatus('connecting');
    const source = new EventSource(SSE_ENDPOINT, { withCredentials: true });

    source.onopen = (): void => {
      setStatus('connected');
    };

    source.onerror = (): void => {
      // EventSource auto-reconnects unless readyState === CLOSED.
      if (source.readyState === EventSource.CLOSED) {
        setStatus('closed');
      } else {
        setStatus('reconnecting');
      }
    };

    source.onmessage = (event: MessageEvent<string>): void => {
      let raw: unknown;
      try {
        raw = JSON.parse(event.data);
      } catch (err) {
        console.warn('SSE: malformed frame, dropped', { data: event.data, err });
        return;
      }

      const envelope = parseSseEnvelope(raw);
      if (!envelope) {
        console.warn('SSE: envelope failed schema, dropped', { data: raw });
        return;
      }

      switch (envelope.type) {
        case 'mode_changed': {
          const parsed = parseSseModeChanged(envelope);
          if (parsed && onModeChangedRef.current) {
            onModeChangedRef.current(parsed);
          }
          break;
        }
        case 'reflector_stats': {
          const parsed = parseSseReflectorStats(envelope);
          if (parsed && onReflectorStatsRef.current) {
            onReflectorStatsRef.current(parsed);
          }
          break;
        }
        case 'test_progress': {
          const parsed = parseSseTestProgress(envelope);
          if (parsed && onTestProgressRef.current) {
            onTestProgressRef.current(parsed);
          }
          break;
        }
        default:
          // Unknown frame type. Forward-compatible: just log and continue.
          console.warn('SSE: unknown frame type, dropped', { type: envelope.type });
      }
    };

    return (): void => {
      source.close();
      setStatus('closed');
    };
    // Intentionally only re-create the EventSource when `enabled` flips.
    // Callback changes flow through refs.
  }, [enabled]);

  return { status };
}

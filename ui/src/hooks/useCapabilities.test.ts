/**
 * useCapabilities tests.
 *
 * The hook gates ReflectorPage's platform-guard banner, so its failure
 * behaviour is the interesting part: any response the UI cannot trust must
 * leave the optimistic fallback in place rather than render a half-populated
 * capability object. These assert the value the caller actually receives, not
 * that fetch was called.
 */

import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { logWarn } from '../utils/logger';
import { type Capabilities, useCapabilities } from './useCapabilities';

vi.mock('../utils/logger', () => ({
  logWarn: vi.fn(),
}));

const FALLBACK: Capabilities = {
  reflector: { supported: true },
  testMaster: { supported: true },
};

function respondWith(body: unknown, init: { ok?: boolean; status?: number } = {}) {
  const ok = init.ok ?? true;
  return vi.fn().mockResolvedValue({
    ok,
    status: init.status ?? (ok ? 200 : 500),
    json: async () => body,
  });
}

describe('useCapabilities', () => {
  beforeEach(() => {
    vi.mocked(logWarn).mockClear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('starts optimistic so callers need no loading state', () => {
    vi.stubGlobal('fetch', respondWith(FALLBACK));
    const { result } = renderHook(() => useCapabilities());
    expect(result.current).toEqual(FALLBACK);
  });

  it('surfaces the reason for an unsupported reflector', async () => {
    vi.stubGlobal(
      'fetch',
      respondWith({
        reflector: { supported: false, reason: 'CGO + Linux required' },
        testMaster: { supported: true },
      }),
    );

    const { result } = renderHook(() => useCapabilities());

    await waitFor(() => {
      expect(result.current.reflector).toEqual({
        supported: false,
        reason: 'CGO + Linux required',
      });
    });
    expect(result.current.testMaster).toEqual({ supported: true });
  });

  it('requests the unauthenticated capabilities endpoint as JSON', async () => {
    const fetchMock = respondWith(FALLBACK);
    vi.stubGlobal('fetch', fetchMock);

    renderHook(() => useCapabilities());

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith('/api/v1/capabilities', {
        headers: { Accept: 'application/json' },
      });
    });
  });

  it('keeps the fallback when the endpoint is missing on an older build', async () => {
    vi.stubGlobal('fetch', respondWith(undefined, { ok: false, status: 404 }));

    const { result } = renderHook(() => useCapabilities());

    await waitFor(() => {
      expect(logWarn).toHaveBeenCalledWith('Failed to fetch /api/v1/capabilities', {
        component: 'useCapabilities',
        additionalData: { error: 'status 404' },
      });
    });
    expect(result.current).toEqual(FALLBACK);
  });

  it('keeps the fallback when the network request rejects', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('offline')));

    const { result } = renderHook(() => useCapabilities());

    await waitFor(() => {
      expect(logWarn).toHaveBeenCalledWith('Failed to fetch /api/v1/capabilities', {
        component: 'useCapabilities',
        additionalData: { error: 'offline' },
      });
    });
    expect(result.current).toEqual(FALLBACK);
  });

  it.each([
    ['null', null],
    ['a bare array', []],
    ['a partial payload missing testMaster', { reflector: { supported: true } }],
    [
      'a non-boolean supported flag',
      { reflector: { supported: 'yes' }, testMaster: { supported: true } },
    ],
    [
      'a non-string reason',
      { reflector: { supported: false, reason: 42 }, testMaster: { supported: true } },
    ],
  ])('rejects %s rather than adopting it', async (_label, body) => {
    vi.stubGlobal('fetch', respondWith(body));

    const { result } = renderHook(() => useCapabilities());

    await waitFor(() => {
      expect(logWarn).toHaveBeenCalledWith('Unexpected /api/v1/capabilities payload shape', {
        component: 'useCapabilities',
      });
    });
    expect(result.current).toEqual(FALLBACK);
  });

  it.each([
    [
      'a warning about an unusable payload',
      { ok: true, status: 200, json: async () => ({ nonsense: true }) },
      undefined,
    ],
    ['a warning about a failed request', undefined, new Error('offline')],
  ])('does not emit %s once the caller has unmounted', async (_label, resolved, rejected) => {
    let settle: () => void = () => {};
    const gate = new Promise<void>((resolve) => {
      settle = resolve;
    });
    vi.stubGlobal(
      'fetch',
      vi.fn().mockReturnValue(
        gate.then(() => {
          if (rejected) {
            throw rejected;
          }
          return resolved;
        }),
      ),
    );

    const { unmount } = renderHook(() => useCapabilities());
    unmount();
    settle();
    // Let the fetch chain settle completely before asserting on the absence
    // of a call — otherwise the assertion passes simply by running first.
    await gate;
    // A macrotask turn drains every microtask the fetch chain queues. Awaiting
    // a fixed number of microtasks instead lets the longer json() path finish
    // after the assertion, which would pass for the wrong reason.
    await new Promise((resolve) => {
      setTimeout(resolve, 0);
    });

    expect(logWarn).not.toHaveBeenCalled();
  });
});

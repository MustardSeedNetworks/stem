import { act, cleanup, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { HEALTHY_INTERVAL_MS, RETRY_INTERVAL_MS, useConnectionStatus } from './useConnectionStatus';

let fetchMock: ReturnType<typeof vi.spyOn>;

async function advance(ms: number): Promise<void> {
  const settled: Promise<void> = act(async () => {
    await vi.advanceTimersByTimeAsync(ms);
  });
  await settled;
}

beforeEach(() => {
  vi.useFakeTimers();
  fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response('{}', { status: 200 }));
});

afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe('useConnectionStatus', () => {
  it('starts checking, then reports connected from /__version', async () => {
    const { result } = renderHook(useConnectionStatus);
    expect(result.current).toBe('checking');
    await advance(0);
    expect(result.current).toBe('connected');
    expect(fetchMock.mock.calls[0]?.[0]).toBe('/__version');
  });

  it('reports disconnected when the daemon does not answer, and retries sooner', async () => {
    fetchMock.mockRejectedValueOnce(new TypeError('Failed to fetch'));
    const { result } = renderHook(useConnectionStatus);
    await advance(0);
    expect(result.current).toBe('disconnected');
    expect(fetchMock).toHaveBeenCalledTimes(1);

    await advance(RETRY_INTERVAL_MS);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(result.current).toBe('connected');
  });

  it('treats a non-ok answer as unreachable', async () => {
    fetchMock.mockResolvedValueOnce(new Response('bad gateway', { status: 502 }));
    const { result } = renderHook(useConnectionStatus);
    await advance(0);
    expect(result.current).toBe('disconnected');
  });

  it('re-confirms a healthy daemon on the slower interval', async () => {
    renderHook(useConnectionStatus);
    await advance(0);
    await advance(HEALTHY_INTERVAL_MS - 1);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    await advance(1);
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it('aborts the pending request and clears its timers on unmount', async () => {
    fetchMock.mockImplementation(
      (_url, init) =>
        new Promise((_resolve, reject) => {
          (init as RequestInit).signal?.addEventListener('abort', () =>
            reject(new DOMException('Aborted', 'AbortError')),
          );
        }),
    );
    const { unmount } = renderHook(useConnectionStatus);
    const init = fetchMock.mock.calls[0]?.[1] as RequestInit | undefined;
    const signal = init?.signal;
    expect(signal?.aborted).toBe(false);
    unmount();
    expect(signal?.aborted).toBe(true);
    await advance(HEALTHY_INTERVAL_MS * 2);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(vi.getTimerCount()).toBe(0);
  });
});

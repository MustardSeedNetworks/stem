/**
 * CSRF token cache — the two hazards that make an invalidation unreliable.
 *
 * `authFetch` drops this cache the moment a token refresh succeeds (#1315),
 * because the daemon keys CSRF tokens by sha256(bearer). Both tests below are
 * about what happens around that drop: several waiters wake from one refresh,
 * and a fetch may already be in flight when it happens.
 */

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { getCsrfToken, invalidateCsrfToken } from './csrf';

function deferred<T>(): { promise: Promise<T>; resolve: (value: T) => void } {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((r) => {
    resolve = r;
  });
  return { promise, resolve };
}

function tokenResponse(token: string): Response {
  return new Response(JSON.stringify({ token }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
}

beforeEach(() => {
  invalidateCsrfToken();
});

afterEach(() => {
  vi.restoreAllMocks();
  invalidateCsrfToken();
});

describe('getCsrfToken', () => {
  it('concurrent callers share one request', async () => {
    const spy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(tokenResponse('token-a'));
    const tokens = await Promise.all([getCsrfToken(), getCsrfToken(), getCsrfToken()]);
    expect(tokens).toEqual(['token-a', 'token-a', 'token-a']);
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it('a fetch that was in flight when the cache was dropped does not repopulate it', async () => {
    const stale = deferred<Response>();
    const spy = vi
      .spyOn(globalThis, 'fetch')
      .mockImplementationOnce(() => stale.promise)
      .mockResolvedValue(tokenResponse('token-new'));

    // Started before the refresh; it will answer with the pre-refresh token.
    const pending = getCsrfToken();
    invalidateCsrfToken();
    stale.resolve(tokenResponse('token-old'));
    await expect(pending).resolves.toBe('token-old');

    // The next caller must not be handed the token the invalidation discarded.
    await expect(getCsrfToken()).resolves.toBe('token-new');
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it('caches after a settled fetch so a second caller makes no request', async () => {
    const spy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(tokenResponse('token-a'));
    await expect(getCsrfToken()).resolves.toBe('token-a');
    await expect(getCsrfToken()).resolves.toBe('token-a');
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it('rejects a response that carries no token', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ token: '' }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    await expect(getCsrfToken()).rejects.toThrow('carried no token');
  });
});

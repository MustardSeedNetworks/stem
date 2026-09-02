import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { invalidateCsrfToken } from '../../../lib/csrf';
import { MFAError, mfaApi } from './mfaApi';

/**
 * These are the tests mfaApi never had.
 *
 * It was converged onto lib/csrf in #953 precisely because carrying its own
 * token helper meant the two could disagree — and they did: the shared helper
 * caches and retries on 403, the local copy did neither. Converging untested
 * code on the strength of a code read alone is how the next disagreement gets
 * introduced, so the behaviour that matters is pinned here.
 */

const CSRF_ENDPOINT = '/api/v1/auth/csrf-token';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('mfaApi CSRF handling', () => {
  beforeEach(() => {
    // The token is module-level cached, so one test's token must not leak into
    // the next and mask a missing fetch.
    invalidateCsrfToken();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('attaches the CSRF header to a mutating POST', async () => {
    const fetchMock = vi.fn(async (input: string | URL | Request) => {
      if (String(input) === CSRF_ENDPOINT) {
        return jsonResponse({ token: 'token-1' });
      }
      return jsonResponse({ success: true, totpEnabled: true });
    });
    vi.stubGlobal('fetch', fetchMock);

    await mfaApi.totpVerify('123456');

    const post = fetchMock.mock.calls.find(([input]) =>
      String(input).endsWith('/auth/totp/verify'),
    );
    expect(post, 'the verify request was never sent').toBeDefined();
    const init = post?.[1] as RequestInit;
    expect((init.headers as Record<string, string>)['X-Csrf-Token']).toBe('token-1');
    expect(init.credentials).toBe('include');
  });

  it('re-fetches the token and retries once when the daemon answers 403', async () => {
    // The daemon rotates the CSRF token on login, so a token minted before a
    // re-login is stale. This is the case the old local helper could not
    // recover from — enrolment simply failed.
    let issued = 0;
    let verifyAttempts = 0;
    const fetchMock = vi.fn(async (input: string | URL | Request) => {
      if (String(input) === CSRF_ENDPOINT) {
        issued += 1;
        return jsonResponse({ token: `token-${issued}` });
      }
      verifyAttempts += 1;
      if (verifyAttempts === 1) {
        return new Response('CSRF token invalid', { status: 403 });
      }
      return jsonResponse({ success: true, totpEnabled: true });
    });
    vi.stubGlobal('fetch', fetchMock);

    await expect(mfaApi.totpVerify('123456')).resolves.toEqual({
      success: true,
      totpEnabled: true,
    });

    expect(verifyAttempts, 'expected exactly one retry').toBe(2);
    expect(issued, 'expected a fresh token for the retry').toBe(2);

    const retry = fetchMock.mock.calls.filter(([input]) =>
      String(input).endsWith('/auth/totp/verify'),
    )[1];
    const init = retry?.[1] as RequestInit;
    expect((init.headers as Record<string, string>)['X-Csrf-Token']).toBe('token-2');
  });

  it('surfaces a persistent 403 as MFAError rather than retrying forever', async () => {
    let verifyAttempts = 0;
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: string | URL | Request) => {
        if (String(input) === CSRF_ENDPOINT) {
          return jsonResponse({ token: 'token-1' });
        }
        verifyAttempts += 1;
        return new Response('forbidden', { status: 403 });
      }),
    );

    await expect(mfaApi.totpVerify('123456')).rejects.toBeInstanceOf(MFAError);
    expect(verifyAttempts, 'one retry, then report it').toBe(2);
  });

  it('maps a non-403 failure to MFAError with the response status', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: string | URL | Request) => {
        if (String(input) === CSRF_ENDPOINT) {
          return jsonResponse({ token: 'token-1' });
        }
        return new Response('code already used', { status: 400 });
      }),
    );

    await expect(mfaApi.totpVerify('123456')).rejects.toMatchObject({
      name: 'MFAError',
      status: 400,
      message: 'code already used',
    });
  });

  it('does not send a CSRF header on the exempt login path', async () => {
    const fetchMock = vi.fn(async () => jsonResponse({ token: 'jwt', expiresAt: 1 }));
    vi.stubGlobal('fetch', fetchMock);

    await mfaApi.loginTotp('mfa-token', '123456');

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const init = fetchMock.mock.calls[0]?.[1] as RequestInit;
    expect((init.headers as Record<string, string>)['X-Csrf-Token']).toBeUndefined();
  });
});

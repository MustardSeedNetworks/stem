/**
 * Auth-store + authFetch tests — focused on the security-sensitive behaviors
 * surfaced in the cross-model design review: single-flight refresh, cache
 * purge on teardown, 401 retry semantics, 403 authz-vs-session branching, and
 * request-shape preservation (FormData, headers, signal, credentials).
 */

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const cancelQueries = vi.fn().mockResolvedValue(undefined);
const clear = vi.fn();
const { passkeyLoginMock } = vi.hoisted(() => ({
  passkeyLoginMock: vi.fn(),
}));

vi.mock('../lib/queryClient', () => ({
  getQueryClient: () => ({ cancelQueries, clear }),
}));

import { invalidateCsrfToken } from '../lib/csrf';

vi.mock('../lib/webauthn', () => ({
  loginWithPasskey: passkeyLoginMock,
}));

import * as http from '../utils/http';
import { authFetch, useAuthStore } from './auth-store';

const AUTH_FLAG_KEY = 'stem-authenticated';

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function textResponse(body: string, status: number): Response {
  return new Response(body, { status, headers: { 'Content-Type': 'text/plain' } });
}

/**
 * Mutating requests now fetch a CSRF token first, so a mock that answers every
 * URL with the case under test would answer the token endpoint with it too.
 * This answers the token endpoint and delegates the rest.
 */
function mockFetchWithCsrf(handler: (url: string) => Response): ReturnType<typeof vi.spyOn> {
  return vi.spyOn(globalThis, 'fetch').mockImplementation((input) => {
    const url = String(input);
    if (url.includes('/auth/csrf-token')) {
      return Promise.resolve(jsonResponse({ token: 'csrf-token-1' }));
    }
    return Promise.resolve(handler(url));
  });
}

beforeEach(() => {
  window.localStorage.clear();
  invalidateCsrfToken();
  cancelQueries.mockClear();
  clear.mockClear();
  passkeyLoginMock.mockReset();
  useAuthStore.setState({
    isAuthenticated: false,
    loginLoading: false,
    loginError: null,
    mfaPending: null,
    setupStatus: null,
    setupChecked: false,
    recoveryStatus: null,
    showRecoveryForm: false,
  });
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('auth-store: login / MFA', () => {
  it('marks authenticated + sets the localStorage flag on success', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(jsonResponse({ token: 'abc' }));
    const result = await useAuthStore.getState().login('admin', 'pw');
    expect(result).toEqual({ status: 'ok' });
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(window.localStorage.getItem(AUTH_FLAG_KEY)).toBe('true');
  });

  it('returns mfa-required and holds the mfaToken without authenticating', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      jsonResponse({ mfaRequired: true, mfaToken: 'tok', factor: 'totp' }),
    );
    const result = await useAuthStore.getState().login('admin', 'pw');
    expect(result).toEqual({ status: 'mfa', mfaToken: 'tok', factor: 'totp' });
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().mfaPending).toEqual({ mfaToken: 'tok', factor: 'totp' });
  });

  it('surfaces the server message on login failure', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(textResponse('bad creds', 401));
    const result = await useAuthStore.getState().login('admin', 'pw');
    expect(result).toEqual({ status: 'error', message: 'bad creds' });
    expect(useAuthStore.getState().loginError).toBe('bad creds');
  });

  it('verifyMfa authenticates and clears the pending challenge', async () => {
    useAuthStore.setState({ mfaPending: { mfaToken: 'tok', factor: 'totp' } });
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(jsonResponse({ token: 'abc' }));
    const result = await useAuthStore.getState().verifyMfa('123456');
    expect(result).toEqual({ status: 'ok' });
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(useAuthStore.getState().mfaPending).toBeNull();
  });

  it('gives login and MFA verification a deadline', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(jsonResponse({ token: 'a' }));
    await useAuthStore.getState().login('admin', 'pw');
    expect((fetchMock.mock.calls[0][1] as RequestInit).signal).toBeInstanceOf(AbortSignal);

    useAuthStore.setState({ mfaPending: { mfaToken: 'tok', factor: 'totp' } });
    await useAuthStore.getState().verifyMfa('123456');
    expect((fetchMock.mock.calls[1][1] as RequestInit).signal).toBeInstanceOf(AbortSignal);
  });

  it('distinguishes a server that did not answer from one it could not reach', async () => {
    // The deadline is driven for real rather than simulated by hand-building
    // an error named TimeoutError. That shape is not what either engine
    // produces for an aborted fetch -- Chromium raises TimeoutError, WebKit
    // AbortError -- so a test written that way agrees with an implementation
    // that is wrong on Safari. Here the injected signal is genuinely aborted
    // and the code reads it, which is what production does.
    const controller = new AbortController();
    vi.spyOn(http, 'requestDeadline').mockReturnValue(controller.signal);
    vi.spyOn(globalThis, 'fetch').mockImplementation(() => {
      controller.abort();
      return Promise.reject(new DOMException('Fetch is aborted', 'AbortError'));
    });

    const result = await useAuthStore.getState().login('admin', 'pw');

    expect(result).toEqual({
      status: 'error',
      message: 'The authentication server did not respond. Check the connection and try again.',
    });
    // Without a deadline this never settles: loginLoading stays true and the
    // form stays disabled with no error and no retry.
    expect(useAuthStore.getState().loginLoading).toBe(false);
    expect(useAuthStore.getState().loginError).toMatch(/did not respond/);
  });

  it('still reports an unreachable server distinctly', async () => {
    vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));

    const result = await useAuthStore.getState().login('admin', 'pw');

    expect(result).toEqual({
      status: 'error',
      message: 'Unable to reach authentication server.',
    });
  });

  it('cancelMfa drops the pending challenge and clears the error', () => {
    useAuthStore.setState({
      mfaPending: { mfaToken: 'tok', factor: 'totp' },
      loginError: 'previous error',
    });
    useAuthStore.getState().cancelMfa();
    expect(useAuthStore.getState().mfaPending).toBeNull();
    expect(useAuthStore.getState().loginError).toBeNull();
  });

  it('marks the session authenticated after passkey sign-in', async () => {
    passkeyLoginMock.mockResolvedValue({ token: 'abc', expiresAt: 1 });

    const result = await useAuthStore.getState().passkeyLogin();

    expect(result).toEqual({ status: 'ok' });
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
    expect(window.localStorage.getItem(AUTH_FLAG_KEY)).toBe('true');
  });

  it('surfaces a rejected passkey ceremony', async () => {
    passkeyLoginMock.mockRejectedValue(new Error('Passkey sign-in was cancelled.'));

    const result = await useAuthStore.getState().passkeyLogin();

    expect(result).toEqual({ status: 'error', message: 'Passkey sign-in was cancelled.' });
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().loginError).toBe('Passkey sign-in was cancelled.');
  });
});

describe('auth-store: logout', () => {
  it('sends the CSRF header, so the server session is actually ended', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    const fetchSpy = mockFetchWithCsrf(() => jsonResponse({ status: 'ok' }));

    await useAuthStore.getState().logout();

    const call = fetchSpy.mock.calls.find((c) => String(c[0]).includes('/auth/logout'));
    expect(call, 'no request reached /auth/logout').toBeDefined();
    // fetchWithCsrf builds a plain header record, not a Headers instance.
    const init = (call as unknown as [string, RequestInit])[1];
    expect((init.headers as Record<string, string>)['X-Csrf-Token']).toBe('csrf-token-1');
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });
});

describe('auth-store: teardown purges the query cache', () => {
  it('logout cancels queries then clears + drops the flag', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    window.localStorage.setItem(AUTH_FLAG_KEY, 'true');
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(null, { status: 204 }));
    await useAuthStore.getState().logout();
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(window.localStorage.getItem(AUTH_FLAG_KEY)).toBeNull();
    await Promise.resolve();
    expect(cancelQueries).toHaveBeenCalled();
    expect(clear).toHaveBeenCalled();
  });

  it('expireSession tears down with a message', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    useAuthStore.getState().expireSession('Session expired. Please sign in again.');
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().loginError).toBe('Session expired. Please sign in again.');
    await Promise.resolve();
    expect(cancelQueries).toHaveBeenCalled();
    expect(clear).toHaveBeenCalled();
  });
});

describe('authFetch', () => {
  it('throws when not authenticated', async () => {
    await expect(authFetch('/api/v1/stats')).rejects.toThrow('Not authenticated');
  });

  it('401 → refresh succeeds → returns the retried response', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    const fetchSpy = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(null, { status: 401 })) // initial
      .mockResolvedValueOnce(new Response(null, { status: 200 })) // refresh
      .mockResolvedValueOnce(jsonResponse({ ok: true })); // retry
    const res = await authFetch('/api/v1/stats');
    expect(res.status).toBe(200);
    expect(fetchSpy).toHaveBeenCalledTimes(3);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it('401 → refresh fails → expires session and throws Unauthorized', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(null, { status: 401 })) // initial
      .mockResolvedValueOnce(new Response(null, { status: 401 })); // refresh fails
    await expect(authFetch('/api/v1/stats')).rejects.toThrow('Unauthorized');
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it('concurrent 401s share a single refresh (single-flight)', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockImplementation((input) => {
      const url = String(input);
      if (url.includes('/auth/refresh')) {
        return Promise.resolve(new Response(null, { status: 200 }));
      }
      // First call per URL is 401, retries succeed — but we only care about refresh count.
      return Promise.resolve(new Response(null, { status: 401 }));
    });
    // Two concurrent requests both hit 401 and trigger refresh.
    await Promise.allSettled([authFetch('/api/v1/a'), authFetch('/api/v1/b')]);
    const refreshCalls = fetchSpy.mock.calls.filter((c) => String(c[0]).includes('/auth/refresh'));
    expect(refreshCalls.length).toBe(1);
  });

  it('403 PERMISSION_DENIED → returns the response WITHOUT expiring', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    mockFetchWithCsrf(() =>
      jsonResponse({ error: 'Insufficient permissions', code: 'PERMISSION_DENIED' }, 403),
    );
    const res = await authFetch('/api/v1/config/import', { method: 'POST', body: '{}' });
    expect(res.status).toBe(403);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it('403 plain-text (CSRF/unknown) → expires session and throws Forbidden', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    mockFetchWithCsrf(() => textResponse('Invalid CSRF token', 403));
    await expect(authFetch('/api/v1/alerts', { method: 'PUT', body: '{}' })).rejects.toThrow(
      'Forbidden',
    );
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
  });

  it('a retried request that returns 5xx is returned to the caller (not an auth failure)', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    vi.spyOn(globalThis, 'fetch')
      .mockResolvedValueOnce(new Response(null, { status: 401 })) // initial
      .mockResolvedValueOnce(new Response(null, { status: 200 })) // refresh
      .mockResolvedValueOnce(new Response(null, { status: 500 })); // retry
    const res = await authFetch('/api/v1/stats');
    expect(res.status).toBe(500);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  // Every mutating route is behind the daemon's per-session CSRF manager and
  // the exempt list holds only pre-session endpoints, so an authFetch POST
  // without the header was answered 403 with a plain-text body — which the
  // branch above reads as a dead session. Starting and stopping a test both
  // go through here, so both signed the operator out instead of running
  // (#1080).
  it('attaches the CSRF header to a mutating request', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    const fetchSpy = mockFetchWithCsrf(() => jsonResponse({ status: 'stopped' }));

    await authFetch('/api/v1/test/stop', { method: 'POST' });

    const stop = fetchSpy.mock.calls.find((c) => String(c[0]).includes('/test/stop'));
    expect(stop, 'no request reached /test/stop').toBeDefined();
    const headers = (stop as unknown as [string, RequestInit])[1].headers as Headers;
    expect(headers.get('X-Csrf-Token')).toBe('csrf-token-1');
  });

  it('does not fetch or attach a token for a safe request', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    const fetchSpy = mockFetchWithCsrf(() => jsonResponse({ ok: true }));

    await authFetch('/api/v1/stats');

    expect(fetchSpy.mock.calls.some((c) => String(c[0]).includes('/auth/csrf-token'))).toBe(false);
    const stats = fetchSpy.mock.calls.find((c) => String(c[0]).includes('/stats'));
    expect(stats, 'no request reached /stats').toBeDefined();
    const headers = (stats as unknown as [string, RequestInit])[1].headers as Headers;
    expect(headers.has('X-Csrf-Token')).toBe(false);
  });

  it('a mutating 403 retries once with a fresh token before giving up', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    let tokenIssue = 0;
    const seen: string[] = [];
    vi.spyOn(globalThis, 'fetch').mockImplementation((input, init) => {
      const url = String(input);
      if (url.includes('/auth/csrf-token')) {
        tokenIssue += 1;
        return Promise.resolve(jsonResponse({ token: `csrf-token-${tokenIssue}` }));
      }
      const sent = ((init as RequestInit).headers as Headers).get('X-Csrf-Token') ?? '';
      seen.push(sent);
      // The daemon rotates the token on login: the first (stale) one is
      // refused, the re-fetched one is accepted.
      return Promise.resolve(
        sent === 'csrf-token-2'
          ? jsonResponse({ status: 'stopped' })
          : textResponse('CSRF token invalid', 403),
      );
    });

    const res = await authFetch('/api/v1/test/stop', { method: 'POST' });

    expect(res.status).toBe(200);
    expect(seen).toEqual(['csrf-token-1', 'csrf-token-2']);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it('preserves FormData (no Content-Type override) and forwards the abort signal + credentials', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    const fetchSpy = mockFetchWithCsrf(() => jsonResponse({ ok: true }));
    const fd = new FormData();
    fd.append('f', 'v');
    const controller = new AbortController();
    await authFetch('/api/v1/upload', { method: 'POST', body: fd, signal: controller.signal });
    const upload = fetchSpy.mock.calls.find((c) => String(c[0]).includes('/upload'));
    expect(upload, 'no request reached /upload').toBeDefined();
    const init = (upload as unknown as [string, RequestInit])[1] as RequestInit & {
      headers: Headers;
    };
    expect(init.credentials).toBe('include');
    expect(init.signal).toBe(controller.signal);
    expect((init.headers as Headers).has('Content-Type')).toBe(false);
  });
});

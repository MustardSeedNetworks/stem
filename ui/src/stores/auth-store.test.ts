/**
 * Auth-store + authFetch tests — focused on the security-sensitive behaviors
 * surfaced in the cross-model design review: single-flight refresh, cache
 * purge on teardown, 401 retry semantics, 403 authz-vs-session branching, and
 * request-shape preservation (FormData, headers, signal, credentials).
 */

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const cancelQueries = vi.fn().mockResolvedValue(undefined);
const clear = vi.fn();

vi.mock('../lib/queryClient', () => ({
  getQueryClient: () => ({ cancelQueries, clear }),
}));

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

beforeEach(() => {
  window.localStorage.clear();
  cancelQueries.mockClear();
  clear.mockClear();
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
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      jsonResponse({ error: 'Insufficient permissions', code: 'PERMISSION_DENIED' }, 403),
    );
    const res = await authFetch('/api/v1/config/import', { method: 'POST', body: '{}' });
    expect(res.status).toBe(403);
    expect(useAuthStore.getState().isAuthenticated).toBe(true);
  });

  it('403 plain-text (CSRF/unknown) → expires session and throws Forbidden', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(textResponse('Invalid CSRF token', 403));
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

  it('preserves FormData (no Content-Type override) and forwards the abort signal + credentials', async () => {
    useAuthStore.setState({ isAuthenticated: true });
    const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(jsonResponse({ ok: true }));
    const fd = new FormData();
    fd.append('f', 'v');
    const controller = new AbortController();
    await authFetch('/api/v1/upload', { method: 'POST', body: fd, signal: controller.signal });
    const init = fetchSpy.mock.calls[0]?.[1] as RequestInit & { headers: Headers };
    expect(init.credentials).toBe('include');
    expect(init.signal).toBe(controller.signal);
    expect((init.headers as Headers).has('Content-Type')).toBe(false);
  });
});

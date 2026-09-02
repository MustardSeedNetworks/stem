/**
 * RoleContext tests.
 *
 * The role decides what a stem instance actually is — a passive Reflector or an
 * active Test Master — and switching it cancels any in-progress test. So the
 * behaviours worth pinning are the ones where getting it wrong leaves the UI
 * claiming a role the backend does not have: a failed switch must not move
 * local state, the server's echoed mode wins over the requested one, and a
 * slow first response must not overwrite a newer second.
 */

import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { invalidateCsrfToken } from '../lib/csrf';
import { ROLE_ENDPOINT, ROLE_STORAGE_KEY, RoleProvider, useRole } from './RoleContext';

function wrapper({ children }: { children: ReactNode }) {
  return <RoleProvider>{children}</RoleProvider>;
}

function renderRole() {
  return renderHook(() => useRole(), { wrapper });
}

/**
 * Let queued promise callbacks run and their state updates land.
 *
 * Needed only where the assertion is about something NOT happening: waitFor can
 * prove a change arrived, but proving a stale result was *ignored* means giving
 * its handler a chance to wrongly apply first. A bare setTimeout is not enough
 * — an update outside act() never reaches the hook result, so the test would
 * pass whether or not the guard exists. waitFor runs its callback inside act,
 * which is what makes the flush observable.
 */
async function flushPending(result: { current: { isSwitchingRole: boolean } }): Promise<void> {
  await waitFor(() => {
    expect(result.current.isSwitchingRole).toBe(false);
  });
}

/**
 * A successful POST /api/v1/mode reply, matching the Go ModeUpdateResponse:
 * status, the mode the server settled on, and the one it came from. All three
 * are required by the schema — a body with only `mode` is rejected as an
 * unexpected shape, which is what the malformed-body case below relies on.
 */
function modeResponse(mode: string, previous = 'reflector'): Response {
  return new Response(JSON.stringify({ status: 'updated', mode, previous }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
}

let fetchMock: ReturnType<typeof vi.fn>;

/**
 * The mode POST is preceded by a CSRF token fetch, because /api/v1/mode is a
 * mutating route behind the CSRF manager. These helpers keep the assertions
 * about the mode call rather than about call ordering — before the header was
 * added, every one of these tests read `fetchMock.mock.calls[0]` and would
 * have passed just as happily on a request the daemon answers with 403.
 */
const CSRF_ENDPOINT = '/api/v1/auth/csrf-token';

function csrfResponse(token = 'test-csrf-token'): Response {
  return new Response(JSON.stringify({ token }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
}

/** The arguments of the POST to /api/v1/mode, whichever call index it landed on. */
function modeCall(): [string, RequestInit] {
  const call = fetchMock.mock.calls.find(([url]) => url === ROLE_ENDPOINT);
  if (call === undefined) {
    throw new Error(
      `no request to ${ROLE_ENDPOINT}; saw ${JSON.stringify(fetchMock.mock.calls.map(([u]) => u))}`,
    );
  }
  return call as [string, RequestInit];
}

/** Answers the CSRF endpoint, then hands out `replies` to the mode calls in order. */
function respondWithSequence(...replies: Array<Response | Promise<Response>>): void {
  const queue = [...replies];
  fetchMock.mockImplementation(async (url: string) => {
    if (url === CSRF_ENDPOINT) {
      return csrfResponse();
    }
    const next = queue.shift();
    if (next === undefined) {
      throw new Error('mode call made with no reply queued');
    }
    return next;
  });
}

/** Answers the CSRF endpoint, and everything else with `modeReply`. */
function respondWith(modeReply: Response | (() => Promise<Response>)): void {
  fetchMock.mockImplementation(async (url: string) => {
    if (url === CSRF_ENDPOINT) {
      return csrfResponse();
    }
    return typeof modeReply === 'function' ? modeReply() : modeReply;
  });
}

beforeEach(() => {
  window.localStorage.clear();
  // The token is cached for the session; without this a test inherits the
  // previous test's token and never exercises the fetch.
  invalidateCsrfToken();
  fetchMock = vi.fn();
  vi.stubGlobal('fetch', fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('initial role', () => {
  it('defaults to reflector with nothing persisted', () => {
    expect(renderRole().result.current.role).toBe('reflector');
  });

  it('restores a persisted role', () => {
    window.localStorage.setItem(ROLE_STORAGE_KEY, 'test_master');

    expect(renderRole().result.current.role).toBe('test_master');
  });

  it('falls back to reflector when the persisted value is not a role', () => {
    // A value written by an older build, or edited by hand, must not put the
    // app into an unrepresentable state.
    window.localStorage.setItem(ROLE_STORAGE_KEY, 'superuser');

    expect(renderRole().result.current.role).toBe('reflector');
  });

  it('persists the role it started with', () => {
    renderRole();

    expect(window.localStorage.getItem(ROLE_STORAGE_KEY)).toBe('reflector');
  });
});

describe('switching succeeds', () => {
  it('posts the requested mode to the endpoint', async () => {
    respondWith(modeResponse('test_master'));
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));
    await waitFor(() => expect(result.current.isSwitchingRole).toBe(false));

    const [url, init] = modeCall();
    expect(url).toBe(ROLE_ENDPOINT);
    expect(init.method).toBe('POST');
    expect(init.credentials).toBe('include');
    expect(JSON.parse(init.body as string)).toEqual({ mode: 'test_master' });
  });

  it('sends the CSRF token, without which the daemon answers 403', async () => {
    // The defect this guards. /api/v1/mode is a mutating route behind the CSRF
    // manager, and this POST carried no X-Csrf-Token: every role switch from
    // the header chip or the RoleGuard banner came back
    //   403  CSRF validation failed  error="CSRF token missing"
    // in a real browser. Nothing caught it because the only E2E covering the
    // path mocks the endpoint with route.fulfill, and a mock accepts a request
    // whether or not it carries the header.
    respondWith(modeResponse('test_master'));
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));
    await waitFor(() => expect(result.current.isSwitchingRole).toBe(false));

    const [, init] = modeCall();
    const headers = init.headers as Record<string, string>;
    expect(headers['X-Csrf-Token']).toBe('test-csrf-token');

    // And the token came from the endpoint rather than being invented.
    expect(fetchMock.mock.calls.map(([url]) => url)).toContain(CSRF_ENDPOINT);
  });

  it('adopts the mode the server echoed, not the one requested', async () => {
    // The server is the authority: if it normalises the value, local state has
    // to follow it or the UI and the daemon disagree about what this stem is.
    // Start from test_master and have the server answer reflector, so the
    // assertion cannot pass by the role simply never changing.
    window.localStorage.setItem(ROLE_STORAGE_KEY, 'test_master');
    respondWith(modeResponse('reflector', 'test_master'));
    const { result } = renderRole();
    expect(result.current.role).toBe('test_master');

    act(() => result.current.setRole('test_master'));

    await waitFor(() => expect(result.current.role).toBe('reflector'));
    expect(result.current.roleSwitchError).toBeNull();
  });

  it('persists the new role', async () => {
    respondWith(modeResponse('test_master'));
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));

    await waitFor(() => expect(window.localStorage.getItem(ROLE_STORAGE_KEY)).toBe('test_master'));
  });

  it('reports the switch as in flight and then done', async () => {
    let settle: (r: Response) => void = () => undefined;
    fetchMock.mockReturnValue(
      new Promise<Response>((resolve) => {
        settle = resolve;
      }),
    );
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));
    expect(result.current.isSwitchingRole).toBe(true);

    act(() => {
      settle(modeResponse('test_master'));
    });

    await waitFor(() => expect(result.current.isSwitchingRole).toBe(false));
  });
});

describe('switching fails', () => {
  it.each([
    [
      'a JSON error body with message',
      new Response(JSON.stringify({ message: 'reflector is busy' }), {
        status: 409,
        headers: { 'Content-Type': 'application/json' },
      }),
      'reflector is busy',
    ],
    [
      'a JSON error body with only error',
      new Response(JSON.stringify({ error: 'mode locked' }), {
        status: 409,
        headers: { 'Content-Type': 'application/json' },
      }),
      'mode locked',
    ],
    [
      'a non-JSON error body',
      new Response('gateway down', { status: 502 }),
      'Role switch failed (HTTP 502)',
    ],
  ])('surfaces %s', async (_label, response, expected) => {
    respondWith(response);
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));

    await waitFor(() => expect(result.current.roleSwitchError).toBe(expected));
  });

  it('leaves the role unchanged when the switch is refused', async () => {
    // The important half: a stem that failed to become Test Master must not
    // render as one.
    respondWith(new Response('nope', { status: 500 }));
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));

    await waitFor(() => expect(result.current.roleSwitchError).not.toBeNull());
    expect(result.current.role).toBe('reflector');
    expect(window.localStorage.getItem(ROLE_STORAGE_KEY)).toBe('reflector');
  });

  it('surfaces a network failure', async () => {
    fetchMock.mockRejectedValue(new Error('connection refused'));
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));

    await waitFor(() =>
      expect(result.current.roleSwitchError).toBe('Role switch failed: connection refused'),
    );
    expect(result.current.role).toBe('reflector');
  });

  it('rejects a 200 whose body is not the expected shape', async () => {
    // A proxy returning an HTML success page must not be read as a role change.
    respondWith(
      new Response(JSON.stringify({ status: 'fine' }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));

    await waitFor(() =>
      expect(result.current.roleSwitchError).toBe('Role switch failed: unexpected server response'),
    );
    expect(result.current.role).toBe('reflector');
  });

  it('clears a previous error when a new switch starts', async () => {
    respondWithSequence(new Response('nope', { status: 500 }), modeResponse('test_master'));
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));
    await waitFor(() => expect(result.current.roleSwitchError).not.toBeNull());

    act(() => result.current.setRole('test_master'));

    expect(result.current.roleSwitchError).toBeNull();
  });

  it('clears the error on request', async () => {
    respondWith(new Response('nope', { status: 500 }));
    const { result } = renderRole();

    act(() => result.current.setRole('test_master'));
    await waitFor(() => expect(result.current.roleSwitchError).not.toBeNull());

    act(() => result.current.clearRoleSwitchError());

    expect(result.current.roleSwitchError).toBeNull();
  });
});

describe('overlapping switches', () => {
  it('ignores a slow first response that lands after a newer one', async () => {
    // The reason inflightToken exists. Clicking Reflector then Test Master and
    // having the first reply arrive last would otherwise leave the UI showing
    // Reflector while the daemon is a Test Master.
    let settleFirst: (r: Response) => void = () => undefined;
    let settleSecond: (r: Response) => void = () => undefined;

    respondWithSequence(
      new Promise<Response>((resolve) => {
        settleFirst = resolve;
      }),
      new Promise<Response>((resolve) => {
        settleSecond = resolve;
      }),
    );

    const { result } = renderRole();

    act(() => result.current.setRole('reflector'));
    act(() => result.current.setRole('test_master'));

    act(() => {
      settleSecond(modeResponse('test_master'));
    });
    await waitFor(() => expect(result.current.role).toBe('test_master'));

    act(() => {
      settleFirst(modeResponse('reflector'));
    });
    await flushPending(result);

    expect(result.current.role).toBe('test_master');
    expect(result.current.isSwitchingRole).toBe(false);
  });

  it('ignores a stale failure that lands after a newer success', async () => {
    let settleFirst: (r: Response) => void = () => undefined;
    let settleSecond: (r: Response) => void = () => undefined;

    respondWithSequence(
      new Promise<Response>((resolve) => {
        settleFirst = resolve;
      }),
      new Promise<Response>((resolve) => {
        settleSecond = resolve;
      }),
    );

    const { result } = renderRole();

    act(() => result.current.setRole('reflector'));
    act(() => result.current.setRole('test_master'));

    act(() => {
      settleSecond(modeResponse('test_master'));
    });
    await waitFor(() => expect(result.current.role).toBe('test_master'));

    act(() => {
      settleFirst(new Response('too late', { status: 500 }));
    });
    await flushPending(result);

    // A stale error banner on a switch that actually succeeded is exactly as
    // misleading as a stale role.
    expect(result.current.roleSwitchError).toBeNull();
    expect(result.current.role).toBe('test_master');
  });
});

describe('useRole outside a provider', () => {
  it('throws rather than handing back a silent default', () => {
    // A component rendered outside RoleProvider would otherwise read the
    // default role and appear to work.
    expect(() => renderHook(() => useRole())).toThrow('useRole must be used inside <RoleProvider>');
  });
});

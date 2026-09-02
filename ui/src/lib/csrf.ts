/**
 * CSRF token access for state-changing requests.
 *
 * Every mutating route on the daemon is behind the per-session CSRF manager,
 * so a POST/PUT/DELETE without `X-Csrf-Token` is answered with 403 and
 * `error="CSRF token missing"`. Before this module the header was set in
 * exactly one file — the MFA API — and `RoleContext.requestModeSwitch` omitted
 * it, so switching role from the header chip or the RoleGuard banner failed
 * 403 every time in a real browser. Nothing caught it because the only E2E
 * covering that path mocks `/api/v1/mode` with `route.fulfill`, and a mock
 * accepts a request whether or not it carries the header.
 *
 * The token is cached for the session and re-fetched once on a 403, because
 * the daemon rotates it on login: a token minted before a re-login is stale,
 * and one retry is the difference between "your session rotated" and "the
 * button does nothing".
 */

const CSRF_ENDPOINT = '/api/v1/auth/csrf-token';

let cached: string | null = null;

/** Drops the cached token so the next call re-fetches. */
export function invalidateCsrfToken(): void {
  cached = null;
}

/** Returns the session's CSRF token, fetching and caching it on first use. */
export async function getCsrfToken(): Promise<string> {
  if (cached !== null) {
    return cached;
  }
  const response = await fetch(CSRF_ENDPOINT, { credentials: 'include' });
  if (!response.ok) {
    throw new Error(`Failed to fetch CSRF token: HTTP ${response.status}`);
  }
  const body = (await response.json()) as { token?: unknown };
  if (typeof body.token !== 'string' || body.token === '') {
    throw new Error('CSRF token response carried no token');
  }
  cached = body.token;
  return cached;
}

/**
 * Performs a mutating fetch with the CSRF header attached, retrying once with
 * a fresh token if the daemon rejects the first attempt with 403.
 */
export async function fetchWithCsrf(input: string, init: RequestInit = {}): Promise<Response> {
  const send = async (token: string): Promise<Response> =>
    fetch(input, {
      ...init,
      credentials: 'include',
      headers: { ...(init.headers ?? {}), 'X-Csrf-Token': token },
    });

  const response = await send(await getCsrfToken());
  if (response.status !== 403) {
    return response;
  }
  // Stale token — the daemon rotates on login. One retry, then give up so a
  // genuinely forbidden request is reported as forbidden rather than looping.
  invalidateCsrfToken();
  return send(await getCsrfToken());
}

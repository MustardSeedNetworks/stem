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
 *
 * The cache is keyed to the session implicitly: the daemon keys tokens by
 * sha256(bearer), so a token refresh mints a new key under which no token
 * exists. `refreshAccessToken` therefore invalidates this cache before its
 * waiters retry (#1315).
 *
 * `fetchWithCsrf` has exactly one caller left, and deliberately so: sign-out.
 * It carries the CSRF header and nothing else — no 401 refresh — because
 * `/api/v1/auth/logout` is registered without `auth: true` and because
 * refreshing a session in order to end it is not a thing to do. Every mutation
 * on an `auth: true` route belongs on `authFetch`, which refreshes and retries;
 * routing one here is the #1318 defect. The remaining pre-session mutations
 * (login, refresh, setup, recovery, passkey sign-in) are on bare `fetch`,
 * which is what the CSRF middleware's exempt list expects.
 */

const CSRF_ENDPOINT = '/api/v1/auth/csrf-token';

let cached: string | null = null;
// Bumped by every invalidation. A fetch that started before an invalidation
// carries the session's OLD token, so it must not write it into the cache
// afterwards — that would hand the next mutation the token the invalidation
// just discarded. Seed hit the same client-side case (MustardSeedNetworks/seed#2660).
let generation = 0;
// Single-flight: the daemon keys tokens by session, so concurrent callers want
// the same value, and one request is enough to get it.
let inFlight: Promise<string> | null = null;

/** Drops the cached token so the next call re-fetches. */
export function invalidateCsrfToken(): void {
  cached = null;
  inFlight = null;
  generation += 1;
}

async function fetchCsrfToken(): Promise<string> {
  const startedAt = generation;
  const response = await fetch(CSRF_ENDPOINT, { credentials: 'include' });
  if (!response.ok) {
    throw new Error(`Failed to fetch CSRF token: HTTP ${response.status}`);
  }
  const body = (await response.json()) as { token?: unknown };
  if (typeof body.token !== 'string' || body.token === '') {
    throw new Error('CSRF token response carried no token');
  }
  if (generation === startedAt) {
    cached = body.token;
  }
  return body.token;
}

/** Returns the session's CSRF token, fetching and caching it on first use. */
export async function getCsrfToken(): Promise<string> {
  if (cached !== null) {
    return cached;
  }
  if (inFlight === null) {
    const started = generation;
    inFlight = fetchCsrfToken().finally(() => {
      if (generation === started) {
        inFlight = null;
      }
    });
  }
  return inFlight;
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

import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * #1315 — a mutation retried after a token refresh must carry a FRESH CSRF token.
 *
 * The daemon keys CSRF tokens by sha256(bearer) (internal/auth/csrf.go), so a
 * successful `/auth/refresh` mints an access token whose session key holds no
 * token. `authFetch` retried with the cached one, the daemon answered 403, and
 * `authFetch` returns a non-401 retry verbatim — so the 403 went straight to
 * the caller and the button failed with nothing useful on screen. The 403
 * branch further down never runs on this path, which is why the defect had no
 * self-healing retry to hide behind.
 *
 * NOTHING on the path under test is stubbed. A `route.fulfill` mock accepts a
 * request whatever header it carries, which is exactly how the defect survived
 * `auth-store.test.ts`'s "401 → refresh succeeds" case — green throughout.
 *
 * The vehicle is `POST /api/v1/test/start` from the Benchmark page: the one
 * mutating route that is both behind `auth: true` (so an absent access token
 * really is a 401) and issued through `authFetch`. Its verdict is the daemon's
 * own — a CGO-less build has no dataplane and will refuse the run — and that
 * is fine: this spec is about the request reaching the handler with a valid
 * token, not about the run succeeding. `/api/v1/license/*` cannot serve here
 * (no `auth: true`, so it never 401s) and the role switch goes through
 * `fetchWithCsrf`, which has no refresh path at all.
 */

const ACCESS_COOKIE = 'stem_access';
const START_ROUTE = '/api/v1/test/start';

test.describe('mutation after token refresh', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    // Test Master is the role that may run tests; as a reflector the module
    // page only warns and never offers a start control.
    await useRole(page, 'test_master');
  });

  test('retries with a new CSRF token and reaches the handler', async ({ page, context }) => {
    const starts: { status: number; csrf: string }[] = [];
    page.on('response', (response) => {
      if (response.url().endsWith(START_ROUTE)) {
        starts.push({
          status: response.status(),
          csrf: response.request().headers()['x-csrf-token'] ?? '',
        });
      }
    });

    await page.goto('/tests/benchmark');
    await expect(page.getByTestId('rfc2544-config-form')).toBeVisible({ timeout: 10000 });
    const iface = page.getByTestId('interface-select');
    const value = await iface.locator('option').nth(1).getAttribute('value');
    expect(value, 'daemon reported no interfaces to select').toBeTruthy();
    await iface.selectOption(value as string);
    await page.getByTestId('peer-input').fill('192.0.2.10');

    const start = page.getByTestId('start-test-button');
    await expect(start).toBeEnabled();

    // First start primes the CSRF cache with a token minted for the CURRENT
    // bearer — the precondition the defect needs.
    await start.click();
    await expect.poll(() => starts.length, { timeout: 10000 }).toBe(1);
    const primedToken = starts[0].csrf;
    expect(primedToken, 'the first start carried no CSRF token').not.toBe('');

    // Expire the access token only; the refresh cookie stays, so the session
    // recovers through a refresh. Which request takes the 401 — this mutation
    // or the dashboard's one-second stats poll — is a race, and the assertions
    // below deliberately do not depend on winning it. Either way the defect
    // shows the same signature: a start answered 403 for reusing the token
    // minted under the old bearer. Pinning the exact
    // 401 -> refresh -> retry ordering is the unit test's job
    // (auth-store.test.ts), which controls every response.
    const kept = (await context.cookies()).filter((c) => c.name !== ACCESS_COOKIE);
    await context.clearCookies();
    await context.addCookies(kept);

    const refreshed = page.waitForResponse(
      (r) => r.url().endsWith('/api/v1/auth/refresh') && r.status() === 200,
    );
    await expect(start).toBeEnabled();
    await start.click();
    // A refresh really happened — otherwise the cookie never expired and the
    // rest of this test would pass without exercising anything.
    await refreshed;
    await expect
      .poll(() => starts.length >= 2 && starts[starts.length - 1].status !== 401, {
        timeout: 10000,
      })
      .toBe(true);

    // 403 is the defect, wherever it lands: a start sent with the token minted
    // under the old bearer, for which the daemon holds nothing.
    expect(starts.map((s) => s.status)).not.toContain(403);
    const settled = starts[starts.length - 1];
    expect(settled.csrf).not.toBe('');
    expect(settled.csrf).not.toBe(primedToken);
  });
});

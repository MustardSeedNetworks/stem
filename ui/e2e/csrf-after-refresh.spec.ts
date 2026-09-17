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

    // Expire the access token only; the refresh cookie stays. Aligning to a
    // stats tick leaves the dashboard's one-second poll a full interval away,
    // so the 401 under test is the mutation's own and not the poll's.
    await page.waitForResponse((r) => r.url().endsWith('/api/v1/stats'));
    const kept = (await context.cookies()).filter((c) => c.name !== ACCESS_COOKIE);
    await context.clearCookies();
    await context.addCookies(kept);

    const refreshed = page.waitForResponse(
      (r) => r.url().endsWith('/api/v1/auth/refresh') && r.status() === 200,
    );
    await expect(start).toBeEnabled();
    await start.click();
    await refreshed;
    await expect.poll(() => starts.length, { timeout: 10000 }).toBe(3);

    const [, rejected, retry] = starts;
    expect(rejected.status).toBe(401);
    expect(rejected.csrf).toBe(primedToken);
    // 403 here is the defect: the retry reusing the token minted for the old
    // bearer, under a session key the daemon has no token for.
    expect(retry.status).not.toBe(403);
    expect(retry.csrf).not.toBe('');
    expect(retry.csrf).not.toBe(primedToken);
  });
});

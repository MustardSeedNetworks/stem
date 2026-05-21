import type { Page } from '@playwright/test';

/**
 * Shared E2E auth helpers.
 *
 * The auth rate limiter in internal/api/ratelimit.go caps real logins
 * at 5/minute per IP. Pre-globalSetup we worked around this by having
 * each spec hydrate auth state locally via mockAuthenticated(); now
 * we do ONE real login in e2e/global-setup.ts and reuse the resulting
 * cookies via playwright.config.ts use.storageState. Specs that
 * exercise the auth flow itself opt back into a clean context with
 *
 *   test.use({ storageState: { cookies: [], origins: [] } });
 *
 * Standardized across the seed/stem/niac trio (see seed#1054).
 */

export const TEST_CREDENTIALS = {
  username: 'admin',
  password: 'admin',
} as const;

/** Path (relative to ui/) where global-setup persists the storage state. */
export const AUTH_STORAGE_STATE = 'playwright/.auth/user.json';

const AUTH_FLAG_KEY = 'stem-authenticated';

/**
 * Skip the login modal for tests whose subject is anything other than the
 * authentication flow itself. Mocks the setup-status probe so the first-run
 * wizard doesn't fire, and pre-sets the same localStorage flag the App reads
 * on mount to hydrate isAuthenticated=true.
 *
 * Must be called before page.goto so the init script runs at the right time.
 */
export async function mockAuthenticated(page: Page): Promise<void> {
  await page.route('**/api/v1/setup/status', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ needsSetup: false }),
    });
  });
  await page.addInitScript((key: string) => {
    window.localStorage.setItem(key, 'true');
  }, AUTH_FLAG_KEY);
}

/**
 * Drive the real login form. Reserve for specs that genuinely test the
 * auth flow end-to-end (auth.spec.ts). Other specs should use
 * mockAuthenticated() to avoid the rate-limit cliff.
 */
export async function loginViaUI(
  page: Page,
  creds: { username: string; password: string } = TEST_CREDENTIALS,
): Promise<void> {
  await page.route('**/api/v1/setup/status', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ needsSetup: false }),
    });
  });
  await page.goto('/');
  await page.getByLabel(/username/i).fill(creds.username);
  await page.getByLabel(/password/i).fill(creds.password);
  await page.getByRole('button', { name: /sign in/i }).click();
  await page.getByRole('button', { name: /logout/i }).waitFor({ state: 'visible' });
}

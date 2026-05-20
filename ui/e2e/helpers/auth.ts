import type { Page } from '@playwright/test';

/**
 * Shared E2E auth helpers.
 *
 * Wave 2 (#77) tightened the auth rate limiter to 5 login attempts per
 * minute per IP (internal/api/ratelimit.go: AuthRateLimit = 5). When every
 * spec drives the real /api/v1/auth/login in beforeEach, the suite blows
 * past that budget after a handful of tests and subsequent specs land on a
 * 429 — the login modal stays mounted and intercepts every shell click.
 *
 * Tests that don't exercise the login flow itself should call
 * mockAuthenticated() instead of pounding the real endpoint.
 */

export const TEST_CREDENTIALS = {
  username: 'admin',
  password: 'admin',
} as const;

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

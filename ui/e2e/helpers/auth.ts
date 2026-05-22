import type { Page } from '@playwright/test';

/**
 * Shared E2E auth helpers.
 *
 * Auth state for the suite is restored from a real login performed
 * once in e2e/global-setup.ts (see #238) and reused via
 * playwright.config.ts use.storageState. Specs exercising the auth
 * flow itself opt back into a clean context with:
 *
 *   test.use({ storageState: { cookies: [], origins: [] } });
 *
 * The only remaining per-spec helper is skipSetupWizard(), which
 * stubs the /api/v1/setup/status probe so the first-run wizard
 * doesn't fire on the synthetic test DB.
 *
 * Standardized across the seed/stem/niac trio (see seed#1054).
 */

export const TEST_CREDENTIALS = {
  username: 'admin',
  password: 'admin',
} as const;

/** Path (relative to ui/) where global-setup persists the storage state. */
export const AUTH_STORAGE_STATE = 'playwright/.auth/user.json';

/**
 * Stub the /api/v1/setup/status probe so the first-run wizard
 * doesn't fire. Must be called before page.goto so the route
 * matcher is attached when the app's initial fetch lands.
 *
 * The legacy localStorage flag (`stem-authenticated`) used to be
 * set here too, but storageState now provides the real auth cookie
 * so the flag is no longer load-bearing.
 */
export async function skipSetupWizard(page: Page): Promise<void> {
  await page.route('**/api/v1/setup/status', (route) => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ needsSetup: false }),
    });
  });
}

/**
 * Drive the real login form. Reserve for specs that genuinely test the
 * auth flow end-to-end (auth.spec.ts). Other specs should use
 * skipSetupWizard() + the global storageState to avoid the
 * rate-limit cliff.
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

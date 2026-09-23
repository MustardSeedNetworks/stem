import type { BrowserContext } from '@playwright/test';

import { TEST_CREDENTIALS } from './auth.ts';

/**
 * Sign the E2E daemon in and persist the session as a Playwright storageState.
 * Two callers share it so they cannot drift apart: e2e/global-setup.ts (the
 * suite) and e2e/phone-width-sign-in.ts (the fleet's 390px gate, which drives
 * its own Playwright and runs this file under plain `node` — hence the
 * explicit `.ts` import above).
 *
 * stem takes its credentials from STEM_AUTH_USERNAME / STEM_AUTH_PASSWORD at
 * startup (scripts/e2e-daemon.sh), so there is no setup wizard to complete
 * first. The login goes through the context's own request client so the
 * Secure session cookies land in the jar the page then uses, and the state is
 * persisted once, after the SPA has mounted signed in.
 */
export async function signInAndPersist(context: BrowserContext, outPath: string): Promise<void> {
  const loginResponse = await context.request.post('/api/v1/auth/login', {
    headers: { 'Content-Type': 'application/json' },
    data: {
      username: TEST_CREDENTIALS.username,
      password: TEST_CREDENTIALS.password,
    },
  });
  if (!loginResponse.ok()) {
    const body = await loginResponse.text();
    throw new Error(
      `sign-in: /api/v1/auth/login returned ${loginResponse.status()}: ${body.slice(0, 200)}`,
    );
  }

  const page = await context.newPage();
  try {
    await page.goto('/');
    // The flag keeps the SPA's auth check from flashing the login modal before
    // the cookie-based session probe lands.
    await page.evaluate(() => {
      window.localStorage.setItem('stem-authenticated', 'true');
    });
  } finally {
    await page.close();
  }

  await context.storageState({ path: outPath });
}

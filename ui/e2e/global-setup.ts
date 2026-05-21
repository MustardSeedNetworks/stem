import { mkdir } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium, type FullConfig, request } from '@playwright/test';

import { AUTH_STORAGE_STATE, TEST_CREDENTIALS } from './helpers/auth';

// ESM equivalent of __dirname (Playwright runs this as ESM).
const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/**
 * One real login at suite start; every spec shares the resulting
 * storageState (cookies + localStorage) via use.storageState in
 * playwright.config.ts. This collapses every spec's per-test login
 * down to exactly 1 real authentication for the whole run, well
 * under the per-IP login rate budget (AuthRateLimit = 5 / minute,
 * internal/api/ratelimit.go).
 *
 * Stem differs from seed: stem's backend takes STEM_AUTH_USERNAME +
 * STEM_AUTH_PASSWORD as env vars at startup (CI sets both to admin),
 * so there is no setup wizard to complete first — global-setup goes
 * straight to /api/v1/auth/login. See e2e/helpers/auth.ts for the
 * shared credential constant and the storage-state path.
 *
 * auth.spec.ts opts back into a clean unauthenticated context with:
 *
 *   test.use({ storageState: { cookies: [], origins: [] } });
 *
 * so it still exercises the real login form end-to-end.
 */
async function globalSetup(config: FullConfig): Promise<void> {
  const [project] = config.projects;
  if (project === undefined) {
    throw new Error('global-setup: no Playwright project configured');
  }
  const baseURL = project.use.baseURL ?? process.env.E2E_BASE_URL ?? 'http://localhost:5173';
  const outPath = resolve(__dirname, '..', AUTH_STORAGE_STATE);

  await mkdir(dirname(outPath), { recursive: true });

  const apiContext = await request.newContext({
    baseURL,
    ignoreHTTPSErrors: true,
  });

  try {
    const loginResponse = await apiContext.post('/api/v1/auth/login', {
      headers: { 'Content-Type': 'application/json' },
      data: {
        username: TEST_CREDENTIALS.username,
        password: TEST_CREDENTIALS.password,
      },
    });
    if (!loginResponse.ok()) {
      const body = await loginResponse.text();
      throw new Error(
        `global-setup: /api/v1/auth/login returned ${loginResponse.status()}: ${body.slice(0, 200)}`,
      );
    }
    await apiContext.storageState({ path: outPath });
  } finally {
    await apiContext.dispose();
  }

  // Attach a localStorage flag for the SPA origin so the in-app auth
  // check doesn't briefly flip the UI back to the login modal before
  // the cookie-based session probe lands. Matches the seed pattern.
  const browser = await chromium.launch();
  try {
    const context = await browser.newContext({
      baseURL,
      ignoreHTTPSErrors: true,
      storageState: outPath,
    });
    const page = await context.newPage();
    await page.goto('/');
    await page.evaluate(() => {
      window.localStorage.setItem('stem-authenticated', 'true');
    });
    await context.storageState({ path: outPath });
  } finally {
    await browser.close();
  }
}

export default globalSetup;

import { mkdir } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { chromium, type FullConfig } from '@playwright/test';

import { AUTH_STORAGE_STATE } from './helpers/auth';
import { signInAndPersist } from './helpers/sign-in';

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
 * The flow itself lives in helpers/sign-in.ts, shared with the
 * phone-width gate's sign-in so the two cannot drift.
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

  const browser = await chromium.launch();
  try {
    const context = await browser.newContext({ baseURL, ignoreHTTPSErrors: true });
    try {
      await signInAndPersist(context, outPath);
    } finally {
      await context.close();
    }
  } finally {
    await browser.close();
  }
}

export default globalSetup;

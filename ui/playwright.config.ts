import { defineConfig, devices } from '@playwright/test';

import { AUTH_STORAGE_STATE } from './e2e/helpers/auth';

/**
 * The suite needs a running stem daemon, not the Vite dev server, so there is
 * no sensible default to fall back to.
 *
 * The dev server is served over http://localhost:3000. stem's session cookies
 * (stem_access, stem_refresh) are Secure, and WebKit will not send a Secure
 * cookie to an insecure origin — every authenticated request 401s, the shell
 * unmounts back to the login overlay, and specs fail on "element was detached
 * from the DOM" (#959). Chromium masks it by treating http://localhost as a
 * trustworthy origin. Defaulting to that URL meant the documented local
 * command could never pass on WebKit, while CI — which sets E2E_BASE_URL to
 * the daemon's HTTPS origin — was green throughout.
 */
function requireBaseURL(): string {
  const fromEnv = process.env.E2E_BASE_URL;
  if (fromEnv) {
    return fromEnv;
  }
  throw new Error(
    'E2E_BASE_URL is not set. Run the suite through ./scripts/run-e2e.sh, which ' +
      'builds stem, starts it on a free port and exports E2E_BASE_URL:\n\n' +
      '  ./scripts/run-e2e.sh --project=webkit\n\n' +
      'To use a daemon you already have running, set it yourself:\n\n' +
      '  E2E_BASE_URL=https://127.0.0.1:8444 npx playwright test\n',
  );
}

/**
 * Playwright E2E Test Configuration
 *
 * End-to-end testing for Stem user flows:
 * - Authentication
 * - License management
 * - Test execution
 * - Settings management
 *
 * Browsers: Chromium (covers Chrome + Edge) and WebKit (covers Safari).
 * Per msn-docs-internal/05-Engineering/E2E_CONVENTIONS.md, no other browsers
 * or viewports are supported.
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  // retries 1 (not 2) — one retry is enough to dodge transient flakes; the
  //   second retry was costing ~30s × N flaky tests with no incremental signal.
  // workers 2 in CI (was 1) — GH Actions runners are 4-vCPU; 1 worker wastes
  //   75% of the box. Mirrors seed #1080 and the cross-repo perf push.
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  timeout: 30000,
  expect: {
    timeout: 10000,
  },
  // Single real login at suite start; persisted to AUTH_STORAGE_STATE
  // and replayed into every test via use.storageState below. See
  // e2e/global-setup.ts and e2e/helpers/auth.ts. Standardized across
  // the seed/stem/niac trio (see seed#1054).
  globalSetup: './e2e/global-setup.ts',
  reporter: [
    ['html', { outputFolder: 'playwright-report' }],
    ['list'],
    ['json', { outputFile: 'playwright-report/results.json' }],
  ],
  use: {
    baseURL: requireBaseURL(),
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'on-first-retry',
    // Gated to local dev only. CI uses the auto-generated self-signed cert
    // and must opt in via PLAYWRIGHT_IGNORE_HTTPS_ERRORS=true in the workflow.
    ignoreHTTPSErrors: process.env.PLAYWRIGHT_IGNORE_HTTPS_ERRORS === 'true' || !process.env.CI,
    // Cookies + localStorage captured by global-setup. Specs that
    // need an unauthenticated context (auth.spec.ts, setup-wizard.spec.ts)
    // override with test.use({ storageState: { cookies: [], origins: [] } }).
    storageState: AUTH_STORAGE_STATE,
  },
  projects: [
    // Per msn-docs-internal/05-Engineering/E2E_CONVENTIONS.md, only chromium
    // (covers Chrome and Edge) and webkit (covers Safari) are supported.
    // Firefox/mobile-chrome/mobile-safari/tablet were never run in CI and
    // had no customer commitment behind them; deleted to stop the config
    // from lying about what's actually tested.
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        launchOptions: { args: ['--enable-features=WebAuthentication'] },
      },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
  ],
  // No webServer. There used to be one starting `npm run dev` on
  // http://localhost:3000, which served a development React build over an
  // insecure origin — neither is what ships, and WebKit cannot authenticate
  // against it at all (see requireBaseURL above).
  //
  // scripts/run-e2e.sh is the entry point. It builds the UI and the binary,
  // starts stem on a free port, waits for /__version, and exports
  // E2E_BASE_URL — the same target CI drives.
});

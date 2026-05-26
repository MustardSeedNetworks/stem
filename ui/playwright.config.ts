import { defineConfig, devices } from '@playwright/test';

import { AUTH_STORAGE_STATE } from './e2e/helpers/auth';

/**
 * Playwright E2E Test Configuration
 *
 * End-to-end testing for Stem user flows:
 * - Authentication
 * - License management
 * - Test execution
 * - Settings management
 *
 * Browsers: Chromium, Firefox, WebKit (Safari)
 * Viewports: Desktop, Tablet, Mobile
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
    baseURL: process.env.E2E_BASE_URL || 'http://localhost:5173',
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
    // Desktop browsers
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    // Mobile viewports
    {
      name: 'mobile-chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'mobile-safari',
      use: { ...devices['iPhone 12'] },
    },
    // Tablet viewport
    {
      name: 'tablet',
      use: { ...devices['iPad (gen 7)'] },
    },
  ],
  // Run local dev server before tests if not in CI
  webServer: process.env.CI
    ? undefined
    : {
        command: 'npm run dev',
        url: 'http://localhost:5173',
        reuseExistingServer: !process.env.CI,
        timeout: 120000,
      },
});

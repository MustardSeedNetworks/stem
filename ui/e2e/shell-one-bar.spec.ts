/**
 * shell-one-bar.spec.ts — one top-of-shell pattern (UI-STEM-9, stem#1268).
 *
 * Owner decision 2026-09-15 (fleet): the shell is the left rail plus the page
 * header, and nothing else sits above the page. Stem shipped a second bar
 * under the header on every route — the connection pill, the role chip, the
 * theme/refresh/logout icons, and, for a Test Master, the interface / peer /
 * port / Start row and the progress bar. The chrome moved into the rail; the
 * run controls moved onto the five pages that run a test.
 *
 * The assertions are structural rather than visual, because that is what the
 * decision is about: where a control lives, not how it looks. `MobileTopBar`
 * is a `<header>`, but it is the phone navigation and sits outside
 * `#main-content`, so scoping the count there reads exactly the rule.
 */

import { expect, type Page, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

const DESKTOP = { width: 1440, height: 900 };
const PHONE = { width: 390, height: 844 };

/** Every `pageRegistry` route — "any route" is only proved by all of them. */
const ROUTES = [
  '/reflector',
  '/tests/benchmark',
  '/tests/servicetest',
  '/tests/trafficgen',
  '/tests/measure',
  '/tests/certify',
  '/history',
  '/account/security',
];

/** The five pages that run a test, and the three that do not. */
const TEST_ROUTES = [
  '/tests/benchmark',
  '/tests/servicetest',
  '/tests/trafficgen',
  '/tests/measure',
  '/tests/certify',
];
const NON_TEST_ROUTES = ['/reflector', '/history', '/account/security'];

/**
 * Both rails are in the DOM at once — the phone drawer and the `hidden lg:flex`
 * desktop rail — so every rail locator is scoped to the one on screen.
 */
const railControl = (page: Page, testId: string) =>
  page.locator(`[data-testid="${testId}"]:visible`);

for (const [name, viewport] of [
  ['desktop 1440x900', DESKTOP],
  ['phone 390x844', PHONE],
] as const) {
  test.describe(`the shell at ${name}`, () => {
    test.beforeEach(async ({ page }) => {
      await skipSetupWizard(page);
      await page.setViewportSize(viewport);
    });

    // A test per route rather than one loop: eight webkit navigations do not
    // fit the suite's per-test budget, and a per-route name says which broke.
    for (const route of ROUTES) {
      test(`${route} renders nothing above the page header`, async ({ page }) => {
        await page.goto(route, { waitUntil: 'domcontentloaded' });
        await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 20000 });
        await expect(page.locator('#main-content header')).toHaveCount(0);
      });
    }
  });
}

test.describe('the rail carries the shell controls', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.setViewportSize(DESKTOP);
    await page.goto('/reflector', { waitUntil: 'domcontentloaded' });
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 20000 });
  });

  test('carries the connection status, and it reads the live state', async ({ page }) => {
    const status = railControl(page, 'rail-status');
    await expect(status).toBeVisible();
    // The dot was hard-coded to the success colour, so the rail said
    // "connected" on a dead socket. The state has to be the live one, and it
    // has to reach a screen reader through a name rather than through colour.
    await expect(status).toHaveAttribute('data-status', /connected|disconnected/);
    await expect(status).toHaveAccessibleName(/connect/i);
  });

  test('carries the theme toggle, and it is keyboard reachable', async ({ page }) => {
    const toggle = railControl(page, 'rail-theme-toggle');
    await expect(toggle).toBeVisible();
    await toggle.focus();
    await expect(toggle).toBeFocused();

    const before = await page.evaluate(() => document.documentElement.classList.contains('dark'));
    await toggle.press('Enter');
    await expect
      .poll(async () => page.evaluate(() => document.documentElement.classList.contains('dark')))
      .toBe(!before);
  });

  test('carries refresh and logout', async ({ page }) => {
    await expect(railControl(page, 'rail-refresh')).toBeVisible();
    await expect(railControl(page, 'rail-logout')).toBeVisible();
  });

  test('carries the role control', async ({ page }) => {
    await expect(railControl(page, 'role-chip-reflector')).toBeVisible();
    await expect(railControl(page, 'role-chip-test_master')).toBeVisible();
  });
});

test.describe('the run controls belong to the pages that run a test', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'test_master');
    await page.setViewportSize(DESKTOP);
  });

  for (const route of TEST_ROUTES) {
    test(`${route} carries them under its header`, async ({ page }) => {
      await page.goto(route, { waitUntil: 'domcontentloaded' });
      await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 20000 });

      const controls = page.getByTestId('test-run-controls');
      await expect(controls).toBeVisible();
      await expect(controls.getByTestId('interface-select')).toBeVisible();
      await expect(controls.getByTestId('peer-input')).toBeVisible();
      await expect(controls.getByTestId('peer-port-input')).toBeVisible();

      // Under the header, not above it — that is the whole decision.
      const header = await page.getByTestId('page-header-title').boundingBox();
      const row = await controls.boundingBox();
      expect(row?.y ?? 0).toBeGreaterThan(header?.y ?? 0);
    });
  }

  for (const route of NON_TEST_ROUTES) {
    test(`${route} does not`, async ({ page }) => {
      await page.goto(route, { waitUntil: 'domcontentloaded' });
      await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 20000 });
      // A Start button on the History page could only ever run someone else's
      // test — that is what the shared bar did on every route.
      await expect(page.getByTestId('test-run-controls')).toHaveCount(0);
    });
  }
});

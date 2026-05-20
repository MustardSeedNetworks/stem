import { expect, test } from '@playwright/test';

/**
 * Module Pages Tests
 *
 * Validates that each test-module page (Benchmark, ServiceTest,
 * TrafficGen, Measure, Certify) listed in the Stem module architecture
 * renders with its module name. After the #66 redesign these moved
 * from "dashboard cards" to dedicated routes under /tests/*; the sidebar
 * has Test group nav links pointing at each.
 *
 * Mocks /api/v1/setup/status so the first-run setup wizard doesn't
 * intercept clicks, then logs in as admin/admin.
 */

const MODULE_NAMES = ['benchmark', 'servicetest', 'trafficgen', 'measure', 'certify'] as const;

test.describe('Module Cards', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/v1/setup/status', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ needsSetup: false }),
      });
    });

    await page.goto('/');
    await page.getByLabel(/username/i).fill('admin');
    await page.getByLabel(/password/i).fill('admin');
    await page.getByRole('button', { name: /sign in/i }).click();
    await expect(page.getByRole('button', { name: /logout/i })).toBeVisible();
  });

  test('should render each module page with its name visible', async ({ page }) => {
    // Navigate to /tests/<module> for each and assert the heading
    // carries the module name. Exercises both routing and the per-
    // module render — stronger than the old "any text visible
    // anywhere" check that broke when the dashboard cards were
    // removed in #66.
    for (const name of MODULE_NAMES) {
      await page.goto(`/tests/${name}`);
      await expect(
        page.getByRole('heading', { name: new RegExp(name, 'i') }).first(),
      ).toBeVisible();
    }
  });
});

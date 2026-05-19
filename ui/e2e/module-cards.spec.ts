import { expect, test } from '@playwright/test';

/**
 * Module Cards Tests
 *
 * Validates that the dashboard exposes the test-module sections
 * (Benchmark, ServiceTest, TrafficGen, Measure, Certify) listed in
 * the Stem module architecture.
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

  test('should display all module names somewhere in the UI', async ({ page }) => {
    // Each module name should appear at least once on the authenticated app.
    for (const name of MODULE_NAMES) {
      await expect(page.getByText(new RegExp(name, 'i')).first()).toBeVisible();
    }
  });
});

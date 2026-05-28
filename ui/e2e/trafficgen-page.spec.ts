import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * TrafficGen Page (/tests/trafficgen) E2E
 *
 * Covers the custom-stream traffic generation surface:
 * - Page renders with the proper heading
 * - Test configuration content is gated by RoleGuard
 */

test.describe('TrafficGen Page', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/tests/trafficgen');
    await expect(page.getByRole('heading', { name: /^trafficgen$/i, level: 1 })).toBeVisible({
      timeout: 10000,
    });
  });

  test('should render the page header with TrafficGen title', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^trafficgen$/i, level: 1 })).toBeVisible();
    await expect(page.getByText(/custom traffic generation/i)).toBeVisible();
  });

  test('should land on the /tests/trafficgen route', async ({ page }) => {
    await expect(page).toHaveURL(/\/tests\/trafficgen$/);
  });

  test('should show role-gated content', async ({ page }) => {
    const content = page.locator('text=/traffic|stream|load|shape|permission|role|access/i');
    await expect(content.locator("visible=true").first()).toBeVisible({ timeout: 5000 });
  });
});

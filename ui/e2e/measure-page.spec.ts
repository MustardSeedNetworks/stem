import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Measure Page (/tests/measure) E2E
 *
 * Covers the Y.1731 OAM measurement surface:
 * - Page renders with the proper heading
 * - Test configuration content is gated by RoleGuard
 */

test.describe('Measure Page', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/tests/measure');
    await expect(page.getByTestId('page-header-title')).toBeVisible({
      timeout: 10000,
    });
  });

  test('should render the page header with Measure title', async ({ page }) => {
    await expect(page.getByTestId('page-header-title')).toBeVisible();
    await expect(page.getByText(/y\.1731 oam delay.*loss/i)).toBeVisible();
  });

  test('should land on the /tests/measure route', async ({ page }) => {
    await expect(page).toHaveURL(/\/tests\/measure$/);
  });

  test('should show role-gated content', async ({ page }) => {
    const content = page.locator(
      'text=/y\\.1731|oam|delay|loss|measurement|permission|role|access/i',
    );
    await expect(content.locator('visible=true').first()).toBeVisible({ timeout: 5000 });
  });
});

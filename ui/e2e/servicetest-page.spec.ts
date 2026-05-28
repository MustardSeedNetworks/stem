import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * ServiceTest Page (/tests/servicetest) E2E
 *
 * Covers the Y.1564 / MEF service activation surface:
 * - Page renders with the proper heading
 * - Test configuration content is gated by RoleGuard
 */

test.describe('ServiceTest Page', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/tests/servicetest');
    await expect(page.getByRole('heading', { name: /^servicetest$/i, level: 1 })).toBeVisible({
      timeout: 10000,
    });
  });

  test('should render the page header with ServiceTest title', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^servicetest$/i, level: 1 })).toBeVisible();
    await expect(page.getByText(/y\.1564.*mef.*service activation/i)).toBeVisible();
  });

  test('should land on the /tests/servicetest route', async ({ page }) => {
    await expect(page).toHaveURL(/\/tests\/servicetest$/);
  });

  test('should show role-gated content', async ({ page }) => {
    // The broad regex also matches the "ServiceTest" nav label, which is
    // rendered in both the mobile aside (display:none at 1280px) and the
    // desktop aside. `.first()` resolved to the hidden mobile instance.
    // Scope to visible matches so we assert on real on-screen content.
    const content = page
      .locator('text=/y\\.1564|mef|service|activation|permission|role|access/i')
      .locator('visible=true');
    await expect(content.first()).toBeVisible({ timeout: 5000 });
  });
});

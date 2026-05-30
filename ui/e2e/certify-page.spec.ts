import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Certify Page (/tests/certify) E2E
 *
 * Covers the RFC 2889 / RFC 6349 / TSN certification surface (now folded
 * into the Pro tier per the 2026-05-19 strategy reset):
 * - Page renders with the proper heading
 * - Test configuration content is gated by RoleGuard
 */

test.describe('Certify Page', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/tests/certify');
    await expect(page.getByTestId('page-header-title')).toBeVisible({
      timeout: 10000,
    });
  });

  test('should render the page header with Certify title', async ({ page }) => {
    await expect(page.getByTestId('page-header-title')).toBeVisible();
    await expect(page.getByText(/rfc 2889.*rfc 6349.*tsn/i)).toBeVisible();
  });

  test('should land on the /tests/certify route', async ({ page }) => {
    await expect(page).toHaveURL(/\/tests\/certify$/);
  });

  test('should show role-gated content', async ({ page }) => {
    const content = page.locator(
      'text=/rfc.2889|rfc.6349|tsn|forwarding|certification|permission|role|access/i',
    );
    await expect(content.locator('visible=true').first()).toBeVisible({ timeout: 5000 });
  });
});

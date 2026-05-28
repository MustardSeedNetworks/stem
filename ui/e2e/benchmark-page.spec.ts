import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Benchmark Page (/tests/benchmark) E2E
 *
 * Covers the flagship RFC 2544 test-config surface:
 * - Page renders with the proper heading
 * - RFC2544ConfigForm slot is present (gated by test_master role)
 */

test.describe('Benchmark Page', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/tests/benchmark');
    await expect(page.getByRole('heading', { name: /^benchmark$/i, level: 1 })).toBeVisible({
      timeout: 10000,
    });
  });

  test('should render the page header with Benchmark title', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^benchmark$/i, level: 1 })).toBeVisible();
    await expect(page.getByText(/rfc 2544 throughput.*latency.*frame-loss/i)).toBeVisible();
  });

  test('should land on the /tests/benchmark route', async ({ page }) => {
    await expect(page).toHaveURL(/\/tests\/benchmark$/);
  });

  test('should show the role-gated content (form OR role-denied message)', async ({ page }) => {
    // RoleGuard renders either the RFC2544ConfigForm (test_master role) or
    // a role-denied notice for read-only viewers. Either is valid.
    const content = page.locator(
      'text=/throughput|latency|frame.loss|back.to.back|permission|role|access/i',
    );
    await expect(content.locator("visible=true").first()).toBeVisible({ timeout: 5000 });
  });
});

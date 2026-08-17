import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * ServiceTest Page (/tests/servicetest) E2E
 *
 * Covers the Y.1564 / MEF service-activation surface and the RoleGuard contract that wraps it.
 *
 * The previous role test asserted one broad regex —
 * /y\\.1564|mef|service|activation|permission|role|access/i
 * — described as "the form OR a role-denied message". Both halves matched
 * ordinary page copy (the header description alone satisfies it), so it passed
 * in either role and could not fail if the module disappeared entirely.
 *
 * These assert the two states separately via testids, so they stay honest
 * under the es locale too.
 */

test.describe('ServiceTest Page', () => {
  test('renders the page header', async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/tests/servicetest');
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });
  });

  test('lands on the /tests/servicetest route', async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/tests/servicetest');
    await expect(page).toHaveURL(/\/tests\/servicetest$/);
  });

  test('shows no role banner as test_master', async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'test_master');
    await page.goto('/tests/servicetest');
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });

    await expect(page.getByTestId('role-guard-banner')).toHaveCount(0);
  });

  test('warns a reflector that the module needs test_master', async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'reflector');
    await page.goto('/tests/servicetest');
    await expect(page.getByTestId('role-guard-banner')).toBeVisible({ timeout: 10000 });
  });
});

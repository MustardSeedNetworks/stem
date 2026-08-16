import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * Benchmark Page (/tests/benchmark) E2E
 *
 * Covers the flagship RFC 2544 test-config surface and the RoleGuard contract that wraps it.
 *
 * The previous role test asserted one broad regex —
 * /throughput|latency|frame.loss|back.to.back|permission|role|access/i
 * — described as "the form OR a role-denied message". Both halves matched
 * ordinary page copy (the header description alone satisfies it), so it passed
 * in either role and could not fail if the module disappeared entirely.
 *
 * These assert the two states separately via testids, so they stay honest
 * under the es locale too.
 */

test.describe('Benchmark Page', () => {
  test('renders the page header', async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/tests/benchmark');
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });
  });

  test('lands on the /tests/benchmark route', async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/tests/benchmark');
    await expect(page).toHaveURL(/\/tests\/benchmark$/);
  });

  test('shows no role banner as test_master', async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'test_master');
    await page.goto('/tests/benchmark');
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });

    // Benchmark is the one module whose tests are selected by default
    // (test-store.ts selectedTests), so its form renders without any
    // Settings setup. The other modules' forms return null until their test
    // type is picked, which is why only this page asserts the form.
    await expect(page.getByTestId('rfc2544-config-form')).toBeVisible();
    await expect(page.getByTestId('role-guard-banner')).toHaveCount(0);
  });

  test('warns a reflector that the module needs test_master', async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'reflector');
    await page.goto('/tests/benchmark');
    await expect(page.getByTestId('role-guard-banner')).toBeVisible({ timeout: 10000 });
  });
});

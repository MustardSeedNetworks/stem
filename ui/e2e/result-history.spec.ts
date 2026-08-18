import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Result History Tests
 *
 * History is a page, not a drawer: the archive used to sit behind a rail
 * button while the page showed only the last run, which was one thing on two
 * surfaces. These exercise the page.
 *
 * Uses skipSetupWizard() to skip the login modal — these tests don't
 * exercise the auth flow itself (see helpers/auth.ts).
 */

test.describe('Result History', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/history');
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });
  });

  test('is reachable from the sidebar', async ({ page }) => {
    await page.goto('/');
    // Rail items are buttons that navigate, not anchors.
    await page.getByRole('button', { name: 'History', exact: true }).first().click();

    await expect(page).toHaveURL(/\/history$/);
    await expect(page.getByTestId('page-header-title')).toHaveText('History');
  });

  test('says so plainly when no run has been recorded', async ({ page }) => {
    // A fresh browser profile has an empty archive; the page must say that
    // rather than render an empty frame.
    await expect(page.getByText(/no test has completed yet/i)).toBeVisible();
  });

  test('renders the page rather than a drawer', async ({ page }) => {
    await expect(page.getByTestId('history-drawer')).toHaveCount(0);
    await expect(page.getByTestId('sidebar-history-button')).toHaveCount(0);
  });
});

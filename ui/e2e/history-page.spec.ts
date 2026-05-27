import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * History Page (/history) E2E
 *
 * Covers the test result history surface:
 * - Page renders with the proper heading
 * - Either a result snapshot or an "open the history drawer" hint is visible
 */

test.describe('History Page', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/history');
    await expect(page.getByRole('heading', { name: /^history$/i, level: 1 })).toBeVisible({
      timeout: 10000,
    });
  });

  test('should render the page header with History title', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /^history$/i, level: 1 })).toBeVisible();
    await expect(page.getByText(/latest test result snapshot|history drawer/i)).toBeVisible();
  });

  test('should land on the /history route', async ({ page }) => {
    await expect(page).toHaveURL(/\/history$/);
  });

  test('should render either a result snapshot or an empty hint', async ({ page }) => {
    // Scope to the page's main content section — the sidebar also contains
    // a "History" label which is hidden on mobile, so an unscoped
    // .first() match races into the sidebar entry and times out.
    const main = page.locator('main, section').first();
    const content = main.locator('text=/test.type|result|snapshot|history|open the.*drawer/i');
    await expect(content.first()).toBeVisible({ timeout: 5000 });
  });
});

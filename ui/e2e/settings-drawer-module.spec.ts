import { expect, test } from '@playwright/test';

/**
 * Settings drawer — Module view
 *
 * Mocks /api/v1/setup/status so the first-run setup wizard doesn't
 * intercept clicks on the settings drawer toggle.
 */

test.describe('Settings drawer module view', () => {
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

  test('switches to module view and shows modules', async ({ page }) => {
    await page.getByRole('button', { name: /open settings/i }).click();

    const drawer = page.getByRole('dialog', { name: /settings/i });
    await expect(drawer).toBeVisible();

    await drawer.getByRole('button', { name: 'Module', exact: true }).click();

    await expect(drawer.getByText(/benchmark/i).first()).toBeVisible();
    await expect(drawer.getByText(/reflector/i).first()).toBeVisible();
  });
});

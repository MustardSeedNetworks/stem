import { expect, test } from '@playwright/test';

test.describe('Settings', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin');
    await page.click('button[type="submit"]');
  });

  test('should navigate to settings page', async ({ page }) => {
    await page.click('[data-testid="settings-link"]');
    await expect(page.getByRole('heading', { name: /settings/i })).toBeVisible();
  });

  test('should display license information', async ({ page }) => {
    await page.goto('/settings');
    await expect(page.getByText(/license|tier/i)).toBeVisible();
  });
});

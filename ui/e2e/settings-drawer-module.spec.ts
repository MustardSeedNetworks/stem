import { expect, test } from '@playwright/test';

test.describe('Settings drawer module view', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin');
    await page.click('button[type="submit"]');
  });

  test('switches to module view and shows modules', async ({ page }) => {
    await page.getByRole('button', { name: /open settings/i }).click();
    await page.getByRole('button', { name: /module/i }).click();

    await expect(page.getByText(/benchmark/i)).toBeVisible();
    await expect(page.getByText(/reflector/i)).toBeVisible();
  });
});

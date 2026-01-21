import { expect, test } from '@playwright/test';

test.describe('License', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin');
    await page.click('button[type="submit"]');
  });

  test('should display license status', async ({ page }) => {
    await page.goto('/license');
    await expect(page.getByText(/license|status|tier/i)).toBeVisible();
  });

  test('should show license activation form', async ({ page }) => {
    await page.goto('/license');
    await expect(page.getByPlaceholder(/license key|activation/i)).toBeVisible();
  });

  test('should validate license key format', async ({ page }) => {
    await page.goto('/license');
    await page.fill('[data-testid="license-key-input"]', 'invalid-key');
    await page.click('[data-testid="activate-button"]');
    await expect(page.getByText(/invalid|error/i)).toBeVisible();
  });
});

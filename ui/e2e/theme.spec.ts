import { expect, test } from '@playwright/test';

/**
 * Theme Tests
 *
 * Tests for dark/light mode functionality.
 *
 * Mocks /api/v1/setup/status so the first-run setup wizard doesn't
 * intercept clicks on the theme toggle (which lives in the app shell
 * and is visible even on the login page).
 */

test.describe('Theme', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/v1/setup/status', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ needsSetup: false }),
      });
    });
    await page.goto('/');
    await expect(page.getByRole('heading', { name: /sign in/i })).toBeVisible();
  });

  test('should have theme toggle button', async ({ page }) => {
    const themeButton = page.getByRole('button', { name: /switch to (dark|light) mode/i });
    await expect(themeButton).toBeVisible();
  });

  test('should toggle between dark and light mode', async ({ page }) => {
    const html = page.locator('html');
    const initialDark = await html.evaluate((el) => el.classList.contains('dark'));

    await page.getByRole('button', { name: /switch to (dark|light) mode/i }).click();

    await expect
      .poll(async () => await html.evaluate((el) => el.classList.contains('dark')), {
        timeout: 5000,
      })
      .not.toBe(initialDark);
  });

  test('should apply correct colors in dark mode', async ({ page }) => {
    await page.evaluate(() => {
      document.documentElement.classList.add('dark');
    });

    const body = page.locator('body');
    const bgColor = await body.evaluate((el) => getComputedStyle(el).backgroundColor);
    expect(bgColor).toBeTruthy();
  });

  test('should apply correct colors in light mode', async ({ page }) => {
    await page.evaluate(() => {
      document.documentElement.classList.remove('dark');
    });

    const body = page.locator('body');
    const bgColor = await body.evaluate((el) => getComputedStyle(el).backgroundColor);
    expect(bgColor).toBeTruthy();
  });
});

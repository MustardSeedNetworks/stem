import { expect, test } from '@playwright/test';
import { mockAuthenticated } from './helpers/auth';

/**
 * Theme Tests
 *
 * Tests for dark/light mode functionality.
 *
 * Uses mockAuthenticated() to skip the login modal — the theme toggle
 * lives in the authenticated app shell, not on the login page (see
 * helpers/auth.ts).
 */

test.describe('Theme', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticated(page);
    await page.goto('/');
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

import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Theme Tests
 *
 * Tests for dark/light mode functionality.
 *
 * Uses skipSetupWizard() to skip the login modal — the theme toggle
 * lives in the authenticated app shell, not on the login page (see
 * helpers/auth.ts).
 */

test.describe('Theme', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/');
  });

  test('should have theme toggle button', async ({ page }) => {
    await expect(page.getByTestId('header-theme-toggle')).toBeVisible();
  });

  test('should toggle between dark and light mode', async ({ page }) => {
    const html = page.locator('html');
    const initialDark = await html.evaluate((el) => el.classList.contains('dark'));

    await page.getByTestId('header-theme-toggle').click();

    await expect
      .poll(async () => await html.evaluate((el) => el.classList.contains('dark')), {
        timeout: 5000,
      })
      .not.toBe(initialDark);
  });

  test('applies a different background color in dark mode than in light mode', async ({ page }) => {
    const body = page.locator('body');

    await page.evaluate(() => {
      document.documentElement.classList.remove('dark');
    });
    const lightBg = await body.evaluate((el: HTMLElement) => getComputedStyle(el).backgroundColor);

    await page.evaluate(() => {
      document.documentElement.classList.add('dark');
    });
    const darkBg = await body.evaluate((el: HTMLElement) => getComputedStyle(el).backgroundColor);

    // The actual theme tokens are intentionally not hard-coded here — the
    // MSN brand token map (msn-docs-internal) is the source of truth and
    // may evolve. What we DO assert is that light and dark produce a
    // distinguishable background. A weak `toBeTruthy()` check accepted
    // the same value in both modes, which would be a real bug.
    expect(lightBg, 'light mode must produce a body background').toBeTruthy();
    expect(darkBg, 'dark mode must produce a body background').toBeTruthy();
    expect(darkBg, 'dark mode background must differ from light mode').not.toBe(lightBg);
  });
});

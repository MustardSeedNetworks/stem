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
    await expect(page.getByTestId('rail-theme-toggle')).toBeVisible();
  });

  test('should toggle between dark and light mode', async ({ page }) => {
    const html = page.locator('html');
    const initialDark = await html.evaluate((el) => el.classList.contains('dark'));

    await page.getByTestId('rail-theme-toggle').click();

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

/**
 * A fresh profile follows the OS (fleet decision 2026-09-15, UI-STEM-4). The
 * 2026-09-15 audit saw dark regardless of `prefers-color-scheme`, because the
 * unset preference resolved to 'dark' and index.html shipped `class="dark"`.
 * Each scheme gets its own test so both directions are proved: a hard-coded
 * default of either value fails one of them.
 */
for (const scheme of ['light', 'dark'] as const) {
  test.describe(`a fresh profile with the OS set to ${scheme}`, () => {
    test.use({ colorScheme: scheme });

    test.beforeEach(async ({ page }) => {
      await skipSetupWizard(page);
      await page.addInitScript(() => window.localStorage.removeItem('stem-theme'));
    });

    test(`opens in ${scheme}`, async ({ page }) => {
      await page.goto('/', { waitUntil: 'domcontentloaded' });
      await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 20000 });

      const isDark = await page.evaluate(() => document.documentElement.classList.contains('dark'));
      expect(isDark).toBe(scheme === 'dark');
    });
  });
}

import { expect, test } from '@playwright/test';

/**
 * Help Drawer Tests
 *
 * Tests for the help documentation drawer:
 * - Open/close functionality
 * - Help content display
 *
 * Mocks /api/v1/setup/status so the first-run setup wizard doesn't
 * intercept clicks on the sidebar's help button.
 */

test.describe('Help Drawer', () => {
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

  test('should have help button in header', async ({ page }) => {
    const helpButton = page.getByRole('button', { name: /open help/i });
    await expect(helpButton).toBeVisible();
  });

  test('should open help drawer when clicking help button', async ({ page }) => {
    await page.getByRole('button', { name: /open help/i }).click();

    const drawer = page.getByRole('dialog', { name: /help.*documentation/i });
    await expect(drawer).toBeVisible();
  });

  test('should close help drawer when clicking close button', async ({ page }) => {
    await page.getByRole('button', { name: /open help/i }).click();

    const drawer = page.getByRole('dialog', { name: /help.*documentation/i });
    await expect(drawer).toBeVisible();

    await page.getByRole('button', { name: /close help/i }).click();
    await expect(drawer).not.toBeVisible();
  });

  test('should display help content', async ({ page }) => {
    await page.getByRole('button', { name: /open help/i }).click();

    const drawer = page.getByRole('dialog', { name: /help.*documentation/i });
    await expect(drawer).toBeVisible();

    const text = await drawer.textContent();
    expect(text?.length ?? 0).toBeGreaterThan(50);
  });
});

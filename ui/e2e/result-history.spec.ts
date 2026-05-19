import { expect, test } from '@playwright/test';

/**
 * Result History Tests
 *
 * Tests for the result history drawer triggered from the header bar.
 *
 * Mocks /api/v1/setup/status so the first-run setup wizard doesn't
 * intercept clicks on the history button.
 */

test.describe('Result History', () => {
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

  test('should have history button in header', async ({ page }) => {
    await expect(page.getByRole('button', { name: /open test history/i })).toBeVisible();
  });

  test('should open history drawer when clicking history button', async ({ page }) => {
    await page.getByRole('button', { name: /open test history/i }).click();

    const drawer = page.getByRole('dialog', { name: /test history/i });
    await expect(drawer).toBeVisible();
  });

  test('should display content in history drawer', async ({ page }) => {
    await page.getByRole('button', { name: /open test history/i }).click();

    const drawer = page.getByRole('dialog', { name: /test history/i });
    await expect(drawer).toBeVisible();

    const text = await drawer.textContent();
    expect(text?.length ?? 0).toBeGreaterThan(0);
  });

  test('should close history drawer', async ({ page }) => {
    await page.getByRole('button', { name: /open test history/i }).click();

    const drawer = page.getByRole('dialog', { name: /test history/i });
    await expect(drawer).toBeVisible();

    await page
      .getByRole('button', { name: /close history drawer|close test history/i })
      .first()
      .click();
    await expect(drawer).not.toBeVisible();
  });
});

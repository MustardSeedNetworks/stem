import { expect, test } from '@playwright/test';
import { mockAuthenticated } from './helpers/auth';

/**
 * Help Drawer Tests
 *
 * Tests for the help documentation drawer:
 * - Open/close functionality
 * - Help content display
 *
 * Uses mockAuthenticated() to skip the login modal — these tests don't
 * exercise the auth flow itself (see helpers/auth.ts).
 */

test.describe('Help Drawer', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticated(page);
    await page.goto('/');
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

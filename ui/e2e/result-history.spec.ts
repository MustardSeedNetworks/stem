import { expect, test } from '@playwright/test';
import { mockAuthenticated } from './helpers/auth';

/**
 * Result History Tests
 *
 * Tests for the result history drawer triggered from the header bar.
 *
 * Uses mockAuthenticated() to skip the login modal — these tests don't
 * exercise the auth flow itself (see helpers/auth.ts).
 */

test.describe('Result History', () => {
  test.beforeEach(async ({ page }) => {
    await mockAuthenticated(page);
    await page.goto('/');
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

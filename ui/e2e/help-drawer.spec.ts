import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Help Drawer Tests
 *
 * Tests for the help documentation drawer:
 * - Open/close functionality
 * - Help content display
 *
 * Uses skipSetupWizard() to skip the login modal — these tests don't
 * exercise the auth flow itself (see helpers/auth.ts).
 */

test.describe('Help Drawer', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/');
  });

  test('should have help button in sidebar', async ({ page }) => {
    await expect(page.getByTestId('sidebar-help-button')).toBeVisible();
  });

  test('should open help drawer when clicking help button', async ({ page }) => {
    await page.getByTestId('sidebar-help-button').click();
    await expect(page.getByTestId('help-drawer')).toBeVisible();
  });

  test('should close help drawer when clicking close button', async ({ page }) => {
    await page.getByTestId('sidebar-help-button').click();

    const drawer = page.getByTestId('help-drawer');
    await expect(drawer).toBeVisible();

    await page.getByTestId('help-drawer-close').click();
    await expect(drawer).not.toBeVisible();
  });

  test('should display help content', async ({ page }) => {
    await page.getByTestId('sidebar-help-button').click();

    const drawer = page.getByTestId('help-drawer');
    await expect(drawer).toBeVisible();

    const text = await drawer.textContent();
    expect(text?.length ?? 0).toBeGreaterThan(50);
  });
});

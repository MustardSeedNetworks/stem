import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Settings drawer — Module view
 *
 * Uses skipSetupWizard() to skip the login modal — these tests don't
 * exercise the auth flow itself (see helpers/auth.ts).
 *
 * The ViewToggle (Standard | Module) only renders when the stem role is
 * test_master (#210 — Reflector role doesn't need test selection). The
 * default persisted role is reflector, so we hydrate the role-storage
 * key to test_master before navigation.
 */

test.describe('Settings drawer module view', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.addInitScript(() => {
      window.localStorage.setItem('stem-role', 'test_master');
    });
    await page.goto('/');
  });

  test('switches to module view and shows modules', async ({ page }) => {
    await page.getByTestId('sidebar-settings-button').click();

    const drawer = page.getByTestId('settings-drawer');
    await expect(drawer).toBeVisible();

    await drawer.getByRole('button', { name: 'Module', exact: true }).click();

    await expect(drawer.getByText(/benchmark/i).first()).toBeVisible();
    await expect(drawer.getByText(/reflector/i).first()).toBeVisible();
  });
});

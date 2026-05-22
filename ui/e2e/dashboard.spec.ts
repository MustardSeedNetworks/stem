import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Dashboard Tests
 *
 * Tests for the main dashboard functionality:
 * - Stats cards display
 * - Interface selection
 * - Connection status
 * - Test controls
 *
 * Uses skipSetupWizard() to skip the login modal — these tests don't
 * exercise the auth flow itself and shouldn't burn the suite-wide
 * 5-per-minute auth rate budget (see helpers/auth.ts for the why).
 */

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/');
  });

  test('should display stats cards', async ({ page }) => {
    await expect(page.getByText(/packets received/i)).toBeVisible();
    await expect(page.getByText(/packets sent/i)).toBeVisible();
    await expect(page.getByText(/current rate/i)).toBeVisible();
    await expect(page.getByText(/uptime/i)).toBeVisible();
  });

  test('should display interface selector', async ({ page }) => {
    const interfaceSelect = page.locator('select').first();
    await expect(interfaceSelect).toBeVisible();
  });

  test('should display connection status', async ({ page }) => {
    const statusBadge = page.locator('.status-badge').first();
    await expect(statusBadge).toBeVisible();
  });

  test('should land on the Reflector page after login', async ({ page }) => {
    // After the #66 redesign there is no "Test Modules" dashboard
    // section anymore — module pages live under /tests/* in the
    // sidebar, and `/` redirects to `/reflector`. Assert we landed
    // there by checking for the Reflector page heading.
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible();
  });

  test('should have start/stop test buttons', async ({ page }) => {
    const testButton = page.getByRole('button', { name: /start|stop/i });
    await expect(testButton.first()).toBeVisible();
  });
});

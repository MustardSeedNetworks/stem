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
    // Stable testids on each StatsCard — i18n-stable. Previously
    // matched by /packets received/i etc., which broke on es locale.
    await expect(page.getByTestId('stats-packets-received')).toBeVisible();
    await expect(page.getByTestId('stats-packets-sent')).toBeVisible();
    await expect(page.getByTestId('stats-current-rate')).toBeVisible();
    await expect(page.getByTestId('stats-uptime')).toBeVisible();
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
    const testButton = page.getByTestId('reflector-start-button');
    await expect(testButton.first()).toBeVisible();
  });
});

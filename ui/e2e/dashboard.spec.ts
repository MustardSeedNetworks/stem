import { expect, test } from '@playwright/test';

/**
 * Dashboard Tests
 *
 * Tests for the main dashboard functionality:
 * - Stats cards display
 * - Interface selection
 * - Connection status
 * - Test controls
 *
 * Mocks /api/v1/setup/status so the first-run setup wizard is dismissed
 * and then performs an admin/admin login, matching CI's STEM_AUTH_USERNAME
 * and STEM_AUTH_PASSWORD env vars.
 */

test.describe('Dashboard', () => {
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

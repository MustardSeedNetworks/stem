import { expect, test } from '@playwright/test';
import { AUTH_STORAGE_STATE } from './helpers/auth';

const VERSION_KEYS = ['version', 'commit', 'buildTime', 'uiBuildHash'] as const;

test.describe('smoke @ unauthenticated', { tag: '@smoke' }, () => {
  test.use({ storageState: { cookies: [], origins: [] } });

  test('GET /__version returns canonical build metadata', async ({ request }) => {
    const res = await request.get('/__version');
    expect(res.status()).toBe(200);
    const body = await res.json();
    for (const k of VERSION_KEYS) {
      expect(body[k], `missing ${k} in /__version`).toBeTruthy();
      expect(typeof body[k]).toBe('string');
    }
  });

  test('login surface renders for unauthenticated visitors', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByTestId('login-title')).toBeVisible({ timeout: 10000 });
  });
});

test.describe('smoke @ authenticated', { tag: '@smoke' }, () => {
  test.use({ storageState: AUTH_STORAGE_STATE });

  test('reflector shell renders with page-header-title', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });
  });

  test('theme toggle is interactive', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });
    const toggle = page.getByTestId('header-theme-toggle');
    await expect(toggle).toBeVisible();
    await toggle.click();
    await expect(toggle).toBeVisible();
  });

  test('settings drawer opens and closes', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });
    await page.getByTestId('sidebar-settings-button').click();
    await expect(page.getByTestId('settings-drawer')).toBeVisible();
    await page.getByTestId('settings-drawer-close').click();
    await expect(page.getByTestId('settings-drawer')).toBeHidden();
  });

  test('help drawer opens with version badge', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });
    await page.getByTestId('sidebar-help-button').click();
    await expect(page.getByTestId('help-drawer')).toBeVisible();
    await expect(page.getByTestId('help-drawer-version')).toContainText(/v\d/);
    await page.getByTestId('help-drawer-close').click();
    await expect(page.getByTestId('help-drawer')).toBeHidden();
  });

  test('history page loads @smoke', async ({ page }) => {
    await page.goto('/history');
    await expect(page.getByTestId('page-header-title')).toHaveText('History', { timeout: 10000 });
  });
});

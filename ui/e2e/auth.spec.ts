import { expect, test } from '@playwright/test';
import { skipSetupWizard, TEST_CREDENTIALS } from './helpers/auth';

// Opt out of the suite-wide authenticated storageState (from
// e2e/global-setup.ts) so each test starts from a clean
// unauthenticated context — this spec exercises the real form.
test.use({ storageState: { cookies: [], origins: [] } });

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  test('should show login page', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByTestId('login-title')).toBeVisible();
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/');
    await page.getByTestId('login-username').fill('invalid');
    await page.getByTestId('login-password').fill('wrongpassword');
    await page.getByTestId('login-submit').click();
    // role=alert is i18n-stable; the login error surface emits it
    // natively. Previous /invalid|error|failed/i regex would miss
    // under es locale ("invalido", "fallido", etc.).
    await expect(page.getByRole('alert')).toBeVisible();
  });

  test('should login with valid credentials', async ({ page }) => {
    await page.goto('/');
    await page.getByTestId('login-username').fill(TEST_CREDENTIALS.username);
    await page.getByTestId('login-password').fill(TEST_CREDENTIALS.password);
    await page.getByTestId('login-submit').click();
    await expect(page.locator('[data-testid="logout-button"]')).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    // Login first
    await page.goto('/');
    await page.getByTestId('login-username').fill(TEST_CREDENTIALS.username);
    await page.getByTestId('login-password').fill(TEST_CREDENTIALS.password);
    await page.getByTestId('login-submit').click();

    // Then logout
    await page.click('[data-testid="logout-button"]');
    await expect(page.getByTestId('login-title')).toBeVisible();
  });
});

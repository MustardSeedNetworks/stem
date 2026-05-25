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
    await expect(page.getByRole('heading', { name: /sign in/i })).toBeVisible();
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/');
    await page.getByLabel(/username/i).fill('invalid');
    await page.getByLabel(/password/i).fill('wrongpassword');
    await page.getByRole('button', { name: /sign in/i }).click();
    await expect(page.getByText(/invalid|error|failed/i)).toBeVisible();
  });

  test('should login with valid credentials', async ({ page }) => {
    await page.goto('/');
    await page.getByLabel(/username/i).fill(TEST_CREDENTIALS.username);
    await page.getByLabel(/password/i).fill(TEST_CREDENTIALS.password);
    await page.getByRole('button', { name: /sign in/i }).click();
    await expect(page.getByRole('button', { name: /logout/i })).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    // Login first
    await page.goto('/');
    await page.getByLabel(/username/i).fill(TEST_CREDENTIALS.username);
    await page.getByLabel(/password/i).fill(TEST_CREDENTIALS.password);
    await page.getByRole('button', { name: /sign in/i }).click();

    // Then logout
    await page.click('[data-testid="logout-button"]');
    await expect(page.getByRole('heading', { name: /sign in/i })).toBeVisible();
  });
});

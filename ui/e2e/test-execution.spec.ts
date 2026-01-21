import { expect, test } from '@playwright/test';

test.describe('Test Execution', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/');
    await page.fill('[name="username"]', 'admin');
    await page.fill('[name="password"]', 'admin');
    await page.click('button[type="submit"]');
  });

  test('should display interface selector', async ({ page }) => {
    await expect(page.getByText(/interface|network/i)).toBeVisible();
  });

  test('should display test type options', async ({ page }) => {
    await expect(page.getByText(/throughput|latency|benchmark/i)).toBeVisible();
  });

  test('should start a test', async ({ page }) => {
    // Select interface
    await page.click('[data-testid="interface-selector"]');
    await page.click('[data-testid="interface-option"]');

    // Select test type
    await page.click('[data-testid="test-type-selector"]');
    await page.click('[data-testid="throughput-option"]');

    // Start test
    await page.click('[data-testid="start-test-button"]');
    await expect(page.getByText(/running|progress|started/i)).toBeVisible();
  });

  test('should display test results', async ({ page }) => {
    await page.goto('/results');
    await expect(page.getByText(/results|history/i)).toBeVisible();
  });
});

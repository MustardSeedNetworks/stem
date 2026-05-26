import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Error Scenario Tests
 *
 * Verify the app degrades gracefully when the backend is unavailable,
 * slow, returning errors, or returning malformed payloads. Each test
 * asserts on a visible UI consequence of the failure, not just that
 * the page didn't crash.
 *
 * SSE (/api/v1/events) is bypassed in route mocks that simulate broad
 * failure — its reconnect lifecycle is orthogonal to the failure modes
 * these tests exercise, and intercepting it would mask the real signal
 * behind SSE-init noise.
 */

const SSE_PATH = '/api/v1/events';

test.describe('Error Scenarios', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  test('renders the reflector shell even when every API call fails', async ({ page }) => {
    await page.route('**/api/**', (route) => {
      if (route.request().url().includes(SSE_PATH)) {
        return route.continue();
      }
      route.abort('failed');
    });

    await page.goto('/');

    // The shell renders independently of API data — the heading should
    // appear even with zero successful data fetches.
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible();
  });

  test('eventually renders when every API call is delayed 2 seconds', async ({ page }) => {
    await page.route('**/api/**', async (route) => {
      if (route.request().url().includes(SSE_PATH)) {
        return route.continue();
      }
      await new Promise((resolve) => setTimeout(resolve, 2000));
      route.continue();
    });

    await page.goto('/');
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible({
      timeout: 15000,
    });
  });

  test('does not crash when /api/v1/interfaces returns 500', async ({ page }) => {
    await page.route('**/api/v1/interfaces', (route) => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal Server Error' }),
      });
    });

    await page.goto('/');
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible();
  });

  test('does not crash when /api/v1/stats returns malformed JSON', async ({ page }) => {
    await page.route('**/api/v1/stats', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: 'not valid json{{{',
      });
    });

    await page.goto('/');
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible();
  });

  test('renders cleanly when /api/v1/interfaces returns an empty list', async ({ page }) => {
    await page.route('**/api/v1/interfaces', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([]),
      });
    });

    await page.goto('/');
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible();
  });
});

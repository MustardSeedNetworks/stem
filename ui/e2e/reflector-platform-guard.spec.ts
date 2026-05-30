import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Reflector platform guard
 *
 * The macOS and Windows builds of stem are pure-Go (CGO disabled) and
 * cannot run the line-rate Reflector dataplane. The backend exposes
 * this via GET /api/v1/capabilities, and the ReflectorPage uses the
 * payload to show a warning banner and disable the Start button.
 *
 * These specs mock the capabilities endpoint so the platform-guard UX
 * is reachable from any CI runner (Linux or otherwise). Uses
 * skipSetupWizard() to skip the login modal (see helpers/auth.ts).
 */

test.describe('Reflector page platform guard', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  test('shows banner and disables Start button when reflector is unsupported', async ({ page }) => {
    await page.route('**/api/v1/capabilities', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          reflector: { supported: false, reason: 'CGO + Linux required' },
          testMaster: { supported: true },
        }),
      });
    });

    await page.goto('/');

    // App auto-redirects "/" to "/reflector"; assert URL just to be sure.
    await expect(page).toHaveURL(/\/reflector(\?|$)/);

    // Banner is visible with the title + the reason from the payload.
    await expect(
      page.getByText(/Reflector mode is not available on this platform\./i),
    ).toBeVisible();
    await expect(page.getByText(/CGO \+ Linux required/i)).toBeVisible();

    // The Switch to Test Master button is reachable inside the banner.
    await expect(page.getByTestId('role-chip-test-master')).toBeVisible();

    // The Start button is disabled with the platform tooltip.
    const startButton = page.getByTestId('reflector-start-button');
    await expect(startButton).toBeVisible();
    await expect(startButton).toBeDisabled();
    await expect(startButton).toHaveAttribute(
      'title',
      /Reflector mode is not available on this platform/i,
    );
  });

  test('renders no platform banner when reflector is supported', async ({ page }) => {
    await page.route('**/api/v1/capabilities', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          reflector: { supported: true },
          testMaster: { supported: true },
        }),
      });
    });

    await page.goto('/');

    await expect(page).toHaveURL(/\/reflector(\?|$)/);

    // No platform-guard banner.
    await expect(page.getByText(/Reflector mode is not available on this platform\./i)).toHaveCount(
      0,
    );

    // Start button has no platform tooltip — it may still be disabled
    // because no interface is selected, but that's a different code
    // path and not the concern of this spec.
    const startButton = page.getByTestId('reflector-start-button');
    await expect(startButton).toBeVisible();
    await expect(startButton).not.toHaveAttribute(
      'title',
      /Reflector mode is not available on this platform/i,
    );
  });
});

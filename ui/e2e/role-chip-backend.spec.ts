import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * RoleChip backend wiring (issue #74)
 *
 * The RoleChip POSTs to /api/v1/mode and only updates local state
 * after the backend accepts the change. These specs mock the
 * mode-switch endpoint so behaviour is reproducible on any runner —
 * we never depend on a real reflector dataplane being available.
 *
 * Coverage:
 *  - POST body is the JSON {"mode":"reflector"|"test_master"} the
 *    backend expects (see internal/api/handlers_settings.go).
 *  - On a 200 OK the chip activates the new role.
 *  - On a 4xx error the inline error tag appears and the role
 *    stays put.
 *
 * Uses skipSetupWizard() to skip the login modal — the auth flow is
 * covered by auth.spec.ts and pounding /api/v1/auth/login from every
 * spec blows past the 5-per-minute rate budget (see helpers/auth.ts).
 */

test.describe('RoleChip backend wiring', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);

    // Capabilities probe — both modes supported so the platform
    // guard does not pre-empt our spec.
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
  });

  test('POSTs mode change and updates the chip on success', async ({ page }) => {
    let observedBody: string | null = null;

    await page.route('**/api/v1/mode', (route) => {
      const request = route.request();
      if (request.method() !== 'POST') {
        route.fallback();
        return;
      }
      observedBody = request.postData();
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          status: 'updated',
          mode: 'test_master',
          previous: 'reflector',
        }),
      });
    });

    await page.goto('/');

    // The chip should be present in the header.
    const testMasterChip = page.getByTestId('role-chip-test_master');
    await expect(testMasterChip).toBeVisible();

    // Click test_master, then confirm the ConfirmModal.
    await testMasterChip.click();
    await page.getByRole('button', { name: /switch role/i }).click();

    // Wait for the POST to have happened and the chip to reflect
    // the server's echoed mode.
    await expect.poll(() => observedBody).not.toBeNull();
    expect(JSON.parse(observedBody ?? '{}')).toEqual({ mode: 'test_master' });

    await expect(testMasterChip).toHaveAttribute('aria-pressed', 'true');
    await expect(page.getByTestId('role-chip-error')).toHaveCount(0);
  });

  test('surfaces an inline error and keeps the role when backend rejects', async ({ page }) => {
    await page.route('**/api/v1/mode', (route) => {
      const request = route.request();
      if (request.method() !== 'POST') {
        route.fallback();
        return;
      }
      route.fulfill({
        status: 403,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Forbidden',
          code: 'PERMISSION_DENIED',
          message: 'CGO + Linux required',
        }),
      });
    });

    await page.goto('/');

    const reflectorChip = page.getByTestId('role-chip-reflector');
    const testMasterChip = page.getByTestId('role-chip-test_master');

    // The app boots into reflector by default — try switching to
    // test_master so the request actually fires (clicking the
    // already-active chip is a no-op).
    await testMasterChip.click();
    await page.getByRole('button', { name: /switch role/i }).click();

    // The inline error appears with the backend reason.
    const errorTag = page.getByTestId('role-chip-error');
    await expect(errorTag).toBeVisible();
    await expect(errorTag).toContainText(/CGO \+ Linux required/i);

    // Local state did not change: reflector is still pressed.
    await expect(reflectorChip).toHaveAttribute('aria-pressed', 'true');
    await expect(testMasterChip).toHaveAttribute('aria-pressed', 'false');

    // Dismissing the error removes the tag.
    await page.getByTestId('role-chip-error-dismiss').click();
    await expect(errorTag).toHaveCount(0);
  });
});

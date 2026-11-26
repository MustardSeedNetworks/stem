import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { chooseRole, rail } from './helpers/rail';

/**
 * The role control lives in the rail since UI-STEM-9, and both rails are always
 * in the DOM — the phone drawer and the `hidden lg:flex` desktop one — so a
 * bare getByTestId matches two. Scope to the copy actually on screen (#941).
 */

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

    // The mode POST now carries an X-Csrf-Token, so the switch first fetches
    // one. Stubbed here for the same reason /api/v1/mode is: this spec is
    // about the chip's wiring, and the token endpoint shares the API rate
    // limiter with every other spec's traffic — leaving it live made the
    // switch fail with "Failed to fetch CSRF token: HTTP 429" under
    // parallelism rather than with the error the test is asserting.
    await page.route('**/api/v1/auth/csrf-token', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ token: 'e2e-csrf-token' }),
      });
    });

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

    // The role switch is the rail's product control.
    const roleSlot = rail(page, 'rail-product-slot');
    await expect(roleSlot).toBeVisible();

    // Choose test_master, then confirm the dialog.
    await chooseRole(page, 'Test Master');
    await page.getByTestId('confirm-dialog').getByRole('button', { name: 'Switch role' }).click();

    // Wait for the POST to have happened and the chip to reflect
    // the server's echoed mode.
    await expect.poll(() => observedBody).not.toBeNull();
    expect(JSON.parse(observedBody ?? '{}')).toEqual({ mode: 'test_master' });

    await expect(roleSlot).toContainText('Test Master');
    await expect(page.getByTestId('role-switch-error')).toHaveCount(0);
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

    const roleSlot = rail(page, 'rail-product-slot');

    // The app boots into reflector by default — try switching to
    // test_master so the request actually fires (clicking the
    // already-active chip is a no-op).
    await chooseRole(page, 'Test Master');
    await page.getByTestId('confirm-dialog').getByRole('button', { name: 'Switch role' }).click();

    // The inline error appears with the backend reason.
    const errorTag = page.getByTestId('role-switch-error');
    await expect(errorTag).toBeVisible();
    await expect(errorTag).toContainText(/CGO \+ Linux required/i);

    // Local state did not change: the slot still reads Reflector.
    await expect(roleSlot).toContainText('Reflector');

    // Dismissing the error removes the tag.
    await errorTag.getByRole('button', { name: /dismiss/i }).click();
    await expect(errorTag).toHaveCount(0);
  });
});

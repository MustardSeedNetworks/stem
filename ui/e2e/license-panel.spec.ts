import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * License panel — against the real daemon, with nothing stubbed.
 *
 * #1247: the panel called `/api/license*` (no `/v1`, so the SPA handler
 * answered index.html with HTTP 200), sent no `X-Csrf-Token` on the two POSTs
 * (403 from the CSRF manager) and posted `{key}` where the Go type declares
 * `licenseKey` (400 from the strict decoder). None of it was visible: every
 * failure was mapped to "connection failed".
 *
 * These deliberately do NOT `route.fulfill` the license endpoints. A mock
 * accepts a request whether or not it carries the CSRF header and whatever
 * path it is sent to, which is exactly how three breaks shipped at once
 * (the same lesson as #1080's role switch).
 *
 * Activation with a valid key is not exercised here: keys are Ed25519-signed
 * and the private key lives only in the keygen tool, so no test key exists to
 * sign one. What a real daemon can prove is that a typed key reaches the
 * activation handler and comes back with the daemon's own verdict — which is
 * the defect. The request/body/header contract is pinned in
 * ui/src/components/LicenseSection.request.test.tsx.
 */

test.describe('License panel', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  test('loads status from the versioned route', async ({ page }) => {
    const status = page.waitForResponse(
      (response) =>
        response.url().endsWith('/api/v1/license') && response.request().method() === 'GET',
    );
    await page.goto('/');
    await page.getByTestId('sidebar-settings-button').click();

    const response = await status;
    expect(response.status()).toBe(200);
    // index.html also comes back 200; the content type is what separates the
    // API from the SPA fallback the unversioned path used to hit.
    expect(response.headers()['content-type']).toContain('application/json');

    // "Loading…" is where the panel sat for the whole of #1247.
    const drawer = page.getByTestId('settings-drawer');
    await expect(drawer.getByText('Not Activated')).toBeVisible();
  });

  test('an activation attempt reaches the daemon with CSRF and is answered', async ({ page }) => {
    await page.goto('/');
    await page.getByTestId('sidebar-settings-button').click();
    const drawer = page.getByTestId('settings-drawer');
    await expect(drawer.getByText('Not Activated')).toBeVisible();

    await drawer.getByPlaceholder('MSN1.<payload>.<signature>').fill('MSN1.eyJ2IjoxfQ.c2ln');

    const activate = page.waitForResponse((response) =>
      response.url().endsWith('/api/v1/license/activate'),
    );
    await drawer.getByRole('button', { name: 'Activate License', exact: true }).click();

    const response = await activate;
    // 403 is the CSRF rejection, 404/HTML the SPA fallback. A 200 means the
    // handler itself answered — with a rejection of this unsigned key, which
    // is the correct answer and the one the operator must see.
    expect(response.status()).toBe(200);
    expect(response.request().headers()['x-csrf-token'] ?? '').not.toBe('');

    const body = (await response.json()) as { success: boolean; message: string };
    expect(body.success).toBe(false);
    expect(body.message).not.toBe('');

    // The daemon's verdict, not the blanket network message.
    await expect(drawer.getByText('Connection failed')).toBeHidden();
    await expect(drawer.getByText(body.message, { exact: false })).toBeVisible();
  });
});

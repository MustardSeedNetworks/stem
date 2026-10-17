import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Automatic update toggle — against the real daemon, nothing stubbed
 * (STM-21 acceptance: the Settings toggle round-trips). A fresh daemon reads
 * off; turning it on survives a reload, which only the file the daemon wrote
 * can explain; the test turns it off again so the shared daemon is left as it
 * was found.
 *
 * One daemon, and so one opt-in file, serves both browser projects in
 * parallel: two copies of this test would flip each other's switch. It runs
 * in Chromium only; the switch's rendering is covered by UpdateSection.test.
 */

test.describe('Automatic update toggle', () => {
  test.skip(
    ({ browserName }) => browserName !== 'chromium',
    'the opt-in is one file on the shared daemon',
  );

  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  test('round-trips through the daemon', async ({ page }) => {
    await page.goto('/');
    await page.getByTestId('sidebar-settings-button').click();
    const section = page.getByTestId('settings-update-section');
    const toggle = section.getByRole('switch', { name: 'Automatic updates' });

    await expect(toggle).toHaveAttribute('aria-checked', 'false');
    await expect(section.getByTestId('settings-update-source')).toHaveText('github');

    const saved = page.waitForResponse(
      (response) =>
        response.url().endsWith('/api/v1/update/settings') &&
        response.request().method() === 'POST',
    );
    await toggle.click();
    expect((await saved).status()).toBe(200);
    await expect(toggle).toHaveAttribute('aria-checked', 'true');

    await page.reload();
    await page.getByTestId('sidebar-settings-button').click();
    await expect(toggle).toHaveAttribute('aria-checked', 'true');

    await toggle.click();
    await expect(toggle).toHaveAttribute('aria-checked', 'false');
    await page.reload();
    await page.getByTestId('sidebar-settings-button').click();
    await expect(toggle).toHaveAttribute('aria-checked', 'false');
  });
});

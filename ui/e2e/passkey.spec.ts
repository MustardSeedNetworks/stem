import { expect, type Locator, type Page, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

async function addVirtualAuthenticator(page: Page): Promise<void> {
  const cdp = await page.context().newCDPSession(page);
  await cdp.send('WebAuthn.enable');
  await cdp.send('WebAuthn.addVirtualAuthenticator', {
    options: {
      protocol: 'ctap2',
      transport: 'internal',
      hasResidentKey: true,
      hasUserVerification: true,
      isUserVerified: true,
      automaticPresenceSimulation: true,
    },
  });
}

async function readPasskeyCount(count: Locator): Promise<number> {
  await expect(count).toHaveText(/Registered passkeys: \d+/);
  return Number((await count.textContent())?.match(/\d+/)?.[0]);
}

test.describe('Passkeys', () => {
  test('enrolls and signs in with a virtual authenticator', async ({ browserName, page }) => {
    test.skip(browserName !== 'chromium', 'CDP virtual authenticators are Chromium-only');
    await addVirtualAuthenticator(page);
    await skipSetupWizard(page);
    await page.goto('/account/security');
    const count = page.getByTestId('passkey-count');
    const initialCount = await readPasskeyCount(count);
    await page.getByRole('button', { name: 'Add a passkey' }).click();
    await expect(page.getByText('Passkey registered.')).toBeVisible();
    await expect(count).toHaveText(`Registered passkeys: ${initialCount + 1}`);

    await page.getByTestId('logout-button').click();
    await page.getByTestId('passkey-login').click();
    await expect(page.getByTestId('logout-button')).toBeVisible();
  });
});

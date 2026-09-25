import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

// An operator who types the daemon's address without https:// reaches the
// same port over plaintext. The listener answers that with a 308 to the https
// URL on the same host, port and path, and serves nothing else (#1296).
test.use({ storageState: { cookies: [], origins: [] } });

test('a plaintext address lands on the login page over https', async ({ page, baseURL }) => {
  if (!baseURL?.startsWith('https://')) {
    throw new Error(`the suite runs against the daemon's https origin, got ${baseURL}`);
  }
  const plaintext = `http://${new URL(baseURL).host}/`;

  await skipSetupWizard(page);
  const response = await page.goto(plaintext);

  expect(new URL(page.url()).protocol).toBe('https:');
  expect(new URL(page.url()).host).toBe(new URL(baseURL).host);
  expect(response?.request().redirectedFrom()?.url()).toBe(plaintext);
  await expect(page.getByTestId('login-title')).toBeVisible();
});

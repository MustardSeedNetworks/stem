import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * MEF Full Test reaches the daemon as a test it knows (#1708).
 *
 * The drawer offered it as `mef_full`, and the daemon calls the full suite
 * `mef`, so every run that included it failed with "Unknown or unsupported
 * test type" before any other check. The Go test
 * `TestUIOfferedTestsStartARun` holds every offered id to the run plan; this
 * drives the operator's path against the real daemon.
 *
 * The E2E daemon is deliberately unlicensed and MEF is Pro, so a known test
 * type is refused by the licence gate (402, naming `mef`). An unknown one is
 * refused earlier, as a 400. The licence refusal is the proof the daemon
 * recognised the test.
 */

const START_ENDPOINT = '**/api/v1/test/start';
const INTERFACES_ENDPOINT = '**/api/v1/interfaces';

/** Stubbed for the reason rfc2544-subtest-selection.spec.ts gives: the rate limiter. */
const STUB_INTERFACE = {
  name: 'e2e0',
  mac: '02:00:00:00:00:01',
  speed: 10000,
  duplex: 'full',
  state: 'up',
  driver: 'e2e',
  physical: true,
  xdp: false,
  score: 90,
  mtu: 1500,
  ipv4: '192.0.2.10',
  ipv6: '',
  usable: true,
};

const DEFAULT_SELECTION = [
  'rfc2544_throughput',
  'rfc2544_latency',
  'rfc2544_frame_loss',
  'rfc2544_back_to_back',
] as const;

test('MEF Full Test starts a run the daemon recognises', async ({ page }) => {
  await skipSetupWizard(page);
  await page.route(INTERFACES_ENDPOINT, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify([STUB_INTERFACE]),
    });
  });
  await useRole(page, 'test_master');
  await page.goto('/tests/servicetest');
  await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 10000 });

  await page.getByTestId('sidebar-settings-button').click();
  await expect(page.getByTestId('settings-drawer')).toBeVisible();

  // Only MEF Full Test, so the daemon's answer is about it alone.
  await page.getByTestId('rfc2544-test-section-toggle').click();
  for (const id of DEFAULT_SELECTION) {
    await page.getByTestId(`test-checkbox-${id}`).uncheck();
  }
  await page.getByTestId('mef-test-section-toggle').click();
  await page.getByTestId('test-checkbox-mef').check();

  // Escape dismisses the focused control's tooltip before its enclosing
  // drawer; clicking close races that tooltip on WebKit.
  await page.getByTestId('settings-drawer-close').focus();
  await expect(page.getByRole('tooltip')).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.getByRole('tooltip')).toBeHidden();
  await page.keyboard.press('Escape');
  await expect(page.getByTestId('settings-drawer')).toBeHidden();

  const iface = page.getByTestId('interface-select');
  await expect(iface.locator(`option[value="${STUB_INTERFACE.name}"]`)).toHaveCount(1);
  await iface.selectOption(STUB_INTERFACE.name);
  await page.getByTestId('peer-input').fill('192.0.2.20');

  const start = page.getByTestId('start-test-button');
  await expect(start).toBeEnabled();
  const answer = page.waitForResponse(START_ENDPOINT);
  await start.click();
  const response = await answer;

  const sent = response.request().postDataJSON() as { tests?: Array<{ testType: string }> };
  expect((sent.tests ?? []).map((step) => step.testType)).toEqual(['mef']);

  expect(response.status()).toBe(402);
  expect(await response.json()).toMatchObject({ requiredFeature: 'mef' });
});

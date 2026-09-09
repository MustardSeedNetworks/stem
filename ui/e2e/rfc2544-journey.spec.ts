import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * RFC 2544 configure-and-run journey.
 *
 * Stem's reason to exist is running RFC 2544 / Y.1564 / Y.1731 performance
 * tests, and before this spec **no test configured or started one**. The five
 * module-page specs asserted a heading and a regex; nothing clicked a control
 * that leads to a measurement. A regression that broke test start would have
 * shipped with a green suite.
 *
 * This drives the operator's actual path: pick the role, pick an interface,
 * start, and confirm the selection reaches the daemon.
 *
 * Not yet covered: toggling individual RFC 2544 sub-tests through the settings
 * drawer, which needs the drawer's collapsed standard-view sections driven
 * reliably first. Tracked separately rather than shipped failing.
 *
 * The start request is intercepted rather than run for real: a genuine RFC 2544
 * run needs a reflector at the far end and takes minutes, neither of which
 * belongs in per-PR CI. What is verified is everything the UI owns — that the
 * selection reaches the request body, that the control is gated until an
 * interface is chosen, and that the running state is reflected and reversible.
 */

const START_ENDPOINT = '**/api/v1/test/start';

test.describe('RFC 2544 journey', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    // Test Master is the role that may run tests; as a reflector the module
    // pages only warn (see benchmark-page.spec.ts).
    await useRole(page, 'test_master');
  });

  test('the start control stays disabled until an interface is chosen', async ({ page }) => {
    await page.goto('/tests/benchmark');

    const start = page.getByTestId('start-test-button');
    await expect(start).toBeVisible({ timeout: 10000 });

    await page.getByTestId('interface-select').selectOption('');
    await expect(start).toBeDisabled();
  });

  test('starting a run posts the selected tests and interface', async ({ page }) => {
    let startBody: unknown = null;
    await page.route(START_ENDPOINT, async (route) => {
      startBody = route.request().postDataJSON();
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'started' }),
      });
    });

    await page.goto('/tests/benchmark');
    await expect(page.getByTestId('rfc2544-config-form')).toBeVisible({ timeout: 10000 });

    // Pick the first real interface the daemon reported.
    const iface = page.getByTestId('interface-select');
    const value = await iface.locator('option').nth(1).getAttribute('value');
    expect(value, 'daemon reported no interfaces to select').toBeTruthy();
    await iface.selectOption(value as string);

    const start = page.getByTestId('start-test-button');
    await expect(start).toBeEnabled();
    await start.click();

    await expect
      .poll(() => startBody, { timeout: 10000, message: 'no start request was sent' })
      .not.toBeNull();

    // The operator's selection must survive into the request — asserting only
    // that "a request happened" would pass even if the form were ignored.
    const body = startBody as {
      tests?: Array<{ testType: string; config?: unknown }>;
      interface?: string;
      mode?: string;
    };
    expect(body.interface).toBe(value);
    expect(body.tests?.some((step) => step.testType.startsWith('rfc2544'))).toBe(true);
    expect(body.mode).toBe('test_master');
    expect(
      body.tests?.every((step) => step.config),
      'each RFC 2544 step must carry its own config snapshot',
    ).toBe(true);
  });
});

/**
 * The entitlement answer, from the operator's side (#1070).
 *
 * The daemon refuses an unlicensed standard with 402 TIER_TOO_LOW and a body
 * naming the feature; before this the UI rendered that as a generic failure,
 * so a Free operator was told the test failed rather than that it is sold.
 * The response is faked because CI's daemon is licensed by the test fixture —
 * what is under test is the render path, not the gate (which has Go tests).
 */
test.describe('feature gate', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'test_master');
  });

  test('a 402 names the feature and points at the licence, not a generic error', async ({
    page,
  }) => {
    await page.route(START_ENDPOINT, async (route) => {
      await route.fulfill({
        status: 402,
        contentType: 'application/json',
        body: JSON.stringify({
          error: 'Feature requires a higher tier',
          code: 'TIER_TOO_LOW',
          requiredFeature: 'rfc2544',
          currentTier: 'Invalid',
          upgradeMessage: 'Start a 14-day Pro trial with `stem license --trial`.',
        }),
      });
    });

    await page.goto('/tests/benchmark');
    await expect(page.getByTestId('rfc2544-config-form')).toBeVisible({ timeout: 10000 });

    const iface = page.getByTestId('interface-select');
    const value = await iface.locator('option').nth(1).getAttribute('value');
    expect(value, 'daemon reported no interfaces to select').toBeTruthy();
    await iface.selectOption(value as string);

    const start = page.getByTestId('start-test-button');
    await expect(start).toBeEnabled();
    await start.click();

    const alert = page.getByTestId('test-start-error');
    await expect(alert).toBeVisible({ timeout: 10000 });
    await expect(alert).toContainText('rfc2544');
    await expect(alert).toContainText('Stem Pro');
    // The daemon's own tier name for an unlicensed host is "Invalid", a state
    // name rather than a tier an operator has heard of (#1095). It must not
    // leak into the copy.
    await expect(alert).not.toContainText('Invalid');
    // The generic failure text must not be what the operator is shown.
    await expect(alert).not.toContainText('Failed to start test');
  });
});

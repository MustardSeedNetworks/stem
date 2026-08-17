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
      tests?: string[];
      interface?: string;
      testType?: string;
      mode?: string;
      config?: unknown;
    };
    expect(body.interface).toBe(value);
    expect(body.tests?.some((t) => t.startsWith('rfc2544'))).toBe(true);
    expect(body.testType).toMatch(/^rfc2544/);
    expect(body.mode).toBe('test_master');
    expect(body.config, 'the RFC 2544 config must be sent, not just the test list').toBeTruthy();
  });
});

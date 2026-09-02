import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * RFC 2544 sub-test selection, through the settings drawer.
 *
 * #637 covered the run journey — role, interface, start, assert the request
 * body — but stopped before the operator's actual first step: choosing which
 * sub-tests to run. `RFC2544ConfigForm` returns null unless a `rfc2544_*` test
 * is selected, so that choice controls whether the form exists at all.
 *
 * It was left out because the drawer opens on the `standard` view, where each
 * standard's tests sit inside a `CollapsibleSection` that starts collapsed,
 * and the section's disclosure control had no test id — the only way in was to
 * match its header by accessible name, which is the substring trap #663 is
 * about. The control now carries `${testId}-toggle`, so the section can be
 * driven directly.
 *
 * The default selection is the four `rfc2544_*` entries in `test-store.ts`.
 */

const START_ENDPOINT = '**/api/v1/test/start';
const INTERFACES_ENDPOINT = '**/api/v1/interfaces';

/**
 * The interface list is stubbed rather than read from the daemon.
 *
 * `/api/v1/interfaces` is behind the API rate limiter, and every worker shares
 * one client IP, so under parallelism a share of those calls come back 429 and
 * the picker renders with nothing but its placeholder. Measured on a local
 * four-worker run: **41 of 99** interface requests were rate-limited. A spec
 * that reads the live list is racing the limiter, and with `retries: 1` and a
 * flake budget of 0 in CI, losing that race once fails the build.
 *
 * Interface *discovery* is covered by rfc2544-journey.spec.ts. What this file
 * is about is which sub-tests reach the request, so the interface is a fixture
 * and the test is hermetic.
 */
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

test.describe('RFC 2544 sub-test selection', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.route(INTERFACES_ENDPOINT, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([STUB_INTERFACE]),
      });
    });
    await useRole(page, 'test_master');
    await page.goto('/tests/benchmark');
    await expect(page.getByTestId('rfc2544-config-form')).toBeVisible({ timeout: 10000 });

    await page.getByTestId('sidebar-settings-button').click();
    await expect(page.getByTestId('settings-drawer')).toBeVisible();
    await page.getByTestId('rfc2544-test-section-toggle').click();
    // The section's own checkboxes are the proof it expanded — asserting on
    // the toggle's aria-expanded would pass on an empty section.
    await expect(page.getByTestId('test-checkbox-rfc2544_throughput')).toBeVisible();
  });

  test('unchecking every rfc2544 test removes the config form, and re-checking one brings it back', async ({
    page,
  }) => {
    for (const id of DEFAULT_SELECTION) {
      await page.getByTestId(`test-checkbox-${id}`).uncheck();
    }

    // hasRFC2544Tests is the guard; with nothing selected the form must not
    // merely be hidden, it must not render.
    await expect(page.getByTestId('rfc2544-config-form')).toHaveCount(0);

    await page.getByTestId('test-checkbox-rfc2544_throughput').check();
    await expect(page.getByTestId('rfc2544-config-form')).toBeVisible();
  });

  test('the surviving selection is what reaches the start request', async ({ page }) => {
    // Reduce six-ish defaults to exactly one, so the assertion below cannot
    // pass by accident on a request that ignored the form.
    for (const id of DEFAULT_SELECTION) {
      if (id !== 'rfc2544_latency') {
        await page.getByTestId(`test-checkbox-${id}`).uncheck();
      }
    }
    await expect(page.getByTestId('test-checkbox-rfc2544_latency')).toBeChecked();

    let startBody: unknown = null;
    await page.route(START_ENDPOINT, async (route) => {
      startBody = route.request().postDataJSON();
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ status: 'started' }),
      });
    });

    // Close the drawer so it is not covering the start control.
    await page.keyboard.press('Escape');
    await expect(page.getByTestId('settings-drawer')).toBeHidden();

    const iface = page.getByTestId('interface-select');
    await expect(iface.locator(`option[value="${STUB_INTERFACE.name}"]`)).toHaveCount(1);
    await iface.selectOption(STUB_INTERFACE.name);

    const start = page.getByTestId('start-test-button');
    await expect(start).toBeEnabled();
    await start.click();

    await expect
      .poll(() => startBody, { timeout: 10000, message: 'no start request was sent' })
      .not.toBeNull();

    const body = startBody as { tests?: string[]; interface?: string };
    expect(body.interface).toBe(STUB_INTERFACE.name);
    const rfc2544 = (body.tests ?? []).filter((id) => id.startsWith('rfc2544'));
    expect(rfc2544).toEqual(['rfc2544_latency']);
  });
});

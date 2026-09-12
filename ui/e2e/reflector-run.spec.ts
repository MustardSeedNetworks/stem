import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * Start and stop a real reflector through the UI (#1080 clause 2, #1201).
 *
 * This is the one spec that runs the product rather than a frozen
 * precondition, and that is why it is opt-in:
 *
 *   1. **It needs a dataplane.** macOS and Windows builds are CGO-less, and so
 *      is CI's E2E daemon — `ci.yml` builds it with `CGO_ENABLED: '0'`. The
 *      daemon says so itself in `/api/v1/capabilities`, and the page renders
 *      that as the platform banner.
 *   2. **A run is exclusive daemon state.** One daemon serves the whole suite
 *      across two browser projects, so a reflector running here makes
 *      `stop-outcome.spec.ts`'s refused stop succeed, and its own start is
 *      refused for concurrent interface ownership. Both failed exactly that
 *      way when this lived in the same file.
 *
 * So it runs only when asked for, on its own:
 *
 *   STEM_E2E_REFLECTOR=1 ./scripts/run-e2e.sh --project=chromium \
 *     e2e/reflector-run.spec.ts
 *
 * on Linux with a CGO build and `cap_net_raw` on the binary (dev-srv-ubuntu is
 * the lab host). It asserts the daemon's own `testStatus`, not the button
 * state: treating a refused stop as a success is the defect this clause is
 * about, and starting the wrong thing entirely was #1201.
 */

const OPT_IN = process.env.STEM_E2E_REFLECTOR === '1';

async function daemonTestStatus(page: import('@playwright/test').Page): Promise<unknown> {
  const response = await page.request.get('/api/v1/stats');
  return ((await response.json()) as Record<string, unknown>).testStatus;
}

test.describe('reflector run', () => {
  test.skip(
    !OPT_IN,
    'set STEM_E2E_REFLECTOR=1 to run a real reflector; it owns the daemon for the duration',
  );

  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'reflector');
  });

  test('starts the reflector and a stop mid-run reaches stopped', async ({ page }) => {
    await page.goto('/reflector');

    const unsupported = await page
      .getByTestId('reflector-platform-banner')
      .isVisible()
      .catch(() => false);
    test.skip(unsupported, 'this daemon reports no reflector dataplane (a CGO-less build)');

    const start = page.getByTestId('reflector-start-button');
    await expect(start).toBeEnabled({ timeout: 10000 });
    await start.click();

    // Start means "reflect". It used to post the Tests page's selection, which
    // a Free daemon answered 402 for (#1201), so assert what is running and
    // not merely that something is.
    await expect.poll(() => daemonTestStatus(page), { timeout: 15000 }).toBe('running');
    const running = await page.request.get('/api/v1/stats');
    expect(((await running.json()) as Record<string, unknown>).currentTest).toBe('reflect');

    const stop = page.getByTestId('reflector-stop-button');
    await expect(stop).toBeVisible({ timeout: 10000 });
    await stop.click();

    // The daemon's state, not the client's opinion of it.
    await expect.poll(() => daemonTestStatus(page), { timeout: 15000 }).toBe('stopped');

    // And the operator is told, as a status rather than an alert.
    const message = page.getByTestId('test-stop-message');
    await expect(message).toBeVisible({ timeout: 10000 });
    await expect(message).toHaveAttribute('role', 'status');

    // The page returns to its resting state, offering to start again.
    await expect(start).toBeVisible();
  });
});

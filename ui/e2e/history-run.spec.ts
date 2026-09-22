import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * A finished run reaches the History page (#1333, plan row D-STEM-17).
 *
 * The page could never show a row: the recorder gated on `completedAt`, and
 * `/api/v1/test/result` carried no per-run timing at all. The daemon now keeps
 * the run's own clock, so this asserts the whole path — daemon stamps, shell
 * records, page lists — and it asserts the row against the daemon's timestamp
 * rather than against a browser clock, because that distinction is the row.
 *
 * It is opt-in for the same two reasons `reflector-run.spec.ts` is, and it is
 * the same run:
 *
 *   1. **It needs a real run, and therefore a dataplane.** macOS and Windows
 *      builds are CGO-less, and so is CI's E2E daemon (`ci.yml` builds it with
 *      `CGO_ENABLED: '0'`). The other way to a terminal run — starting a Pro
 *      standard — is 402 on the E2E daemon by design: nothing activates a
 *      licence or a trial, and `license-panel.spec.ts` asserts "Not Activated",
 *      so a spec that started one would break it.
 *   2. **A run is exclusive daemon state.** One daemon serves the whole suite,
 *      so a run started here changes what `stop-outcome.spec.ts` sees.
 *
 *   STEM_E2E_REFLECTOR=1 ./scripts/run-e2e.sh --project=chromium \
 *     e2e/history-run.spec.ts
 *
 * on Linux with a CGO build and `cap_net_raw` on the binary (an `msn-stem-cdev`
 * container or dev-srv-ubuntu).
 *
 * What a stopped reflector cannot prove is the metrics panel with a measured
 * verdict in it: that needs a peer at the far end. The projection from the
 * daemon's `data` onto the panel is covered by `src/stores/record-run.test.ts`.
 */

const OPT_IN = process.env.STEM_E2E_REFLECTOR === '1';

interface DaemonResult {
  status?: string;
  startedAt?: string;
  completedAt?: string;
  duration?: number;
}

async function daemonResult(page: import('@playwright/test').Page): Promise<DaemonResult> {
  const response = await page.request.get('/api/v1/test/result');
  return (await response.json()) as DaemonResult;
}

/** The value beside a label in the detail pane's facts list. */
function fact(
  page: import('@playwright/test').Page,
  label: string,
): import('@playwright/test').Locator {
  return page
    .locator('dl > div')
    .filter({ has: page.getByText(label, { exact: true }) })
    .locator('dd');
}

test.describe('history records a finished run', () => {
  test.skip(
    !OPT_IN,
    'set STEM_E2E_REFLECTOR=1 to run a real reflector; it owns the daemon for the duration',
  );

  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'reflector');
  });

  test('a stopped run appears in History with the daemon’s own timing', async ({ page }) => {
    await page.goto('/reflector');

    const unsupported = await page
      .getByTestId('reflector-platform-banner')
      .isVisible()
      .catch(() => false);
    test.skip(unsupported, 'this daemon reports no reflector dataplane (a CGO-less build)');

    await page.getByTestId('reflector-start-button').click();
    await expect
      .poll(async () => (await daemonResult(page)).status, { timeout: 15000 })
      .toBe('running');

    await page.getByTestId('reflector-stop-button').click();
    await expect
      .poll(async () => (await daemonResult(page)).status, { timeout: 15000 })
      .toBe('stopped');

    // The daemon's record of the run, which is what the row must carry.
    const result = await daemonResult(page);
    expect(result.startedAt).toBeTruthy();
    expect(result.completedAt).toBeTruthy();
    expect(result.duration).toBeGreaterThanOrEqual(0);

    await page.goto('/history');
    const row = page.getByTestId(`history-row-${result.completedAt}-reflect`);
    await expect(row).toBeVisible({ timeout: 10000 });

    // The detail pane reports the daemon's stamps, not the browser's. The id
    // above already pins `completedAt`; these pin that the pane reads all three.
    await row.click();
    await expect(fact(page, 'Started')).toHaveText(
      new Date(result.startedAt ?? '').toLocaleString(),
    );
    await expect(fact(page, 'Completed')).toHaveText(
      new Date(result.completedAt ?? '').toLocaleString(),
    );
    // "—" is what the page showed for the whole of #1333.
    await expect(fact(page, 'Duration')).not.toHaveText('—');
  });
});

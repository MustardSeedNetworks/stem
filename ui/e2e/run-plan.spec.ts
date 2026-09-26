import { expect, type Page, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * A two-step run plan, RFC 2544 throughput then Y.1564, run for real (#1078
 * clause 3, plan row ST-2).
 *
 * The Go test `TestRunPlanStopsAfterFailedStep` drives the plan over HTTP with
 * a stubbed executor. This drives it from the operator's side against a daemon
 * that measures: select both standards in the settings drawer, start, and
 * watch the daemon run the steps in order. The run plan is published on
 * `GET /api/v1/stats` (`steps[]`), which is what is asserted, beside the step
 * list the page renders from it.
 *
 * A step transition alone proves nothing: a zero-frame run once reported
 * `passed` (#1217) and an offered rate was once reported as measured (#1233).
 * So the first step must carry a measured, non-zero rate before the second is
 * accepted as running.
 *
 * Opt-in, for the reasons `reflector-run.spec.ts` gives, plus one: the daemon
 * needs a Professional licence (both standards are Pro, and the E2E daemon is
 * deliberately unlicensed), a CGO dataplane, and a reflector at the far end.
 *
 *   STEM_E2E_RUN_PLAN_INTERFACE=eth0 STEM_E2E_RUN_PLAN_PEER=172.18.0.2 \
 *   E2E_BASE_URL=https://localhost:18644 \
 *     npx playwright test --project=chromium e2e/run-plan.spec.ts
 *
 * The bench is two Linux containers on one bridge, each running
 * `scripts/e2e-daemon.sh`: one runs `stem reflect -i eth0`, the other started
 * a trial with `stem license --trial` before its daemon and publishes its port.
 * `docs/audits/ST-2_RUN_PLAN_E2E_2026-09-25.md` records the commands.
 */

const IFACE = process.env.STEM_E2E_RUN_PLAN_INTERFACE ?? '';
const PEER = process.env.STEM_E2E_RUN_PLAN_PEER ?? '';

interface Step {
  testType: string;
  status: string;
  result?: { data?: Record<string, unknown> };
}

interface DaemonStats {
  testStatus?: string;
  suiteId?: string;
  steps?: Step[];
}

async function daemonStats(page: Page): Promise<DaemonStats> {
  const response = await page.request.get('/api/v1/stats');
  return (await response.json()) as DaemonStats;
}

const TERMINAL = ['completed', 'error', 'cancelled', 'stopped'];

test.describe('run plan', () => {
  test.skip(
    !IFACE || !PEER,
    'set STEM_E2E_RUN_PLAN_INTERFACE and STEM_E2E_RUN_PLAN_PEER to run a real plan against a reflector',
  );

  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'test_master');
  });

  test('runs RFC 2544 throughput then Y.1564, in order, each measuring', async ({ page }) => {
    // A one-second-trial search plus four 2 s Y.1564 steps with their warmups.
    test.setTimeout(180_000);

    await page.goto('/tests/benchmark');
    await expect(page.getByTestId('rfc2544-config-form')).toBeVisible({ timeout: 10000 });

    await page.getByTestId('sidebar-settings-button').click();
    await expect(page.getByTestId('settings-drawer')).toBeVisible();

    // Frame sizes stay at their defaults: the daemon measures only the first
    // (#1464), and a frame-size change made after these edits would not reach
    // the run anyway (#1465). RFC 2544 starts at 64 bytes, Y.1564 at 128.
    //
    // RFC 2544: throughput only, one-second trials.
    await page.getByTestId('rfc2544-test-section-toggle').click();
    for (const id of ['rfc2544_latency', 'rfc2544_frame_loss', 'rfc2544_back_to_back']) {
      await page.getByTestId(`test-checkbox-${id}`).uncheck();
    }
    await expect(page.getByTestId('test-checkbox-rfc2544_throughput')).toBeChecked();
    const rfc2544 = page.getByTestId('rfc2544-config-form-drawer');
    await rfc2544.locator('#rfc2544-duration').fill('1');
    await rfc2544.locator('#rfc2544-trials').fill('1');

    // Y.1564: the configuration test at 10 Mbps, 2 s a step.
    await page.getByTestId('y1564-test-section-toggle').click();
    await page.getByTestId('test-checkbox-y1564_config').check();
    const y1564 = page.getByTestId('y1564-config-form');
    await y1564.locator('#y1564-cir').fill('10');
    await y1564.locator('#y1564-config-duration').fill('2');

    await page.getByTestId('settings-drawer-close').click();
    await expect(page.getByTestId('settings-drawer')).toBeHidden();

    await page.getByTestId('interface-select').selectOption(IFACE);
    await page.getByTestId('peer-input').fill(PEER);
    const start = page.getByTestId('start-test-button');
    await expect(start).toBeEnabled();
    await start.click();

    // The daemon's plan is the operator's selection, in selection order.
    await expect
      .poll(async () => (await daemonStats(page)).steps?.map((step) => step.testType), {
        timeout: 15000,
      })
      .toEqual(['rfc2544_throughput', 'y1564_config']);

    // The second step running after the first passed is the ordering claim.
    // Polled as one snapshot, so both statuses come from the same answer.
    let snapshot: DaemonStats = {};
    await expect
      .poll(
        async () => {
          snapshot = await daemonStats(page);
          return snapshot.steps?.map((step) => step.status);
        },
        { timeout: 120_000, intervals: [250] },
      )
      .toEqual(['passed', 'running']);
    expect(snapshot.testStatus).toBe('running');

    // ...and the step that passed measured something on the wire. With the
    // reflector stopped it still reports passed, at 0 pps (#1466), so this is
    // the assertion that fails on that bench, not the one above.
    const throughput = snapshot.steps?.[0]?.result?.data ?? {};
    expect(
      Number(throughput.MaxRatePPS),
      'RFC 2544 step reported no measured rate',
    ).toBeGreaterThan(0);

    // The page renders the same plan the daemon published.
    await expect(page.getByTestId('run-plan-step-0')).toHaveAttribute('data-status', 'passed');
    await expect(page.getByTestId('run-plan-step-1')).toHaveAttribute('data-status', 'running');

    // The second step finishes in the same run, having stepped through all four
    // rates with frames sent and reflected at each.
    let done: DaemonStats = {};
    await expect
      .poll(
        async () => {
          done = await daemonStats(page);
          return TERMINAL.includes(done.testStatus ?? '');
        },
        { timeout: 60_000 },
      )
      .toBe(true);
    expect(done.suiteId).toBe(snapshot.suiteId);
    expect(done.steps?.[0]?.status).toBe('passed');
    const y1564Result = done.steps?.[1]?.result?.data ?? {};
    const y1564Steps = (y1564Result.Steps ?? []) as Array<{
      OfferedRatePct: number;
      FramesTx: number;
      FramesRx: number;
    }>;
    expect(y1564Steps.map((step) => step.OfferedRatePct)).toEqual([25, 50, 75, 100]);
    for (const step of y1564Steps) {
      expect(step.FramesTx, `Y.1564 ${step.OfferedRatePct} % step sent nothing`).toBeGreaterThan(0);
      expect(
        step.FramesRx,
        `Y.1564 ${step.OfferedRatePct} % step got nothing back`,
      ).toBeGreaterThan(0);
    }

    // The step's verdict is the service's, not the run's (#1463): on the docker
    // bridge every step misses the default 5 ms FDV and the service fails, so
    // the plan ends in error. Asserted against ServicePass, so a bench quiet
    // enough to pass reads the same claim the other way round.
    expect(typeof y1564Result.ServicePass, 'Y.1564 result carries no verdict').toBe('boolean');
    const verdict = y1564Result.ServicePass ? 'passed' : 'failed';
    expect(done.steps?.[1]?.status).toBe(verdict);
    expect(done.testStatus).toBe(y1564Result.ServicePass ? 'completed' : 'error');
    // The live step list closes with the run; the results card keeps the plan.
    await expect(page.getByTestId('run-plan-result-1')).toHaveAttribute('data-status', verdict);
  });
});

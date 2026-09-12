import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * Stop is a request with an outcome (#1080).
 *
 * `handleStopTest` used to ignore `response.ok` and log any failure as
 * "non-critical", so a stop the daemon refused looked exactly like a stop it
 * performed: the spinner ended and nothing was said. The daemon answers
 * `400 INVALID_REQUEST` with "No test is currently running", and that sentence
 * has to reach the operator.
 *
 * **The stop request is real.** What is frozen is the *precondition*: only
 * `GET /api/v1/stats` is intercepted, and only its `testStatus` field is
 * rewritten to `running` — the rest of the payload is the daemon's own,
 * fetched per request. That is needed because both Stop controls render only
 * while polled status says a run is in flight, so the stale-poll window this
 * row is about (the run ended, another tab stopped it) is otherwise
 * unreachable from a browser. `POST /api/v1/test/stop` is never intercepted:
 * the 400 asserted below is the daemon's, and an E2E that faked it would be
 * mocking the thing under test (memory `niac-p1b2-editor-rewire`).
 *
 * "Stop mid-run reaches stopped" is the second test below, and it runs a real
 * reflector: `reflect` is Free and ungated (docs/EDITIONS.md §3) and
 * `handleTestStop` has a first-class reflector branch. It needs a daemon that
 * actually has a dataplane, so it skips where the daemon says it has none —
 * the macOS and Windows builds are CGO-less, and so is CI's E2E daemon, which
 * `ci.yml` builds with `CGO_ENABLED: '0'`. The daemon states this itself, in
 * `/api/v1/capabilities`, and the page renders that as the platform banner;
 * the skip reads the product's own answer rather than sniffing the OS.
 *
 * Where it does run — Linux with CGO, e.g. dev-srv-ubuntu with
 * `cap_net_raw` on the binary — it asserts the daemon's own `testStatus`, not
 * just the button state, because the client used to call a refused stop a
 * success. Every Pro standard would instead answer 402 anywhere: nothing in
 * `scripts/run-e2e.sh`, `ci.yml` or the fixtures licenses the E2E daemon.
 */

const STATS_ENDPOINT = '**/api/v1/stats';

/** Report a run in flight, leaving every other field the daemon's own. */
async function freezeStatusRunning(page: import('@playwright/test').Page): Promise<void> {
  await page.route(STATS_ENDPOINT, async (route) => {
    const response = await route.fetch();
    const body = (await response.json()) as Record<string, unknown>;
    await route.fulfill({ response, json: { ...body, testStatus: 'running' } });
  });
}

test.describe('stop outcome', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'test_master');
  });

  test.afterEach(async ({ page }) => {
    await page.unrouteAll({ behavior: 'wait' });
  });

  test('a refused stop shows the daemon’s own sentence', async ({ page }) => {
    await freezeStatusRunning(page);

    await page.goto('/tests/benchmark');

    const stop = page.getByTestId('stop-test-button');
    await expect(stop).toBeVisible({ timeout: 10000 });
    await stop.click();

    const message = page.getByTestId('test-stop-message');
    await expect(message).toBeVisible({ timeout: 10000 });
    // The daemon's text, not a client paraphrase and not the envelope's error
    // *type* ("Bad Request"), which is what the shared classifier used to read.
    await expect(message).toHaveText(/No test is currently running/);
    await expect(message).not.toHaveText(/Bad Request/);

    // A refusal is announced, not merely displayed.
    await expect(message).toHaveAttribute('role', 'alert');

    // The control comes back: `stopping` is not a terminal state.
    await expect(stop).toBeEnabled();
  });

  test('a stop mid-run reaches stopped', async ({ page }) => {
    // The reflector role, not the describe's test_master: this is the
    // reflector operator's own page and its Start control means "reflect".
    await useRole(page, 'reflector');
    await page.goto('/reflector');

    // The daemon's own answer about its dataplane, rendered by the page.
    const unsupported = await page
      .getByTestId('reflector-platform-banner')
      .isVisible()
      .catch(() => false);
    test.skip(
      unsupported,
      'this daemon reports no reflector dataplane (a CGO-less build); the clause needs Linux with CGO',
    );

    const start = page.getByTestId('reflector-start-button');
    await expect(start).toBeEnabled({ timeout: 10000 });
    await start.click();

    const stop = page.getByTestId('reflector-stop-button');
    await expect(stop).toBeVisible({ timeout: 15000 });

    // The reflector is genuinely up before it is stopped, or "stopped" would
    // mean nothing.
    await expect
      .poll(async () => (await (await page.request.get('/api/v1/stats')).json()).testStatus, {
        timeout: 15000,
      })
      .toBe('running');

    await stop.click();

    // The daemon's state, not the client's opinion of it: ignoring
    // `response.ok` is exactly the defect this row is about.
    await expect
      .poll(async () => (await (await page.request.get('/api/v1/stats')).json()).testStatus, {
        timeout: 15000,
      })
      .toBe('stopped');

    // And the operator is told, as a status rather than an alert.
    const message = page.getByTestId('test-stop-message');
    await expect(message).toBeVisible({ timeout: 10000 });
    await expect(message).toHaveAttribute('role', 'status');

    // The page returns to its resting state, offering to start again.
    await expect(start).toBeVisible();
  });
});

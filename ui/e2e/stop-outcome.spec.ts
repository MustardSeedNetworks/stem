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
 * Not covered here: "Stop mid-run reaches stopped". That needs the daemon
 * actually running something, which this host cannot do — the macOS and
 * Windows builds are CGO-less and have no dataplane, so a started run errors
 * within two seconds. It IS reachable on Linux with CGO: `reflect` is Free and
 * ungated (docs/EDITIONS.md §3) and `handleTestStop` has a first-class
 * reflector branch, so a reflector started on dev-srv-ubuntu and stopped
 * through this UI would close the clause. Every Pro standard would instead
 * answer 402 there: nothing in `scripts/run-e2e.sh`, `ci.yml` or the fixtures
 * licenses the E2E daemon. The transition itself is pinned by unit tests over
 * `resolveStopOutcome` and `StopOutcomeMessage`.
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
});

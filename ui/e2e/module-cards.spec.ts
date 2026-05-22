import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Module Pages Tests
 *
 * Validates that each test-module page (Benchmark, ServiceTest,
 * TrafficGen, Measure, Certify) listed in the Stem module architecture
 * renders with its module name. After the #66 redesign these moved
 * from "dashboard cards" to dedicated routes under /tests/*; the sidebar
 * has Test group nav links pointing at each.
 *
 * Uses skipSetupWizard() to skip the login modal — these tests don't
 * exercise the auth flow itself (see helpers/auth.ts).
 */

const MODULE_NAMES = ['benchmark', 'servicetest', 'trafficgen', 'measure', 'certify'] as const;

test.describe('Module Cards', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  test('should render each module page with its name visible', async ({ page }) => {
    // Navigate to /tests/<module> for each and assert the heading
    // carries the module name. Exercises both routing and the per-
    // module render — stronger than the old "any text visible
    // anywhere" check that broke when the dashboard cards were
    // removed in #66.
    for (const name of MODULE_NAMES) {
      await page.goto(`/tests/${name}`);
      await expect(
        page.getByRole('heading', { name: new RegExp(name, 'i') }).first(),
      ).toBeVisible();
    }
  });
});

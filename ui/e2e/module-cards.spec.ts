import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Module Pages Tests
 *
 * Validates that each test-module page (Benchmark, ServiceTest, TrafficGen,
 * Measure, Certify) listed in the Stem module architecture renders with its
 * module name. After the #66 redesign these moved from "dashboard cards" to
 * dedicated routes under /tests/*; the sidebar has Test group nav links
 * pointing at each.
 *
 * Uses skipSetupWizard() to skip the login modal — these tests don't exercise
 * the auth flow itself (see helpers/auth.ts).
 *
 * LOCATOR NOTE (#663). This used
 *   page.getByRole('heading', { name: new RegExp(name, 'i') }).first()
 * which is wrong twice over. Playwright matches accessible names by SUBSTRING
 * unless `exact: true`, and an unanchored regex is looser still — so a heading
 * merely containing "measure" satisfied the "Measure" case. `.first()` then
 * hid the ambiguity: when a second match appeared the test kept passing
 * against whichever came first, rather than reporting a strict-mode violation
 * at the point of the mistake. Across the fleet that surfaced later, on an
 * unrelated change, when a page header gained an "Open help for <page>" button
 * whose name contains the rail button's whole name.
 *
 * Now: the page title's own test id, and the exact expected string. The route
 * slug is not the title (`servicetest` -> "ServiceTest", `trafficgen` ->
 * "TrafficGen"), so the mapping is spelled out rather than derived — deriving
 * it is what pushed the old version toward a case-insensitive regex.
 */
const MODULE_PAGES = [
  { slug: 'benchmark', title: 'Benchmark' },
  { slug: 'servicetest', title: 'ServiceTest' },
  { slug: 'trafficgen', title: 'TrafficGen' },
  { slug: 'measure', title: 'Measure' },
  { slug: 'certify', title: 'Certify' },
] as const;

test.describe('Module Cards', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  test('should render each module page with its name visible', async ({ page }) => {
    for (const { slug, title } of MODULE_PAGES) {
      await page.goto(`/tests/${slug}`);

      const heading = page.getByTestId('page-header-title');
      // No .first(): a second page title on one page is a defect, and
      // strict mode should say so here rather than somewhere else later.
      await expect(heading).toBeVisible();
      await expect(heading).toHaveText(title);
    }
  });
});

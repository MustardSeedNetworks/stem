/**
 * page-identity.spec.ts — every route names itself (UI-STEM-4, stem#1260).
 *
 * `document.title` read "Stem | Mustard Seed Networks" on every route, so two
 * tabs or two bookmarks could not be told apart, and five of the six test
 * modules wore no eyebrow above their title. The title is asserted per route
 * because a single shared value is exactly the defect.
 */

import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/** Every `pageRegistry` route, with the tab title it must carry. */
const TITLES: Record<string, string> = {
  '/reflector': 'Reflector | Stem',
  '/tests/benchmark': 'Benchmark | Stem',
  '/tests/servicetest': 'ServiceTest | Stem',
  '/tests/trafficgen': 'TrafficGen | Stem',
  '/tests/measure': 'Measure | Stem',
  '/tests/certify': 'Certify | Stem',
  '/history': 'History | Stem',
  '/account/security': 'Security | Stem',
};

const MODULE_ROUTES = new Set([
  '/reflector',
  '/tests/benchmark',
  '/tests/servicetest',
  '/tests/trafficgen',
  '/tests/measure',
  '/tests/certify',
]);

test.describe('each route names itself', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  for (const [route, title] of Object.entries(TITLES)) {
    test(`${route} titles the tab and ${MODULE_ROUTES.has(route) ? 'wears' : 'omits'} the module eyebrow`, async ({
      page,
    }) => {
      await page.goto(route, { waitUntil: 'domcontentloaded' });
      await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 20000 });
      await expect(page).toHaveTitle(title);

      const eyebrow = page.getByTestId('page-header-eyebrow');
      if (MODULE_ROUTES.has(route)) {
        await expect(eyebrow).toHaveText('Test module');
      } else {
        await expect(eyebrow).toHaveCount(0);
      }
    });
  }

  test('the title follows client-side navigation', async ({ page }) => {
    await page.goto('/reflector', { waitUntil: 'domcontentloaded' });
    await expect(page).toHaveTitle(TITLES['/reflector']);

    await page.getByTestId('desktop-sidebar').getByRole('button', { name: 'History' }).click();
    await expect(page).toHaveTitle(TITLES['/history']);
  });
});

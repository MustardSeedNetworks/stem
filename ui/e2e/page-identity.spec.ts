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

const EYEBROWS: Record<string, string> = {
  '/reflector': 'Test module',
  '/tests/benchmark': 'Test module',
  '/tests/servicetest': 'Test module',
  '/tests/trafficgen': 'Test module',
  '/tests/measure': 'Test module',
  '/tests/certify': 'Test module',
  '/history': 'Results',
  '/account/security': 'Account',
};

test.describe('each route names itself', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  for (const [route, title] of Object.entries(TITLES)) {
    test(`${route} titles the tab and wears its kicker`, async ({ page }) => {
      await page.goto(route, { waitUntil: 'domcontentloaded' });
      await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 20000 });
      await expect(page).toHaveTitle(title);

      await expect(page.getByTestId('page-header-eyebrow')).toHaveText(EYEBROWS[route] ?? '');
      // The title is the nav label, and the trail names the page it ends on.
      await expect(page.getByRole('navigation', { name: 'Breadcrumb' })).toContainText(
        await page.getByTestId('page-header-title').innerText(),
      );
    });
  }

  test('the title follows client-side navigation', async ({ page }) => {
    await page.goto('/reflector', { waitUntil: 'domcontentloaded' });
    await expect(page).toHaveTitle(TITLES['/reflector']);

    await page.getByTestId('desktop-sidebar').getByRole('button', { name: 'History' }).click();
    await expect(page).toHaveTitle(TITLES['/history']);
  });
});

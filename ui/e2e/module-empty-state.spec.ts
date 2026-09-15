import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * Module empty state (#1257) E2E.
 *
 * `selectedTests` starts as the four RFC 2544 ids, so on a first visit every
 * module but Benchmark had each of its ConfigForms return null and the page
 * showed a title and nothing else. The per-page specs only asserted the header
 * AppShell renders, so an empty body satisfied them.
 *
 * These assert the body: each page offers an empty state whose action opens
 * the Settings drawer, and the results placeholder names the button the
 * toolbar actually has.
 */

const modules = [
  { name: 'ServiceTest', path: '/tests/servicetest' },
  { name: 'TrafficGen', path: '/tests/trafficgen' },
  { name: 'Measure', path: '/tests/measure' },
  { name: 'Certify', path: '/tests/certify' },
] as const;

test.describe('Module empty state', () => {
  for (const mod of modules) {
    test(`${mod.name} shows an empty state on a first visit`, async ({ page }) => {
      await skipSetupWizard(page);
      await useRole(page, 'test_master');
      await page.goto(mod.path);

      await expect(page.getByTestId('module-empty-state')).toBeVisible({ timeout: 10000 });
      await expect(page.getByTestId('module-empty-state-open-settings')).toBeVisible();
    });

    test(`${mod.name}'s empty state opens Settings`, async ({ page }) => {
      await skipSetupWizard(page);
      await useRole(page, 'test_master');
      await page.goto(mod.path);

      await page.getByTestId('module-empty-state-open-settings').click();
      await expect(page.getByTestId('settings-drawer')).toBeVisible({ timeout: 10000 });
    });
  }

  test('the results placeholder names the toolbar button', async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'test_master');
    await page.goto('/tests/benchmark');

    const runTest = await page.getByTestId('start-test-button').textContent();
    expect(runTest?.trim()).toBeTruthy();
    await expect(page.getByText(`click ${runTest?.trim()}.`)).toBeVisible({ timeout: 10000 });
  });
});

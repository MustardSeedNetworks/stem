/**
 * module-gate.spec.ts — a Pro module on Free says so before the run
 * (UI-STEM-15, stem#1283).
 *
 * A Free stem could configure a whole RFC 2544 run and only learn the module
 * was not included when Start came back 402. The module now renders its pitch
 * above the form, and the form below it is the preview: real, readable, inert.
 *
 * The suite's daemon runs unlicensed and its activation state is the run
 * directory's HOME (scripts/run-e2e.sh), so Free is the daemon's own answer.
 * Professional is driven by intercepting `/api/v1/license` with the payload
 * the Go handler produces — the same shape `LicenseStatus` generates the TS
 * type from, so a DTO change breaks the build rather than the assertion.
 */

import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { PRO_FEATURES, stubLicence } from './helpers/license';
import { useRole } from './helpers/role';

/**
 * The module routes the licence prices, and what each pitch must name.
 *
 * `configured` is whether the suite's default selection includes one of that
 * module's tests, which decides what the page body is: RFC 2544 is selected by
 * default, so Benchmark shows its form and the gate shows it as an inert
 * preview. Certify has nothing selected, so its body is the empty state, which
 * stays live under the pitch — its Settings link is how an operator would
 * configure the module in the first place.
 */
const GATED_MODULES = [
  { path: '/tests/benchmark', pitch: 'Benchmark is a Professional module', configured: true },
  { path: '/tests/certify', pitch: 'Certify is a Professional module', configured: false },
] as const;

test.describe('a Pro module on Free', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await useRole(page, 'test_master');
  });

  for (const { path, pitch, configured } of GATED_MODULES) {
    test(`${path} pitches the module before the run`, async ({ page }) => {
      // No stub: the suite's own unlicensed daemon answers.
      await page.goto(path, { waitUntil: 'domcontentloaded' });
      await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 20000 });

      await expect(page.getByTestId('module-gate-pitch')).toContainText(pitch, { timeout: 10000 });
      await expect(page.getByText('stem license trial')).toBeVisible();

      const preview = page.locator('[data-testid="module-gate"] [inert]');
      if (configured) {
        // The form is still there — it is what an operator reads to decide
        // whether to buy the module — and it is not empty.
        await expect(preview).toBeVisible();
        expect((await preview.innerText()).trim().length).toBeGreaterThan(0);
      } else {
        // No form to preview: the body is the empty state, and it stays live.
        await expect(preview).toHaveCount(0);
        await expect(page.getByTestId('module-empty-state-open-settings')).toBeEnabled();
      }
    });

    test(`${path} leaves the form unfillable but in the accessibility tree`, async ({ page }) => {
      test.skip(!configured, 'no test of this module is selected, so it has no form to preview');
      await page.goto(path, { waitUntil: 'domcontentloaded' });
      await expect(page.getByTestId('module-gate')).toBeVisible({ timeout: 20000 });

      const preview = page.locator('[data-testid="module-gate"] [inert]');
      // Readable: aria-hidden would take the whole pitch-worthy form away from
      // a screen reader, which is the opposite of the point.
      await expect(preview).not.toHaveAttribute('aria-hidden', 'true');
      await expect(preview).toHaveAttribute('aria-describedby', /module-gate-pitch/);

      // Unfillable: `inert` is what enforces it, and tabIndex would not show it.
      const focusable = await preview.locator('input, select, button, textarea').evaluateAll(
        (nodes) =>
          nodes.filter((node) => {
            (node as HTMLElement).focus();
            return document.activeElement === node;
          }).length,
      );
      expect(focusable).toBe(0);
    });

    test(`${path} is unchanged on Professional`, async ({ page }) => {
      await stubLicence(page, PRO_FEATURES);
      await page.goto(path, { waitUntil: 'domcontentloaded' });
      await expect(page.getByTestId('page-header-title')).toBeVisible({ timeout: 20000 });

      await expect(page.getByTestId('module-gate')).toHaveCount(0);
    });
  }

  test('the pitch does not strand an operator who has configured nothing', async ({ page }) => {
    // Certify's body on Free is the empty state, whose Settings link is how an
    // operator selects a test at all. Gating the whole page body would make it
    // inert and leave them with a pitch and no way to act on it.
    await page.goto('/tests/certify', { waitUntil: 'domcontentloaded' });
    await expect(page.getByTestId('module-gate-pitch')).toBeVisible({ timeout: 20000 });

    await page.getByTestId('module-empty-state-open-settings').click();
    await expect(page.getByTestId('settings-drawer')).toBeVisible({ timeout: 10000 });
  });

  test('the run-time 402 is still the backstop', async ({ page }) => {
    // The UI gate explains; the server decides. Start stays live on Free — a
    // stale or unreachable licence read must never block a paying operator —
    // and the daemon's own 402 is what stops the run.
    await page.goto('/tests/benchmark', { waitUntil: 'domcontentloaded' });
    await expect(page.getByTestId('module-gate')).toBeVisible({ timeout: 20000 });

    const iface = page.getByTestId('interface-select');
    const value = await iface.locator('option').nth(1).getAttribute('value');
    expect(value, 'daemon reported no interfaces to select').toBeTruthy();
    await iface.selectOption(value as string);
    await page.getByTestId('peer-input').fill('192.0.2.10');

    const start = page.getByTestId('start-test-button');
    await expect(start).toBeEnabled();
    await start.click();

    // The unlicensed daemon answers 402 and the page says so, exactly as it
    // did before this row — the pitch is an addition, not a replacement.
    await expect(page.getByTestId('test-start-error')).toBeVisible({ timeout: 10000 });
  });
});

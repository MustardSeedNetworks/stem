import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Help Drawer Smoke Test
 *
 * Locks the post-refactor behavior of the extracted, i18n'd help drawer:
 * - the sidebar footer help button opens the drawer
 * - the drawer surfaces real (non-trivial) content
 * - switching tabs (Tests -> Glossary) re-renders content
 * - the close button dismisses the drawer
 *
 * Selectors prefer the stable data-testids (help-drawer / help-drawer-close)
 * added recently so the assertions survive chrome-string localization. The
 * trigger is reached via its accessible name (sidebar footer "Open help"),
 * matching e2e/help-drawer.spec.ts.
 *
 * Uses skipSetupWizard() to skip the first-run wizard — this test does not
 * exercise the auth flow itself (see helpers/auth.ts).
 */

test.describe('Help Drawer (smoke)', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/');
  });

  test('opens, shows content, switches a tab, and closes', async ({ page }) => {
    // Open via the sidebar footer help button.
    await page.getByRole('button', { name: /open help/i }).click();

    const drawer = page.getByTestId('help-drawer');
    await expect(drawer).toBeVisible();

    // Real content renders (well beyond an empty shell).
    const initialText = await drawer.textContent();
    expect((initialText ?? '').length).toBeGreaterThan(100);

    // The default Tests tab lists a known standard.
    await expect(drawer.getByText(/RFC 2544/i).first()).toBeVisible();

    // Switch to the Glossary tab and confirm the view re-renders.
    await drawer.getByRole('button', { name: /^glossary$/i }).click();
    const glossaryText = await drawer.textContent();
    expect((glossaryText ?? '').length).toBeGreaterThan(100);

    // Close via the stable testid control.
    await page.getByTestId('help-drawer-close').click();
    await expect(drawer).not.toBeVisible();
  });
});

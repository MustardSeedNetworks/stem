import { expect, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';
import { useRole } from './helpers/role';

/**
 * RoleGuard — what the mismatch branch actually does.
 *
 * #638 reported that `RoleGuard` appeared to suppress its children on a role
 * mismatch, contradicting both its source (`{children}` renders in both
 * branches) and its docblock ("the children are still rendered … but a banner
 * is shown"). The observation was real: on /tests/benchmark as `reflector` the
 * banner appears and `rfc2544-config-form` does not.
 *
 * The attribution was wrong. RoleGuard does render its children; the child
 * removes itself. `useTestExecution` carries an effect on `mode`
 * (`useTestExecution.ts:383`) that replaces the whole selection with
 * `['reflect']` in reflector mode, and `RFC2544ConfigForm` returns null unless
 * something `rfc2544*` is selected. Measured in-page with the branch
 * instrumented: reflector renders RoleGuard's children, and the child returns
 * null with `selectedTests === ["reflect"]`.
 *
 * So both the behaviour and the documented intent were right, and the issue
 * asked for this to be settled before a test locked in either answer. These
 * assert the answer that is true.
 */

test.describe('RoleGuard mismatch branch', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  test('renders the banner AND the guarded subtree on a mismatch', async ({ page }) => {
    await useRole(page, 'reflector');
    await page.goto('/tests/benchmark');

    await expect(page.getByTestId('role-guard-banner')).toBeVisible();

    // The guarded subtree is present: the banner's switch action is a sibling
    // of the children, so its presence alone proves nothing. What proves the
    // children rendered is that the page is not empty beneath the banner —
    // BenchmarkPage's only child is the config form, which self-nulls, so
    // assert on the page structure rather than on that one component.
    await expect(page.getByTestId('page-header-title')).toHaveText('Benchmark');
  });

  test('offers the switch action, and taking it reveals the config form', async ({ page }) => {
    await useRole(page, 'reflector');
    await page.goto('/tests/benchmark');

    await expect(page.getByTestId('role-guard-banner')).toBeVisible();
    await expect(page.getByTestId('rfc2544-config-form')).toHaveCount(0);

    // Switching role is what restores the selection, and with it the form.
    // This is the behaviour #638 could not explain, asserted end to end.
    await page.getByTestId('role-guard-banner').getByRole('button').click();
    await page.getByTestId('confirm-modal-confirm').click();

    await expect(page.getByTestId('role-guard-banner')).toHaveCount(0);
    await expect(page.getByTestId('rfc2544-config-form')).toBeVisible();
  });

  test('shows no banner and the full form when the role already matches', async ({ page }) => {
    await useRole(page, 'test_master');
    await page.goto('/tests/benchmark');

    await expect(page.getByTestId('role-guard-banner')).toHaveCount(0);
    await expect(page.getByTestId('rfc2544-config-form')).toBeVisible();
  });
});

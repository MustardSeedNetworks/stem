import { expect, type Page, test } from '@playwright/test';
import { skipSetupWizard } from './helpers/auth';

/**
 * Responsive Design Tests
 *
 * Verify the SPA layout adapts across mobile / tablet / desktop without
 * horizontal overflow, with touch-target sizes that meet accessibility
 * minimums, and that the primary navigation remains usable.
 */

const viewports = {
  mobile: { width: 375, height: 667 },
  tablet: { width: 768, height: 1024 },
  desktop: { width: 1920, height: 1080 },
} as const;

const MIN_TOUCH_TARGET_PX = 32;
const MIN_READABLE_FONT_PX = 12;

async function expectNoHorizontalOverflow(page: Page): Promise<void> {
  const body = page.locator('body');
  const scrollWidth = await body.evaluate((el: HTMLElement) => el.scrollWidth);
  const clientWidth = await body.evaluate((el: HTMLElement) => el.clientWidth);
  // 10px tolerance for scrollbar/border rounding.
  expect(scrollWidth, 'page should not require horizontal scroll').toBeLessThanOrEqual(
    clientWidth + 10,
  );
}

async function expectTouchTargetsMeetMinimum(page: Page): Promise<void> {
  const buttons = page.locator('button');
  const count = await buttons.count();
  expect(count, 'page should render at least one button').toBeGreaterThan(0);

  // Sample up to 10 visible buttons — keeps the assertion fast while
  // catching any genuinely tiny target.
  for (let i = 0; i < Math.min(count, 10); i++) {
    const button = buttons.nth(i);
    if (!(await button.isVisible())) continue;
    const box = await button.boundingBox();
    if (!box) continue;
    expect(box.width).toBeGreaterThanOrEqual(MIN_TOUCH_TARGET_PX);
    expect(box.height).toBeGreaterThanOrEqual(MIN_TOUCH_TARGET_PX);
  }
}

test.describe('Responsive Design', () => {
  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
  });

  test('renders on mobile without horizontal overflow', async ({ page }) => {
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    await expect(page.getByTestId('page-header-title')).toBeVisible();

    await expectNoHorizontalOverflow(page);
    await expectTouchTargetsMeetMinimum(page);
  });

  test('renders on tablet with primary heading visible', async ({ page }) => {
    await page.setViewportSize(viewports.tablet);
    await page.goto('/');
    await expect(page.getByTestId('page-header-title')).toBeVisible();
    await expectTouchTargetsMeetMinimum(page);
  });

  test('renders on desktop with primary heading visible', async ({ page }) => {
    await page.setViewportSize(viewports.desktop);
    await page.goto('/');
    await expect(page.getByTestId('page-header-title')).toBeVisible();
  });

  test('layout reflows from mobile to desktop without losing the primary heading', async ({
    page,
  }) => {
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    const heading = page.getByTestId('page-header-title');
    await expect(heading).toBeVisible();

    await page.setViewportSize(viewports.desktop);
    // The heading should survive the resize. Playwright's expect.toBeVisible
    // auto-waits for layout to settle; no explicit transition delay needed.
    await expect(heading).toBeVisible();
    await expectNoHorizontalOverflow(page);
  });

  test('mobile body text meets minimum readable size', async ({ page }) => {
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    await expect(page.getByTestId('page-header-title')).toBeVisible();
    await expectTouchTargetsMeetMinimum(page);

    // Sample up to 5 visible text elements; flag anything below the
    // readable-font threshold.
    const textElements = page.locator('p, span, div');
    const count = await textElements.count();
    let sampled = 0;
    for (let i = 0; i < count && sampled < 5; i++) {
      const el = textElements.nth(i);
      if (!(await el.isVisible())) continue;
      const fontSize = await el.evaluate((node: Element) =>
        Number.parseFloat(getComputedStyle(node).fontSize),
      );
      expect(fontSize).toBeGreaterThanOrEqual(MIN_READABLE_FONT_PX);
      sampled++;
    }
    expect(
      sampled,
      'page should expose at least one visible text element to size-check',
    ).toBeGreaterThan(0);
  });
});

/**
 * Mobile navigation.
 *
 * #639 asked why no phone or tablet layout is exercised. Half of that was
 * already wrong — the tests above drive 375px and 768px viewports and assert
 * overflow, touch-target size and font size. What genuinely could not be
 * tested was the part the issue cares most about: "a control that is
 * unreachable behind a collapsed nav".
 *
 * Below the `lg` breakpoint the desktop rail is `display:none`, so every
 * existing spec's `sidebar-*` locator is unclickable, and the mobile drawer
 * that replaces it rendered with no test ids at all (`body(false)`). There was
 * no way in. The toggle now carries `mobile-nav-toggle` and the drawer
 * `mobile-sidebar`.
 *
 * These run with `hasTouch` so the taps are real touch events rather than
 * synthetic mouse clicks — a target that only responds to a mouse would pass
 * a click and fail a tap.
 */
test.describe('Mobile navigation', () => {
  test.use({ viewport: viewports.mobile, hasTouch: true, isMobile: true });

  test.beforeEach(async ({ page }) => {
    await skipSetupWizard(page);
    await page.goto('/');
    // Wait for the shell to finish hydrating before interacting. On webkit the
    // toggle was still being detached and re-attached when a tap landed
    // ("element was detached from the DOM, retrying"), which chromium never
    // showed. Settling on the first paint of real page content is the
    // difference between testing the app and testing its mount.
    await expect(page.getByTestId('page-header-title')).toBeVisible();
  });

  test('the desktop rail is not reachable at a phone width', async ({ page }) => {
    // The premise every other assertion here rests on. If this ever fails,
    // the mobile drawer is redundant and these tests are testing nothing.
    await expect(page.getByTestId('sidebar-settings-button')).toBeHidden();
  });

  test('the drawer opens on tap and exposes the module links', async ({ page }) => {
    const drawer = page.getByTestId('mobile-sidebar');
    const toggle = page.getByTestId('mobile-nav-toggle');

    await expect(toggle).toBeVisible();
    await toggle.tap();

    await expect(drawer).toBeInViewport();
    await expect(drawer.getByRole('button', { name: 'Benchmark', exact: true })).toBeVisible();
  });

  test('a module link in the drawer navigates', async ({ page }) => {
    await page.getByTestId('mobile-nav-toggle').tap();

    const drawer = page.getByTestId('mobile-sidebar');
    await drawer.getByRole('button', { name: 'Benchmark', exact: true }).tap();

    await expect(page).toHaveURL(/\/tests\/benchmark$/);
    await expect(page.getByTestId('page-header-title')).toHaveText('Benchmark');
  });

  test('the toggle is labelled for its current state, in the active locale', async ({ page }) => {
    const toggle = page.getByTestId('mobile-nav-toggle');

    // Both attributes were hardcoded English regardless of locale.
    await expect(toggle).toHaveAttribute('aria-label', 'Open menu');
    await toggle.tap();
    await expect(toggle).toHaveAttribute('aria-label', 'Close menu');
  });
});

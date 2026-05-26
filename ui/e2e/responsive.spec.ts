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
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible();

    await expectNoHorizontalOverflow(page);
    await expectTouchTargetsMeetMinimum(page);
  });

  test('renders on tablet with primary heading visible', async ({ page }) => {
    await page.setViewportSize(viewports.tablet);
    await page.goto('/');
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible();
    await expectTouchTargetsMeetMinimum(page);
  });

  test('renders on desktop with primary heading visible', async ({ page }) => {
    await page.setViewportSize(viewports.desktop);
    await page.goto('/');
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible();
  });

  test('layout reflows from mobile to desktop without losing the primary heading', async ({
    page,
  }) => {
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    const heading = page.getByRole('heading', { name: /reflector/i });
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
    await expect(page.getByRole('heading', { name: /reflector/i })).toBeVisible();
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

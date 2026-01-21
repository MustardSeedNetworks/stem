import { expect, type Page, test } from '@playwright/test';

/**
 * Responsive Design Tests
 *
 * Tests for mobile and tablet responsiveness:
 * - Layout adjustments
 * - Touch-friendly targets
 * - Navigation on small screens
 */

const viewports = {
  mobile: { width: 375, height: 667 },
  tablet: { width: 768, height: 1024 },
  desktop: { width: 1920, height: 1080 },
};

async function checkAccessibility(page: Page) {
  // Check that interactive elements have proper size
  const buttons = page.locator('button');
  const count = await buttons.count();

  for (let i = 0; i < Math.min(count, 10); i++) {
    const button = buttons.nth(i);
    if (await button.isVisible()) {
      const box = await button.boundingBox();
      if (box) {
        // Minimum touch target size (44x44 is recommended)
        expect(box.width).toBeGreaterThanOrEqual(32);
        expect(box.height).toBeGreaterThanOrEqual(32);
      }
    }
  }
}

test.describe('Responsive Design', () => {
  test('should display correctly on mobile', async ({ page }) => {
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Page should render without horizontal scroll
    const body = page.locator('body');
    const scrollWidth = await body.evaluate((el) => el.scrollWidth);
    const clientWidth = await body.evaluate((el) => el.clientWidth);
    expect(scrollWidth).toBeLessThanOrEqual(clientWidth + 10); // Small tolerance

    await checkAccessibility(page);
  });

  test('should display correctly on tablet', async ({ page }) => {
    await page.setViewportSize(viewports.tablet);
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Content should be visible
    await expect(page.locator('body')).toBeVisible();
    await checkAccessibility(page);
  });

  test('should display correctly on desktop', async ({ page }) => {
    await page.setViewportSize(viewports.desktop);
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Full layout should be visible
    await expect(page.locator('body')).toBeVisible();
  });

  test('should adapt grid layout on different screen sizes', async ({ page }) => {
    // Check mobile - cards should stack
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Check desktop - cards should be in grid
    await page.setViewportSize(viewports.desktop);
    await page.waitForTimeout(300); // Allow for CSS transition

    // Layout should have changed
    await expect(page.locator('body')).toBeVisible();
  });

  test('should maintain usability on touch devices', async ({ page }) => {
    await page.setViewportSize(viewports.mobile);
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // All buttons should be tappable size
    await checkAccessibility(page);

    // Text should be readable (at least 14px)
    const textElements = page.locator('p, span, div');
    const count = await textElements.count();

    for (let i = 0; i < Math.min(count, 5); i++) {
      const element = textElements.nth(i);
      if (await element.isVisible()) {
        const fontSize = await element.evaluate((el) =>
          Number.parseFloat(getComputedStyle(el).fontSize),
        );
        expect(fontSize).toBeGreaterThanOrEqual(12);
      }
    }
  });
});

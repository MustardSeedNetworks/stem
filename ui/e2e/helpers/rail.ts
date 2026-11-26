import type { Locator, Page } from '@playwright/test';

/**
 * The shared RailFooter's controls carry fixed `rail-*` test ids. Both rails
 * (the phone drawer and the desktop rail) are in the DOM at once, so a control
 * is scoped to the one a spec means: the desktop rail by default.
 */
export function rail(page: Page, testId: string): Locator {
  return page.getByTestId('desktop-sidebar').getByTestId(testId);
}

/** Choose the stem role from the footer's product menu. */
export async function chooseRole(page: Page, name: 'Reflector' | 'Test Master'): Promise<void> {
  await rail(page, 'rail-product-slot').click();
  await page.getByRole('menuitemradio', { name }).click();
}

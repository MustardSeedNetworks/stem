import type { Locator, Page } from '@playwright/test';

/**
 * The Settings drawer keeps test selection on its Tests tab, one foldable
 * section per standard. These helpers open that tab and a section, and find a
 * test's checkbox by the name the drawer shows for it.
 */
const TEST_NAMES: Record<string, string> = {
  rfc2544_throughput: 'Throughput',
  rfc2544_latency: 'Latency',
  rfc2544_frame_loss: 'Frame Loss',
  rfc2544_back_to_back: 'Back-to-Back',
  y1564_config: 'Configuration Test',
  mef: 'Full Test',
};

export async function openSettingsTab(page: Page, name: string): Promise<void> {
  await page.getByTestId('settings-drawer').getByRole('tab', { name }).click();
}

/** Opens a standard's foldable section on the Tests tab, e.g. `RFC 2544`. */
export async function expandTestSection(page: Page, standard: RegExp): Promise<void> {
  await page.getByTestId('settings-drawer').getByRole('button', { name: standard }).click();
}

export function testCheckbox(page: Page, id: string): Locator {
  const name = TEST_NAMES[id];
  if (!name) {
    throw new Error(`no checkbox name recorded for test id ${id}`);
  }
  return page.getByTestId('settings-drawer').getByRole('checkbox', { name, exact: true });
}

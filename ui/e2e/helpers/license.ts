/**
 * The suite's daemon runs unlicensed — its activation state is the run
 * directory's HOME (scripts/run-e2e.sh) — so every module the licence prices
 * renders behind <ModuleGate> as an inert preview. A spec that needs to drive
 * one of those forms says so here rather than working around the gate.
 *
 * The payload is the shape the Go handler produces, which is also what
 * LicenseStatus generates the TypeScript type from: a DTO change breaks the
 * build rather than quietly making the stub a lie.
 */

import type { Page } from '@playwright/test';

export const LICENSE_ENDPOINT = '**/api/v1/license';

/** Every feature stem sells, so no module is gated. */
export const PRO_FEATURES = [
  'reflector',
  'mef',
  'rfc2544',
  'rfc2889',
  'rfc6349',
  'trafficgen',
  'tsn',
  'y1564',
  'y1731',
];

/** Answer the licence read with `features`; an empty list is Free. */
export async function stubLicence(page: Page, features: string[]): Promise<void> {
  await page.route(LICENSE_ENDPOINT, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        activated: features.length > 0,
        isTrialMode: false,
        tier: features.length > 0 ? 2 : 0,
        tierName: features.length > 0 ? 'Professional' : 'Reflector',
        daysRemaining: 0,
        features,
        deviceHash: 'e2e',
      }),
    });
  });
}

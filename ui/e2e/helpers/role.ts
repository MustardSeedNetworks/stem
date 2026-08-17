import type { Page } from '@playwright/test';

/**
 * Role helpers for the module-page specs.
 *
 * RoleContext persists the active role to localStorage under `stem-role`
 * (RoleContext.ts ROLE_STORAGE_KEY) and reads it on mount, so seeding the key
 * before the app boots is enough to choose which RoleGuard branch renders.
 * Doing it via addInitScript rather than clicking the header RoleChip keeps
 * these specs independent of the chip's own behaviour, which
 * role-chip-backend.spec.ts already covers against the real /api/v1/mode call.
 */

export const ROLE_STORAGE_KEY = 'stem-role';

export type StemRole = 'reflector' | 'test_master';

/**
 * Seed the persisted role before the app boots. Must be called before
 * page.goto, like skipSetupWizard.
 */
export async function useRole(page: Page, role: StemRole): Promise<void> {
  await page.addInitScript(
    ([key, value]) => {
      window.localStorage.setItem(key, value);
    },
    [ROLE_STORAGE_KEY, role] as const,
  );
}

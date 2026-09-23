/**
 * Which licence feature each test module is sold under.
 *
 * The server prices a capability in `featuresByTestType()`
 * (`internal/api/features.go`), which its own comment calls the single place a
 * capability is priced. This is the UI's view of the same fact, and
 * `scripts/check-module-feature-parity.py` holds the two together: a module
 * may not name a feature the server does not charge for.
 *
 * A module lists every feature it can run. It is gated only when the licence
 * grants NONE of them — Certify covers three standards, and holding one of the
 * three should show the form, not a pitch for what is already paid for.
 */

export interface ModuleFeatures {
  /** Feature ids from internal/license/policy.go. */
  features: readonly string[];
  /** Locale key under `license.gated.modules.*` for the name and the pitch. */
  i18nKey: string;
}

export const MODULE_FEATURES = {
  '/tests/benchmark': { features: ['rfc2544'], i18nKey: 'benchmark' },
  '/tests/servicetest': { features: ['y1564', 'mef'], i18nKey: 'serviceTest' },
  '/tests/trafficgen': { features: ['trafficgen'], i18nKey: 'trafficGen' },
  '/tests/measure': { features: ['y1731'], i18nKey: 'measure' },
  '/tests/certify': { features: ['rfc2889', 'rfc6349', 'tsn'], i18nKey: 'certify' },
} as const satisfies Record<string, ModuleFeatures>;

export type GatedModulePath = keyof typeof MODULE_FEATURES;

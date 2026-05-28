/**
 * i18n TypeScript Types
 *
 * Provides type-safe translation keys for react-i18next.
 * These types are generated from the English locale files.
 *
 * Usage:
 * ```tsx
 * import { useTranslation } from 'react-i18next';
 *
 * function MyComponent() {
 *   const { t } = useTranslation('common');
 *   // TypeScript will autocomplete 'buttons.save', 'status.connected', etc.
 *   return <button>{t('buttons.save')}</button>;
 * }
 * ```
 */

import type enCli from '@locales/en/cli.json';
import type enCommon from '@locales/en/common.json';
import type enErrors from '@locales/en/errors.json';
import type enHelp from '@locales/en/help.json';
import type enModules from '@locales/en/modules.json';
import type enParams from '@locales/en/params.json';
import type enRecovery from '@locales/en/recovery.json';
import type enSecurity from '@locales/en/security.json';
import type enSettings from '@locales/en/settings.json';
import type enSetup from '@locales/en/setup.json';

/**
 * Type definitions for each namespace.
 */
export type CommonTranslations = typeof enCommon;
export type ErrorsTranslations = typeof enErrors;
export type ModulesTranslations = typeof enModules;
export type RecoveryTranslations = typeof enRecovery;
export type SecurityTranslations = typeof enSecurity;
export type SettingsTranslations = typeof enSettings;
export type SetupTranslations = typeof enSetup;
export type CliTranslations = typeof enCli;
export type ParamsTranslations = typeof enParams;
export type HelpTranslations = typeof enHelp;

/**
 * All translations combined.
 */
export interface Translations {
  common: CommonTranslations;
  errors: ErrorsTranslations;
  modules: ModulesTranslations;
  recovery: RecoveryTranslations;
  security: SecurityTranslations;
  settings: SettingsTranslations;
  setup: SetupTranslations;
  cli: CliTranslations;
  params: ParamsTranslations;
  help: HelpTranslations;
}

/**
 * Declaration merging for react-i18next.
 * This enables autocomplete for translation keys.
 */
declare module 'i18next' {
  interface CustomTypeOptions {
    defaultNS: 'common';
    resources: Translations;
  }
}

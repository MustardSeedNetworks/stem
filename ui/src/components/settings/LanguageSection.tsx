/**
 * LanguageSection — the Settings control for the interface language.
 *
 * The choice persists to localStorage ('stem-language') through i18next's
 * language detector, so it survives a reload.
 */
import { Languages } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { CollapsibleSection } from '../CollapsibleSection';
import { LanguageSwitcher } from '../LanguageSwitcher';

export function LanguageSection(): ReactElement {
  const { t } = useTranslation('common');
  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Languages className="w-4 h-4" aria-hidden="true" />
          <span>{t('language.title')}</span>
        </div>
      }
      defaultOpen={true}
    >
      <LanguageSwitcher className="w-full rounded-lg border border-surface-border bg-surface-base px-3 py-row text-sm text-text-primary" />
    </CollapsibleSection>
  );
}

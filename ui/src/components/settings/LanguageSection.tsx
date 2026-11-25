/**
 * LanguageSection — the Settings control for the interface language.
 *
 * The choice persists to localStorage ('stem-language') through i18next's
 * language detector, so it survives a reload.
 */
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { languages } from '../../i18n';
import { Card } from '../../shared/Card';
import { FormField } from '../../shared/FormField';
import { Select } from '../../shared/Select';

export function LanguageSection(): ReactElement {
  const { t, i18n } = useTranslation('common');
  return (
    <Card>
      <FormField label={t('language.title')}>
        {(control) => (
          <Select
            {...control}
            value={i18n.resolvedLanguage ?? i18n.language}
            onValueChange={(code): void => {
              i18n.changeLanguage(code).catch(() => {
                // Language change failure is handled gracefully by i18next
              });
            }}
            options={languages.map((lang) => ({ value: lang.code, label: lang.nativeLabel }))}
          />
        )}
      </FormField>
    </Card>
  );
}

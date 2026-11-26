/**
 * AppearanceSection — the Settings control for the colour theme.
 *
 * Three choices: follow the OS (the default), light, dark. The rail's quick
 * toggle writes the same store, so the two always agree.
 */
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { type Theme, useTheme } from '../../hooks/useTheme';
import { Card } from '../../shared/Card';
import { SegmentedControl } from '../../shared/SegmentedControl';

export function AppearanceSection(): ReactElement {
  const { t } = useTranslation('settings');
  const { theme, setTheme } = useTheme();
  return (
    <Card title={t('general.theme')} data-testid="settings-appearance">
      <SegmentedControl
        label={t('general.theme')}
        value={theme}
        onChange={(value): void => setTheme(value as Theme)}
        options={[
          { value: 'system', label: t('general.themeSystem') },
          { value: 'light', label: t('general.themeLight') },
          { value: 'dark', label: t('general.themeDark') },
        ]}
      />
    </Card>
  );
}

/**
 * Y1731Section Component
 *
 * ITU-T Y.1731 Ethernet OAM tests.
 * Includes: Delay (DMM/DMR), Loss (LMM/LMR), Synthetic Loss, Loopback.
 */

import { Radio } from 'lucide-react';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { CollapsibleSection } from '../../CollapsibleSection';
import { type Y1731Config, Y1731ConfigForm } from '../../Y1731ConfigForm';
import { TestCheckbox } from '../TestCheckbox';
import type { TestDefinition, TestSectionProps } from '../types';

interface Y1731SectionProps extends TestSectionProps {
  config: Y1731Config;
  onConfigChange: (config: Y1731Config) => void;
}

/** Test ID to translation key mapping - keys match backend test identifiers */
const testKeyMap: Map<string, string> = new Map([
  ['y1731_delay', 'delay'],
  ['y1731_loss', 'loss'],
  ['y1731_slm', 'slm'],
  ['y1731_loopback', 'loopback'],
]);

const testIds: string[] = [...testKeyMap.keys()];

export function Y1731Section({
  selectedTests,
  onToggleTest,
  config,
  onConfigChange,
}: Y1731SectionProps): React.JSX.Element {
  const { t } = useTranslation('settings');

  const tests: TestDefinition[] = useMemo(
    () =>
      testIds.map((id) => {
        const key = testKeyMap.get(id);
        return {
          id,
          name: t(`tests.y1731.${key}.name` as never),
          desc: t(`tests.y1731.${key}.desc` as never),
          tooltip: t(`tests.y1731.${key}.tooltip` as never),
        };
      }),
    [t],
  );

  const selectedCount = useMemo(
    () => selectedTests.filter((test) => test.startsWith('y1731')).length,
    [selectedTests],
  );

  const hasSelectedTests = selectedCount > 0;

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Radio className="w-4 h-4" aria-hidden="true" />
          <span>{t('settings.tests.y1731.title', 'Y.1731 OAM')}</span>
          <span className="caption text-text-muted">({selectedCount}/4)</span>
        </div>
      }
    >
      <div className="space-y-4">
        {/* Test Selection */}
        <div className="space-y-2">
          {tests.map((test) => (
            <TestCheckbox
              key={test.id}
              test={test}
              checked={selectedTests.includes(test.id)}
              onChange={(): void => onToggleTest(test.id)}
            />
          ))}
        </div>

        {/* Configuration Form */}
        {hasSelectedTests && (
          <div className="border-t border-surface-border pt-4">
            <Y1731ConfigForm
              config={config}
              setConfig={onConfigChange}
              selectedTests={selectedTests}
            />
          </div>
        )}
      </div>
    </CollapsibleSection>
  );
}

export default Y1731Section;

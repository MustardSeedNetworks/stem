/**
 * MEFSection Component
 *
 * MEF Carrier Ethernet service validation tests.
 * Includes: Configuration, Performance, Full Test.
 * Note: Uses same config as Y.1564 for service parameters.
 */

import { Settings2 } from 'lucide-react';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { CollapsibleSection } from '../../CollapsibleSection';
import { TestCheckbox } from '../TestCheckbox';
import type { TestDefinition, TestSectionProps } from '../types';

/** Test ID to translation key mapping - keys match backend test identifiers */
const testKeyMap: Map<string, string> = new Map([
  ['mef_config', 'config'],
  ['mef_perf', 'performance'],
  ['mef_full', 'full'],
]);

const testIds: string[] = [...testKeyMap.keys()];

export function MEFSection({ selectedTests, onToggleTest }: TestSectionProps): React.JSX.Element {
  const { t } = useTranslation('settings');

  const tests: TestDefinition[] = useMemo(
    () =>
      testIds.map((id) => {
        const key = testKeyMap.get(id);
        return {
          id,
          name: t(`tests.mef.${key}.name` as never),
          desc: t(`tests.mef.${key}.desc` as never),
          tooltip: t(`tests.mef.${key}.tooltip` as never),
        };
      }),
    [t],
  );

  const selectedCount = useMemo(
    () => selectedTests.filter((test) => test.startsWith('mef')).length,
    [selectedTests],
  );

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-compact">
          <Settings2 className="w-4 h-4" aria-hidden="true" />
          <span>{t('settings:tests.mef.title', 'MEF Service')}</span>
          <span className="caption text-text-muted">({selectedCount}/3)</span>
        </div>
      }
    >
      <div className="stack-sm">
        {tests.map((test) => (
          <TestCheckbox
            key={test.id}
            test={test}
            checked={selectedTests.includes(test.id)}
            onChange={(): void => onToggleTest(test.id)}
          />
        ))}
      </div>
      {/* Note: MEF uses same config as Y.1564 for service parameters */}
    </CollapsibleSection>
  );
}

export default MEFSection;

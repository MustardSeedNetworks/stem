/**
 * TSNSection Component
 *
 * IEEE 802.1Qbv Time-Sensitive Networking tests.
 * Includes: Gate Timing, Traffic Isolation, Scheduled Latency, Full Suite.
 */

import { Cpu } from 'lucide-react';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { CollapsibleSection } from '../../CollapsibleSection';
import { type TSNConfig, TSNConfigForm } from '../../TSNConfigForm';
import { TestCheckbox } from '../TestCheckbox';
import type { TestDefinition, TestSectionProps } from '../types';

interface TSNSectionProps extends TestSectionProps {
  config: TSNConfig;
  onConfigChange: (config: TSNConfig) => void;
}

/** Test ID to translation key mapping - keys match backend test identifiers */
const testKeyMap: Map<string, string> = new Map([
  ['tsn_timing', 'timing'],
  ['tsn_isolation', 'isolation'],
  ['tsn_latency', 'latency'],
  ['tsn_full', 'full'],
]);

const testIds: string[] = [...testKeyMap.keys()];

export function TSNSection({
  selectedTests,
  onToggleTest,
  config,
  onConfigChange,
}: TSNSectionProps): React.JSX.Element {
  const { t } = useTranslation('settings');

  const tests: TestDefinition[] = useMemo(
    () =>
      testIds.map((id) => {
        const key = testKeyMap.get(id);
        return {
          id,
          name: t(`tests.tsn.${key}.name` as never),
          desc: t(`tests.tsn.${key}.desc` as never),
          tooltip: t(`tests.tsn.${key}.tooltip` as never),
        };
      }),
    [t],
  );

  const selectedCount = useMemo(
    () => selectedTests.filter((test) => test.startsWith('tsn')).length,
    [selectedTests],
  );

  const hasSelectedTests = selectedCount > 0;

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Cpu className="w-4 h-4" aria-hidden="true" />
          <span>{t('settings.tests.tsn.title', 'TSN 802.1Qbv')}</span>
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
            <TSNConfigForm
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

export default TSNSection;

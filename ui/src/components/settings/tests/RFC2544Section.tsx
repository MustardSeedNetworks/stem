// Copyright (c) 2025 Mustard Seed Networks. All rights reserved.

/**
 * RFC2544Section Component
 *
 * RFC 2544 network benchmarking test selection and configuration.
 * Includes: Throughput, Latency, Frame Loss, Back-to-Back, System Recovery, Reset.
 */

import { Zap } from 'lucide-react';
import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { CollapsibleSection } from '../../CollapsibleSection';
import { type RFC2544Config, RFC2544ConfigForm } from '../../RFC2544ConfigForm';
import { TestCheckbox } from '../TestCheckbox';
import type { TestDefinition, TestSectionProps } from '../types';

interface RFC2544SectionProps extends TestSectionProps {
  config: RFC2544Config;
  onConfigChange: (config: RFC2544Config) => void;
}

/** Test ID to translation key mapping - keys match backend test identifiers */
const testKeyMap: Map<string, string> = new Map([
  ['rfc2544_throughput', 'throughput'],
  ['rfc2544_latency', 'latency'],
  ['rfc2544_frame_loss', 'frameLoss'],
  ['rfc2544_back_to_back', 'backToBack'],
  ['rfc2544_system_recovery', 'systemRecovery'],
  ['rfc2544_reset', 'reset'],
]);

const testIds: string[] = [...testKeyMap.keys()];

export function RFC2544Section({
  selectedTests,
  onToggleTest,
  config,
  onConfigChange,
}: RFC2544SectionProps): React.JSX.Element {
  const { t } = useTranslation('settings');

  const tests: TestDefinition[] = useMemo(
    () =>
      testIds.map((id) => {
        const key = testKeyMap.get(id);
        return {
          id,
          name: t(`tests.rfc2544.${key}.name`),
          desc: t(`tests.rfc2544.${key}.desc`),
          tooltip: t(`tests.rfc2544.${key}.tooltip`),
        };
      }),
    [t],
  );

  const selectedCount = useMemo(
    () => selectedTests.filter((test) => test.startsWith('rfc2544')).length,
    [selectedTests],
  );

  const hasSelectedTests = selectedCount > 0;

  return (
    <CollapsibleSection
      title={
        <div className="flex items-center gap-2">
          <Zap className="w-4 h-4" aria-hidden="true" />
          <span>{t('settings.tests.rfc2544.title', 'RFC 2544 Tests')}</span>
          <span className="caption text-text-muted">({selectedCount}/6)</span>
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
            <RFC2544ConfigForm
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

export default RFC2544Section;

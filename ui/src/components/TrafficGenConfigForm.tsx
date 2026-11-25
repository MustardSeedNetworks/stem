/**
 * @fileoverview Stem - Traffic Generator Configuration
 * @description Migrated to react-hook-form + valibot per #325. Offers only
 *              what the generator applies: no burst, VLAN or MAC settings
 *              (#1744).
 */

import { Radio } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { FRAME_SIZE_OPTIONS } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { hasGroupTests } from '../lib/testGroups';
import { TrafficGenConfigSchema } from '../schemas/configs';
import { FormField } from '../shared/FormField';
import { Input } from '../shared/Input';
import { CollapsibleSection } from './CollapsibleSection';
import { NumberSelectField } from './ConfigFormFields';
import { FormSection } from './FormSection';
import { TestSummary } from './TestSummary';

/** Traffic generator configuration parameters */
export interface TrafficGenConfig {
  frameSize: number;
  ratePct: number;
  duration: number;
  warmup: number;
  streamId: number;
}

/** Default traffic generator configuration */
export const defaultTrafficGenConfig: TrafficGenConfig = {
  frameSize: 128,
  ratePct: 100,
  duration: 60,
  warmup: 2,
  streamId: 1,
};

const RATE_PRESETS: Array<{ value: number; label: string }> = [
  { value: 10, label: '10%' },
  { value: 25, label: '25%' },
  { value: 50, label: '50%' },
  { value: 75, label: '75%' },
  { value: 90, label: '90%' },
  { value: 100, label: '100%' },
];

interface TrafficGenConfigFormProps {
  config: TrafficGenConfig;
  setConfig: (config: TrafficGenConfig) => void;
  selectedTests: string[];
}

export function TrafficGenConfigForm({
  config,
  setConfig,
  selectedTests,
}: TrafficGenConfigFormProps): ReactElement | null {
  const hasTrafficGenTests = hasGroupTests('trafficgen', selectedTests);

  const form = useConfigForm<TrafficGenConfig>({
    schema: TrafficGenConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    control,
    watch,
    setValue,
    formState: { errors },
  } = form;
  const { t } = useTranslation(['settings', 'help']);

  if (!hasTrafficGenTests) {
    return null;
  }

  const frameSize = watch('frameSize') ?? 0;
  const ratePct = watch('ratePct') ?? 0;
  const duration = watch('duration') ?? 0;
  const warmup = watch('warmup') ?? 0;

  const calculateThroughput = (): string => {
    const lineRateMbps = 10000;
    const throughputMbps = (lineRateMbps * ratePct) / 100;
    if (throughputMbps >= 1000) return `${(throughputMbps / 1000).toFixed(1)} Gbps`;
    return `${throughputMbps.toFixed(0)} Mbps`;
  };

  return (
    <CollapsibleSection
      testId="trafficgen-config-form"
      title={
        <div className="flex items-center gap-compact">
          <Radio className="w-4 h-4" />
          <span>{t('testConfig.trafficgen.title')}</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="stack-lg">
        <FormSection title={t('testConfig.trafficgen.traffic.title')}>
          <div className="grid grid-cols-2 gap-default">
            <NumberSelectField
              control={control}
              name="frameSize"
              label={t('testConfig.trafficgen.traffic.frameSize')}
              hint={t('help:tooltips.testConfig.trafficgen.traffic.frameSizeHelp')}
              error={errors.frameSize?.message}
              options={FRAME_SIZE_OPTIONS.map((option) => ({
                value: option.value,
                label: t(`testConfig.common.${option.qualifier}`, { size: option.value }),
              }))}
            />

            <div>
              <FormField
                label={t('testConfig.trafficgen.traffic.rate')}
                hint={t('help:tooltips.testConfig.trafficgen.traffic.rateHelp')}
                error={errors.ratePct?.message}
              >
                {(control) => (
                  <Input
                    {...control}
                    type="number"
                    step={0.01}
                    {...register('ratePct', { valueAsNumber: true })}
                  />
                )}
              </FormField>
              <div className="mt-tight flex gap-tight flex-wrap">
                {RATE_PRESETS.map((preset) => (
                  <button
                    key={preset.value}
                    type="button"
                    onClick={() =>
                      setValue('ratePct', preset.value, {
                        shouldValidate: true,
                        shouldDirty: true,
                      })
                    }
                    className={`text-xs px-cell py-0.5 rounded border ${
                      ratePct === preset.value
                        ? 'bg-brand-primary text-on-brand border-brand-primary'
                        : 'bg-surface-base border-surface-border text-text-muted'
                    }`}
                  >
                    {preset.label}
                  </button>
                ))}
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-default">
            <FormField
              label={t('testConfig.trafficgen.traffic.duration')}
              hint={t('help:tooltips.testConfig.trafficgen.traffic.durationHelp')}
              error={errors.duration?.message}
            >
              {(control) => (
                <Input
                  {...control}
                  type="number"
                  step={1}
                  {...register('duration', { valueAsNumber: true })}
                />
              )}
            </FormField>

            <FormField
              label={t('testConfig.trafficgen.traffic.warmup')}
              hint={t('help:tooltips.testConfig.trafficgen.traffic.warmupHelp')}
              error={errors.warmup?.message}
            >
              {(control) => (
                <Input
                  {...control}
                  type="number"
                  step={1}
                  {...register('warmup', { valueAsNumber: true })}
                />
              )}
            </FormField>
          </div>
        </FormSection>

        <FormSection title={t('testConfig.trafficgen.stream.title')}>
          <FormField
            label={t('testConfig.trafficgen.stream.id')}
            hint={t('help:tooltips.testConfig.trafficgen.stream.idHelp')}
            error={errors.streamId?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={1}
                {...register('streamId', { valueAsNumber: true })}
              />
            )}
          </FormField>
        </FormSection>

        <TestSummary>
          <div>
            {t('testConfig.common.selectedTests')}: {t('testConfig.trafficgen.tests.customStream')}
          </div>
          <div>
            {t('testConfig.trafficgen.summary.frame', {
              size: frameSize,
              rate: ratePct,
              throughput: calculateThroughput(),
            })}
          </div>
          <div>{t('testConfig.trafficgen.summary.duration', { duration, warmup })}</div>
        </TestSummary>
      </div>
    </CollapsibleSection>
  );
}

export default TrafficGenConfigForm;

/**
 * @fileoverview The Stem - Traffic Generator Configuration
 * @description Migrated to react-hook-form + valibot per #325. MAC fields
 *              now validate format (or accept empty string for "auto").
 */

import { Radio } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { FRAME_SIZE_OPTIONS } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { TrafficGenConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { FieldError } from './FieldError';
import { FormSection } from './FormSection';
import { HelpIcon } from './HelpIcon';
import { TestSummary } from './TestSummary';

/** Traffic generator configuration parameters */
export interface TrafficGenConfig {
  frameSize: number;
  ratePct: number;
  duration: number;
  warmup: number;
  streamId: number;
  burstMode: boolean;
  burstSize: number;
  interBurstGapUs: number;
  srcMac: string;
  dstMac: string;
  vlanId: number;
  vlanPriority: number;
}

/** Default traffic generator configuration */
export const defaultTrafficGenConfig: TrafficGenConfig = {
  frameSize: 64,
  ratePct: 100,
  duration: 60,
  warmup: 2,
  streamId: 1,
  burstMode: false,
  burstSize: 100,
  interBurstGapUs: 1000,
  srcMac: '',
  dstMac: '',
  vlanId: 0,
  vlanPriority: 0,
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
  const hasTrafficGenTests = selectedTests.some(
    (t) => t.startsWith('trafficgen_') || t === 'custom_stream',
  );

  const form = useConfigForm<TrafficGenConfig>({
    schema: TrafficGenConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    watch,
    setValue,
    formState: { errors },
  } = form;
  const { t } = useTranslation('settings');

  if (!hasTrafficGenTests) {
    return null;
  }

  const frameSize = watch('frameSize') ?? 0;
  const ratePct = watch('ratePct') ?? 0;
  const duration = watch('duration') ?? 0;
  const warmup = watch('warmup') ?? 0;
  const burstMode = watch('burstMode') ?? false;
  const burstSize = watch('burstSize') ?? 0;
  const interBurstGapUs = watch('interBurstGapUs') ?? 0;
  const vlanId = watch('vlanId') ?? 0;
  const vlanPriority = watch('vlanPriority') ?? 0;

  const hasCustomStream = selectedTests.includes('custom_stream');
  const hasBurst = selectedTests.includes('trafficgen_burst');
  const hasMultiStream = selectedTests.includes('trafficgen_multistream');

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
            <div>
              <label htmlFor="tgen-framesize" className="flex items-center gap-tight label">
                {t('testConfig.trafficgen.traffic.frameSize')}
                <HelpIcon tooltip={t('testConfig.trafficgen.traffic.frameSizeHelp')} />
              </label>
              <select
                id="tgen-framesize"
                {...register('frameSize', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {FRAME_SIZE_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {t(`testConfig.common.${option.qualifier}`, { size: option.value })}
                  </option>
                ))}
              </select>
              <FieldError message={errors.frameSize?.message} />
            </div>

            <div>
              <label htmlFor="tgen-rate" className="flex items-center gap-tight label">
                {t('testConfig.trafficgen.traffic.rate')}
                <HelpIcon tooltip={t('testConfig.trafficgen.traffic.rateHelp')} />
              </label>
              <div className="mt-tight flex gap-compact">
                <input
                  id="tgen-rate"
                  type="number"
                  step={0.01}
                  {...register('ratePct', { valueAsNumber: true })}
                  className="w-full"
                />
              </div>
              <FieldError message={errors.ratePct?.message} />
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
            <div>
              <label htmlFor="tgen-duration" className="flex items-center gap-tight label">
                {t('testConfig.trafficgen.traffic.duration')}
                <HelpIcon tooltip={t('testConfig.trafficgen.traffic.durationHelp')} />
              </label>
              <input
                id="tgen-duration"
                type="number"
                step={1}
                {...register('duration', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.duration?.message} />
            </div>

            <div>
              <label htmlFor="tgen-warmup" className="flex items-center gap-tight label">
                {t('testConfig.trafficgen.traffic.warmup')}
                <HelpIcon tooltip={t('testConfig.trafficgen.traffic.warmupHelp')} />
              </label>
              <input
                id="tgen-warmup"
                type="number"
                step={1}
                {...register('warmup', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.warmup?.message} />
            </div>
          </div>
        </FormSection>

        {hasMultiStream ? (
          <FormSection title={t('testConfig.trafficgen.stream.title')}>
            <div>
              <label htmlFor="tgen-streamid" className="flex items-center gap-tight label">
                {t('testConfig.trafficgen.stream.id')}
                <HelpIcon tooltip={t('testConfig.trafficgen.stream.idHelp')} />
              </label>
              <input
                id="tgen-streamid"
                type="number"
                step={1}
                {...register('streamId', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.streamId?.message} />
            </div>
          </FormSection>
        ) : null}

        {hasBurst ? (
          <FormSection title={t('testConfig.trafficgen.burst.title')}>
            <div className="flex items-center gap-compact">
              <input
                id="tgen-burstmode"
                type="checkbox"
                {...register('burstMode')}
                aria-label={t('testConfig.trafficgen.burst.enableAria')}
                className="rounded border-surface-border"
              />
              <label
                htmlFor="tgen-burstmode"
                title={t('testConfig.trafficgen.burst.enableTitle')}
                className="text-sm text-text-primary"
              >
                {t('testConfig.trafficgen.burst.enable')}
              </label>
            </div>

            {burstMode ? (
              <div className="grid grid-cols-2 gap-default">
                <div>
                  <label htmlFor="tgen-burstsize" className="flex items-center gap-tight label">
                    {t('testConfig.trafficgen.burst.size')}
                    <HelpIcon tooltip={t('testConfig.trafficgen.burst.sizeHelp')} />
                  </label>
                  <input
                    id="tgen-burstsize"
                    type="number"
                    step={1}
                    {...register('burstSize', { valueAsNumber: true })}
                    className="mt-tight w-full"
                  />
                  <FieldError message={errors.burstSize?.message} />
                </div>
                <div>
                  <label htmlFor="tgen-ibg" className="flex items-center gap-tight label">
                    {t('testConfig.trafficgen.burst.gap')}
                    <HelpIcon tooltip={t('testConfig.trafficgen.burst.gapHelp')} />
                  </label>
                  <input
                    id="tgen-ibg"
                    type="number"
                    step={1}
                    {...register('interBurstGapUs', { valueAsNumber: true })}
                    className="mt-tight w-full"
                  />
                  <FieldError message={errors.interBurstGapUs?.message} />
                </div>
              </div>
            ) : null}
          </FormSection>
        ) : null}

        <FormSection title={t('testConfig.trafficgen.vlan.title')}>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="tgen-vlanid" className="flex items-center gap-tight label">
                {t('testConfig.trafficgen.vlan.id')}
                <HelpIcon tooltip={t('testConfig.trafficgen.vlan.idHelp')} />
              </label>
              <input
                id="tgen-vlanid"
                type="number"
                step={1}
                {...register('vlanId', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.vlanId?.message} />
            </div>
            <div>
              <label htmlFor="tgen-vlanpri" className="flex items-center gap-tight label">
                {t('testConfig.trafficgen.vlan.priority')}
                <HelpIcon tooltip={t('testConfig.trafficgen.vlan.priorityHelp')} />
              </label>
              <input
                id="tgen-vlanpri"
                type="number"
                step={1}
                disabled={vlanId === 0}
                {...register('vlanPriority', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.vlanPriority?.message} />
            </div>
          </div>
        </FormSection>

        <FormSection title={t('testConfig.trafficgen.mac.title')}>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="tgen-srcmac" className="flex items-center gap-tight label">
                {t('testConfig.trafficgen.mac.src')}
                <HelpIcon tooltip={t('testConfig.trafficgen.mac.srcHelp')} />
              </label>
              <input
                id="tgen-srcmac"
                type="text"
                placeholder="aa:bb:cc:dd:ee:ff"
                {...register('srcMac')}
                className="mt-tight w-full"
              />
              <FieldError message={errors.srcMac?.message} />
            </div>
            <div>
              <label htmlFor="tgen-dstmac" className="flex items-center gap-tight label">
                {t('testConfig.trafficgen.mac.dst')}
                <HelpIcon tooltip={t('testConfig.trafficgen.mac.dstHelp')} />
              </label>
              <input
                id="tgen-dstmac"
                type="text"
                placeholder="aa:bb:cc:dd:ee:ff"
                {...register('dstMac')}
                className="mt-tight w-full"
              />
              <FieldError message={errors.dstMac?.message} />
            </div>
          </div>
        </FormSection>

        <TestSummary>
          <div>
            {t('testConfig.common.selectedTests')}:{' '}
            {[
              hasCustomStream && t('testConfig.trafficgen.tests.customStream'),
              hasBurst && t('testConfig.trafficgen.tests.burstMode'),
              hasMultiStream && t('testConfig.trafficgen.tests.multiStream'),
            ]
              .filter(Boolean)
              .join(', ')}
          </div>
          <div>
            {t('testConfig.trafficgen.summary.frame', {
              size: frameSize,
              rate: ratePct,
              throughput: calculateThroughput(),
            })}
          </div>
          {vlanId > 0 ? (
            <div>
              {t('testConfig.trafficgen.summary.vlan', { vlan: vlanId, priority: vlanPriority })}
            </div>
          ) : null}
          {burstMode ? (
            <div>
              {t('testConfig.trafficgen.summary.burst', {
                frames: burstSize,
                gap: interBurstGapUs,
              })}
            </div>
          ) : null}
          <div>{t('testConfig.trafficgen.summary.duration', { duration, warmup })}</div>
        </TestSummary>
      </div>
    </CollapsibleSection>
  );
}

export default TrafficGenConfigForm;

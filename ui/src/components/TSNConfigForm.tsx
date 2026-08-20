/**
 * @fileoverview The Stem - TSN Test Configuration
 * @description Migrated to react-hook-form + valibot per #325. Uses
 *              FormProvider/useFormContext so the sub-component
 *              decomposition (test params / timing / PTP / scheduling /
 *              summary) doesn't have to thread the form instance
 *              through props.
 */

import { AlertTriangle, Clock } from 'lucide-react';
import type { ReactElement } from 'react';
import { FormProvider, useFormContext } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import type { FrameSizeOption } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { TSNConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { FieldError } from './FieldError';
import { FormSection } from './FormSection';
import { HelpIcon } from './HelpIcon';
import { TestSummary } from './TestSummary';

/** TSN test configuration parameters */
export interface TSNConfig {
  duration: number;
  warmup: number;
  frameSize: number;
  maxLatencyNs: number;
  maxJitterNs: number;
  requirePTPSync: boolean;
  maxSyncOffsetNs: number;
  ptpEnabled: boolean;
  preemptionEnabled: boolean;
  numTrafficClasses: number;
  baseTimeNs: number;
  cycleTimeNs: number;
  trafficClass: number;
}

/** Default TSN configuration */
export const defaultTSNConfig: TSNConfig = {
  duration: 60,
  warmup: 5,
  frameSize: 64,
  maxLatencyNs: 1000000,
  maxJitterNs: 100000,
  requirePTPSync: true,
  maxSyncOffsetNs: 1000,
  ptpEnabled: true,
  preemptionEnabled: false,
  numTrafficClasses: 8,
  baseTimeNs: 0,
  cycleTimeNs: 1000000,
  trafficClass: 7,
};

/** TSN measures bounded latency at standard sizes; no jumbo. */
const FRAME_SIZE_OPTIONS: FrameSizeOption[] = [
  { value: 64, qualifier: 'frameSizeMin' },
  { value: 128, qualifier: 'frameSize' },
  { value: 256, qualifier: 'frameSize' },
  { value: 512, qualifier: 'frameSize' },
  { value: 1024, qualifier: 'frameSize' },
  { value: 1518, qualifier: 'frameSizeMax' },
];

const CYCLE_TIME_OPTIONS: Array<{ value: number; label: string }> = [
  { value: 125000, label: '125 us' },
  { value: 250000, label: '250 us' },
  { value: 500000, label: '500 us' },
  { value: 1000000, label: '1 ms' },
  { value: 2000000, label: '2 ms' },
  { value: 4000000, label: '4 ms' },
];

function formatNs(ns: number): string {
  if (ns >= 1000000000) return `${(ns / 1000000000).toFixed(1)} s`;
  if (ns >= 1000000) return `${(ns / 1000000).toFixed(1)} ms`;
  if (ns >= 1000) return `${(ns / 1000).toFixed(1)} us`;
  return `${ns} ns`;
}

function TestParametersSection(): ReactElement {
  const {
    register,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  const { t } = useTranslation('settings');
  return (
    <FormSection title={t('testConfig.tsn.params.title')}>
      <div className="grid grid-cols-3 gap-default">
        <div>
          <label htmlFor="tsn-duration" className="flex items-center gap-tight label">
            {t('testConfig.tsn.params.duration')}
            <HelpIcon tooltip={t('testConfig.tsn.params.durationHelp')} />
          </label>
          <input
            id="tsn-duration"
            type="number"
            step={1}
            {...register('duration', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.duration?.message} />
        </div>
        <div>
          <label htmlFor="tsn-warmup" className="flex items-center gap-tight label">
            {t('testConfig.tsn.params.warmup')}
            <HelpIcon tooltip={t('testConfig.tsn.params.warmupHelp')} />
          </label>
          <input
            id="tsn-warmup"
            type="number"
            step={1}
            {...register('warmup', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.warmup?.message} />
        </div>
        <div>
          <label htmlFor="tsn-framesize" className="flex items-center gap-tight label">
            {t('testConfig.tsn.params.frameSize')}
            <HelpIcon tooltip={t('testConfig.tsn.params.frameSizeHelp')} />
          </label>
          <select
            id="tsn-framesize"
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
      </div>
    </FormSection>
  );
}

function TimingRequirementsSection(): ReactElement {
  const {
    register,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  const { t } = useTranslation('settings');
  return (
    <FormSection title={t('testConfig.tsn.timing.title')}>
      <div className="grid grid-cols-2 gap-default">
        <div>
          <label htmlFor="tsn-maxlatency" className="flex items-center gap-tight label">
            {t('testConfig.tsn.timing.maxLatency')}
            <HelpIcon tooltip={t('testConfig.tsn.timing.maxLatencyHelp')} />
          </label>
          <input
            id="tsn-maxlatency"
            type="number"
            step={1000}
            {...register('maxLatencyNs', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.maxLatencyNs?.message} />
        </div>
        <div>
          <label htmlFor="tsn-maxjitter" className="flex items-center gap-tight label">
            {t('testConfig.tsn.timing.maxJitter')}
            <HelpIcon tooltip={t('testConfig.tsn.timing.maxJitterHelp')} />
          </label>
          <input
            id="tsn-maxjitter"
            type="number"
            step={1000}
            {...register('maxJitterNs', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.maxJitterNs?.message} />
        </div>
      </div>
    </FormSection>
  );
}

function PTPConfigSection(): ReactElement {
  const {
    register,
    watch,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  const { t } = useTranslation('settings');
  const ptpEnabled = watch('ptpEnabled');
  return (
    <FormSection title={t('testConfig.tsn.ptp.title')}>
      <div className="stack-sm">
        <div className="flex items-center gap-compact">
          <input
            id="tsn-ptpenabled"
            type="checkbox"
            {...register('ptpEnabled')}
            aria-label={t('testConfig.tsn.ptp.enableAria')}
            className="rounded border-surface-border"
          />
          <label
            htmlFor="tsn-ptpenabled"
            title={t('testConfig.tsn.ptp.enableTitle')}
            className="text-sm text-text-primary"
          >
            {t('testConfig.tsn.ptp.enable')}
          </label>
        </div>
        <div className="flex items-center gap-compact">
          <input
            id="tsn-requiresync"
            type="checkbox"
            {...register('requirePTPSync')}
            aria-label={t('testConfig.tsn.ptp.requireAria')}
            className="rounded border-surface-border"
          />
          <label
            htmlFor="tsn-requiresync"
            title={t('testConfig.tsn.ptp.requireTitle')}
            className="text-sm text-text-primary"
          >
            {t('testConfig.tsn.ptp.require')}
          </label>
        </div>
      </div>
      {ptpEnabled ? (
        <div>
          <label htmlFor="tsn-syncoffset" className="flex items-center gap-tight label">
            {t('testConfig.tsn.ptp.syncOffset')}
            <HelpIcon tooltip={t('testConfig.tsn.ptp.syncOffsetHelp')} />
          </label>
          <input
            id="tsn-syncoffset"
            type="number"
            step={1}
            {...register('maxSyncOffsetNs', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.maxSyncOffsetNs?.message} />
        </div>
      ) : null}
    </FormSection>
  );
}

function SchedulingConfigSection(): ReactElement {
  const {
    register,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  const { t } = useTranslation('settings');
  return (
    <FormSection title={t('testConfig.tsn.scheduling.title')}>
      <div className="flex items-center gap-compact">
        <input
          id="tsn-preemption"
          type="checkbox"
          {...register('preemptionEnabled')}
          aria-label={t('testConfig.tsn.scheduling.preemptionAria')}
          className="rounded border-surface-border"
        />
        <label
          htmlFor="tsn-preemption"
          title={t('testConfig.tsn.scheduling.preemptionTitle')}
          className="text-sm text-text-primary"
        >
          {t('testConfig.tsn.scheduling.preemption')}
        </label>
      </div>
      <div className="grid grid-cols-2 gap-default">
        <div>
          <label htmlFor="tsn-cycletime" className="flex items-center gap-tight label">
            {t('testConfig.tsn.scheduling.cycleTime')}
            <HelpIcon tooltip={t('testConfig.tsn.scheduling.cycleTimeHelp')} />
          </label>
          <select
            id="tsn-cycletime"
            {...register('cycleTimeNs', { valueAsNumber: true })}
            className="mt-tight w-full"
          >
            {CYCLE_TIME_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          <FieldError message={errors.cycleTimeNs?.message} />
        </div>
        <div>
          <label htmlFor="tsn-trafficclass" className="flex items-center gap-tight label">
            {t('testConfig.tsn.scheduling.trafficClass')}
            <HelpIcon tooltip={t('testConfig.tsn.scheduling.trafficClassHelp')} />
          </label>
          <input
            id="tsn-trafficclass"
            type="number"
            step={1}
            {...register('trafficClass', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.trafficClass?.message} />
        </div>
      </div>
      <div>
        <label htmlFor="tsn-numclasses" className="flex items-center gap-tight label">
          {t('testConfig.tsn.scheduling.numClasses')}
          <HelpIcon tooltip={t('testConfig.tsn.scheduling.numClassesHelp')} />
        </label>
        <input
          id="tsn-numclasses"
          type="number"
          step={1}
          {...register('numTrafficClasses', { valueAsNumber: true })}
          className="mt-tight w-full"
        />
        <FieldError message={errors.numTrafficClasses?.message} />
      </div>
      <div>
        <label htmlFor="tsn-basetime" className="flex items-center gap-tight label">
          {t('testConfig.tsn.scheduling.baseTime')}
          <HelpIcon tooltip={t('testConfig.tsn.scheduling.baseTimeHelp')} />
        </label>
        <input
          id="tsn-basetime"
          type="number"
          step={1}
          {...register('baseTimeNs', { valueAsNumber: true })}
          className="mt-tight w-full"
        />
        <FieldError message={errors.baseTimeNs?.message} />
      </div>
    </FormSection>
  );
}

interface TestSummarySectionProps {
  hasLatency: boolean;
  hasJitter: boolean;
  hasSync: boolean;
  hasPreemption: boolean;
  hasScheduling: boolean;
}

function TestSummarySection({
  hasLatency,
  hasJitter,
  hasSync,
  hasPreemption,
  hasScheduling,
}: TestSummarySectionProps): ReactElement {
  const { watch } = useFormContext<TSNConfig>();
  const { t } = useTranslation('settings');
  const v = watch();
  const selectedTestNames = [
    hasLatency && t('testConfig.tsn.tests.latency'),
    hasJitter && t('testConfig.tsn.tests.jitter'),
    hasSync && t('testConfig.tsn.tests.sync'),
    hasPreemption && t('testConfig.tsn.tests.preemption'),
    hasScheduling && t('testConfig.tsn.tests.scheduling'),
  ].filter(Boolean);

  const ptpKey = !v.ptpEnabled
    ? 'testConfig.tsn.summary.ptpDisabled'
    : v.requirePTPSync
      ? 'testConfig.tsn.summary.ptpRequired'
      : 'testConfig.tsn.summary.ptpEnabled';

  return (
    <TestSummary>
      <div>
        {t('testConfig.common.selectedTests')}: {selectedTestNames.join(', ')}
      </div>
      <div>{t('testConfig.tsn.summary.frameSize', { size: v.frameSize })}</div>
      <div>
        {t('testConfig.tsn.summary.timing', {
          latency: formatNs(v.maxLatencyNs),
          jitter: formatNs(v.maxJitterNs),
        })}
      </div>
      <div>{t(ptpKey as never, { offset: formatNs(v.maxSyncOffsetNs) })}</div>
      {hasScheduling || hasPreemption ? (
        <div>
          {t(
            v.preemptionEnabled
              ? 'testConfig.tsn.summary.schedulingPreemption'
              : 'testConfig.tsn.summary.scheduling',
            { cycle: formatNs(v.cycleTimeNs), trafficClass: v.trafficClass },
          )}
        </div>
      ) : null}
      <div>{t('testConfig.tsn.summary.duration', { duration: v.duration, warmup: v.warmup })}</div>
    </TestSummary>
  );
}

interface TSNConfigFormProps {
  config: TSNConfig;
  setConfig: (config: TSNConfig) => void;
  selectedTests: string[];
}

export function TSNConfigForm({
  config,
  setConfig,
  selectedTests,
}: TSNConfigFormProps): ReactElement | null {
  const hasTSNTests = selectedTests.some((t) => t.startsWith('tsn_'));

  const form = useConfigForm<TSNConfig>({
    schema: TSNConfigSchema,
    config,
    setConfig,
  });

  if (!hasTSNTests) {
    return null;
  }

  const hasLatency = selectedTests.includes('tsn_latency');
  const hasJitter = selectedTests.includes('tsn_jitter');
  const hasSync = selectedTests.includes('tsn_sync');
  const hasPreemption = selectedTests.includes('tsn_preemption');
  const hasScheduling = selectedTests.includes('tsn_scheduling');

  const rootErrors = form.formState.errors.root;
  const crossFieldError = rootErrors
    ? Object.values(rootErrors).find(
        (e): e is { message: string } =>
          typeof e === 'object' && e !== null && 'message' in e && typeof e.message === 'string',
      )
    : undefined;

  return (
    <CollapsibleSection
      testId="tsn-config-form"
      title={
        <div className="flex items-center gap-compact">
          <Clock className="w-4 h-4" />
          <span>TSN Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <FormProvider {...form}>
        <div className="stack-lg">
          <TestParametersSection />
          <TimingRequirementsSection />
          <PTPConfigSection />

          {hasScheduling || hasPreemption ? <SchedulingConfigSection /> : null}

          {crossFieldError && (
            <div className="pad-xs rounded-lg bg-status-error/10 text-status-error text-sm flex items-center gap-compact">
              <AlertTriangle className="w-4 h-4" />
              {crossFieldError.message}
            </div>
          )}

          <TestSummarySection
            hasLatency={hasLatency}
            hasJitter={hasJitter}
            hasSync={hasSync}
            hasPreemption={hasPreemption}
            hasScheduling={hasScheduling}
          />
        </div>
      </FormProvider>
    </CollapsibleSection>
  );
}

export default TSNConfigForm;

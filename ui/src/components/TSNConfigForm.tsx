/**
 * @fileoverview Stem - TSN Test Configuration
 * @description Migrated to react-hook-form + valibot per #325. Uses
 *              FormProvider/useFormContext so the sub-component
 *              decomposition (test params / timing / PTP / scheduling /
 *              summary) doesn't have to thread the form instance
 *              through props.
 */

import { AlertTriangle } from 'lucide-react';
import type { ReactElement } from 'react';
import { FormProvider, useFormContext } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
import type { FrameSizeOption } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { hasGroupTests } from '../lib/testGroups';
import { TSNConfigSchema } from '../schemas/configs';
import { FormField } from '../shared/FormField';
import { Input } from '../shared/Input';
import { CheckboxField, NumberSelectField } from './ConfigFormFields';
import { FormSection } from './FormSection';
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
  frameSize: 128,
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

/**
 * TSN measures bounded latency at standard sizes; no jumbo, and no 64 — the
 * extended payload does not fit one (MIN_EXTENDED_FRAME_SIZE).
 */
const FRAME_SIZE_OPTIONS: FrameSizeOption[] = [
  { value: 128, qualifier: 'frameSizeMin' },
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
    control,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  const { t } = useTranslation(['settings', 'help']);
  return (
    <FormSection title={t('testConfig.tsn.params.title')}>
      <div className="grid grid-cols-3 gap-default">
        <FormField
          label={t('testConfig.tsn.params.duration')}
          hint={t('help:tooltips.testConfig.tsn.params.durationHelp')}
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
          label={t('testConfig.tsn.params.warmup')}
          hint={t('help:tooltips.testConfig.tsn.params.warmupHelp')}
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
        <NumberSelectField
          control={control}
          name="frameSize"
          label={t('testConfig.tsn.params.frameSize')}
          hint={t('help:tooltips.testConfig.tsn.params.frameSizeHelp')}
          error={errors.frameSize?.message}
          options={FRAME_SIZE_OPTIONS.map((option) => ({
            value: option.value,
            label: t(`testConfig.common.${option.qualifier}`, { size: option.value }),
          }))}
        />
      </div>
    </FormSection>
  );
}

function TimingRequirementsSection(): ReactElement {
  const {
    register,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  const { t } = useTranslation(['settings', 'help']);
  return (
    <FormSection title={t('testConfig.tsn.timing.title')}>
      <div className="grid grid-cols-2 gap-default">
        <FormField
          label={t('testConfig.tsn.timing.maxLatency')}
          hint={t('help:tooltips.testConfig.tsn.timing.maxLatencyHelp')}
          error={errors.maxLatencyNs?.message}
        >
          {(control) => (
            <Input
              {...control}
              type="number"
              step={1000}
              {...register('maxLatencyNs', { valueAsNumber: true })}
            />
          )}
        </FormField>
        <FormField
          label={t('testConfig.tsn.timing.maxJitter')}
          hint={t('help:tooltips.testConfig.tsn.timing.maxJitterHelp')}
          error={errors.maxJitterNs?.message}
        >
          {(control) => (
            <Input
              {...control}
              type="number"
              step={1000}
              {...register('maxJitterNs', { valueAsNumber: true })}
            />
          )}
        </FormField>
      </div>
    </FormSection>
  );
}

function PTPConfigSection(): ReactElement {
  const {
    register,
    control,
    watch,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  const { t } = useTranslation(['settings', 'help']);
  const ptpEnabled = watch('ptpEnabled');
  return (
    <FormSection title={t('testConfig.tsn.ptp.title')}>
      <div className="stack-sm">
        <CheckboxField
          control={control}
          name="ptpEnabled"
          label={t('testConfig.tsn.ptp.enable')}
          help={t('help:tooltips.testConfig.tsn.ptp.enableTitle')}
        />
        <CheckboxField
          control={control}
          name="requirePTPSync"
          label={t('testConfig.tsn.ptp.require')}
          help={t('help:tooltips.testConfig.tsn.ptp.requireTitle')}
        />
      </div>
      {ptpEnabled ? (
        <FormField
          label={t('testConfig.tsn.ptp.syncOffset')}
          hint={t('help:tooltips.testConfig.tsn.ptp.syncOffsetHelp')}
          error={errors.maxSyncOffsetNs?.message}
        >
          {(control) => (
            <Input
              {...control}
              type="number"
              step={1}
              {...register('maxSyncOffsetNs', { valueAsNumber: true })}
            />
          )}
        </FormField>
      ) : null}
    </FormSection>
  );
}

function SchedulingConfigSection(): ReactElement {
  const {
    register,
    control,
    formState: { errors },
  } = useFormContext<TSNConfig>();
  const { t } = useTranslation(['settings', 'help']);
  return (
    <FormSection title={t('testConfig.tsn.scheduling.title')}>
      <CheckboxField
        control={control}
        name="preemptionEnabled"
        label={t('testConfig.tsn.scheduling.preemption')}
        help={t('help:tooltips.testConfig.tsn.scheduling.preemptionTitle')}
      />
      <div className="grid grid-cols-2 gap-default">
        <NumberSelectField
          control={control}
          name="cycleTimeNs"
          label={t('testConfig.tsn.scheduling.cycleTime')}
          hint={t('help:tooltips.testConfig.tsn.scheduling.cycleTimeHelp')}
          error={errors.cycleTimeNs?.message}
          options={CYCLE_TIME_OPTIONS}
        />
        <FormField
          label={t('testConfig.tsn.scheduling.trafficClass')}
          hint={t('help:tooltips.testConfig.tsn.scheduling.trafficClassHelp')}
          error={errors.trafficClass?.message}
        >
          {(control) => (
            <Input
              {...control}
              type="number"
              step={1}
              {...register('trafficClass', { valueAsNumber: true })}
            />
          )}
        </FormField>
      </div>
      <FormField
        label={t('testConfig.tsn.scheduling.numClasses')}
        hint={t('help:tooltips.testConfig.tsn.scheduling.numClassesHelp')}
        error={errors.numTrafficClasses?.message}
      >
        {(control) => (
          <Input
            {...control}
            type="number"
            step={1}
            {...register('numTrafficClasses', { valueAsNumber: true })}
          />
        )}
      </FormField>
      <FormField
        label={t('testConfig.tsn.scheduling.baseTime')}
        hint={t('help:tooltips.testConfig.tsn.scheduling.baseTimeHelp')}
        error={errors.baseTimeNs?.message}
      >
        {(control) => (
          <Input
            {...control}
            type="number"
            step={1}
            {...register('baseTimeNs', { valueAsNumber: true })}
          />
        )}
      </FormField>
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
  const { t } = useTranslation(['settings', 'help']);
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
  /** Overrides the root test id: the drawer and the page can both render this form at once. */
  testId?: string;
}

export function TSNConfigForm({
  config,
  setConfig,
  selectedTests,
  testId = 'tsn-config-form',
}: TSNConfigFormProps): ReactElement | null {
  const hasTSNTests = hasGroupTests('tsn', selectedTests);

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
    <div data-testid={testId}>
      <FormProvider {...form}>
        <div className="stack-lg">
          <TestParametersSection />
          <TimingRequirementsSection />
          <PTPConfigSection />

          {hasScheduling || hasPreemption ? <SchedulingConfigSection /> : null}

          {crossFieldError && (
            <div className="pad-xs rounded-lg bg-status-error/10 text-status-error-strong text-sm flex items-center gap-compact">
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
    </div>
  );
}

export default TSNConfigForm;

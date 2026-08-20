/**
 * @fileoverview The Stem - RFC 2544 Benchmark Test Configuration
 * @description Advanced configuration form for RFC 2544 Benchmarking Tests.
 *              Migrated to react-hook-form + valibot per #325. The schema
 *              lives at src/schemas/configs.ts (RFC2544ConfigSchema).
 */

import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { FRAME_SIZE_OPTIONS } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { RFC2544ConfigSchema } from '../schemas/configs';
import { FieldError } from './FieldError';
import { FormSection } from './FormSection';
import { HelpIcon } from './HelpIcon';
import { TestSummary } from './TestSummary';

/** RFC 2544 test configuration parameters */
export interface RFC2544Config {
  /** Test duration in seconds */
  duration: number;
  /** Frame sizes to test (bytes) */
  frameSizes: number[];
  /** Resolution for binary search (percentage) */
  resolution: number;
  /** Maximum acceptable frame loss (percentage) */
  maxLoss: number;
  /** Warmup duration before measurement (seconds) */
  warmup: number;
  /** Number of trials per test point */
  trials: number;
  /** Step size for frame loss rate test (percentage) */
  stepSize: number;
  /** Enable bidirectional testing */
  bidirectional: boolean;
}

/** Default RFC 2544 configuration */
export const defaultRFC2544Config: RFC2544Config = {
  duration: 60,
  frameSizes: [64, 128, 256, 512, 1024, 1280, 1518],
  resolution: 0.1,
  maxLoss: 0.0,
  warmup: 2,
  trials: 3,
  stepSize: 10,
  bidirectional: false,
};

interface RFC2544ConfigFormProps {
  config: RFC2544Config;
  setConfig: (config: RFC2544Config) => void;
  selectedTests: string[];
}

export function RFC2544ConfigForm({
  config,
  setConfig,
  selectedTests,
}: RFC2544ConfigFormProps): ReactElement | null {
  const { t } = useTranslation('settings');
  const hasRFC2544Tests = selectedTests.some((id) => id.startsWith('rfc2544'));

  const form = useConfigForm<RFC2544Config>({
    schema: RFC2544ConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    watch,
    setValue,
    formState: { errors },
  } = form;

  if (!hasRFC2544Tests) {
    return null;
  }

  const frameSizes = watch('frameSizes') ?? [];
  const duration = watch('duration') ?? 0;
  const warmup = watch('warmup') ?? 0;
  const trials = watch('trials') ?? 0;
  const resolution = watch('resolution') ?? 0;
  const maxLoss = watch('maxLoss') ?? 0;
  const stepSize = watch('stepSize') ?? 0;
  const bidirectional = watch('bidirectional') ?? false;

  const toggleFrameSize = (size: number): void => {
    if (frameSizes.includes(size)) {
      setValue(
        'frameSizes',
        frameSizes.filter((s) => s !== size),
        { shouldValidate: true, shouldDirty: true },
      );
    } else {
      setValue(
        'frameSizes',
        [...frameSizes, size].sort((a, b) => a - b),
        {
          shouldValidate: true,
          shouldDirty: true,
        },
      );
    }
  };

  const hasThroughput = selectedTests.includes('rfc2544_throughput');
  const hasLatency = selectedTests.includes('rfc2544_latency');
  const hasFrameLoss = selectedTests.includes('rfc2544_frame_loss');
  const hasBackToBack = selectedTests.includes('rfc2544_back_to_back');

  return (
    <div data-testid="rfc2544-config-form" className="stack-lg">
      <FormSection title={t('testConfig.rfc2544.duration.title')}>
        <div>
          <label htmlFor="rfc2544-duration" className="flex items-center gap-tight label">
            {t('testConfig.rfc2544.duration.perTest')}
            <HelpIcon tooltip={t('testConfig.rfc2544.duration.perTestHelp')} />
          </label>
          <input
            id="rfc2544-duration"
            type="number"
            step={1}
            {...register('duration', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.duration?.message} />
        </div>

        <div>
          <label htmlFor="rfc2544-warmup" className="flex items-center gap-tight label">
            {t('testConfig.rfc2544.duration.warmup')}
            <HelpIcon tooltip={t('testConfig.rfc2544.duration.warmupHelp')} />
          </label>
          <input
            id="rfc2544-warmup"
            type="number"
            step={1}
            {...register('warmup', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.warmup?.message} />
        </div>

        <div>
          <label htmlFor="rfc2544-trials" className="flex items-center gap-tight label">
            {t('testConfig.rfc2544.duration.trials')}
            <HelpIcon tooltip={t('testConfig.rfc2544.duration.trialsHelp')} />
          </label>
          <input
            id="rfc2544-trials"
            type="number"
            step={1}
            {...register('trials', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.trials?.message} />
        </div>
      </FormSection>

      {hasThroughput ? (
        <FormSection title={t('testConfig.rfc2544.throughput.title')}>
          <div>
            <label htmlFor="rfc2544-resolution" className="flex items-center gap-tight label">
              {t('testConfig.rfc2544.throughput.resolution')}
              <HelpIcon tooltip={t('testConfig.rfc2544.throughput.resolutionHelp')} />
            </label>
            <input
              id="rfc2544-resolution"
              type="number"
              step={0.01}
              {...register('resolution', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.resolution?.message} />
          </div>

          <div>
            <label htmlFor="rfc2544-maxloss" className="flex items-center gap-tight label">
              {t('testConfig.rfc2544.throughput.maxLoss')}
              <HelpIcon tooltip={t('testConfig.rfc2544.throughput.maxLossHelp')} />
            </label>
            <input
              id="rfc2544-maxloss"
              type="number"
              step={0.001}
              {...register('maxLoss', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.maxLoss?.message} />
          </div>
        </FormSection>
      ) : null}

      {hasFrameLoss ? (
        <FormSection title={t('testConfig.rfc2544.frameLoss.title')}>
          <div>
            <label htmlFor="rfc2544-stepsize" className="flex items-center gap-tight label">
              {t('testConfig.rfc2544.frameLoss.stepSize')}
              <HelpIcon tooltip={t('testConfig.rfc2544.frameLoss.stepSizeHelp')} />
            </label>
            <input
              id="rfc2544-stepsize"
              type="number"
              step={1}
              {...register('stepSize', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.stepSize?.message} />
            <div className="text-xs text-text-muted mt-tight">
              {t('testConfig.rfc2544.frameLoss.testsAt', {
                points: Array.from(
                  { length: Math.floor(100 / Math.max(1, stepSize)) + 1 },
                  (_, i) => `${i * stepSize}%`,
                ).join(', '),
              })}
            </div>
          </div>
        </FormSection>
      ) : null}

      <FormSection
        title={t('testConfig.common.frameSizesTitle')}
        help={<HelpIcon tooltip={t('testConfig.rfc2544.frameSizes.help')} />}
      >
        <div className="grid grid-cols-2 gap-compact">
          {FRAME_SIZE_OPTIONS.map((option) => (
            <label
              key={option.value}
              title={t('testConfig.rfc2544.frameSizes.includeTitle', { size: option.value })}
              className="flex items-center gap-compact pad-xs rounded-lg cursor-pointer hover:bg-surface-hover text-sm"
            >
              <input
                type="checkbox"
                checked={frameSizes.includes(option.value)}
                onChange={() => toggleFrameSize(option.value)}
                aria-label={t('testConfig.common.frameSizeAria', { size: option.value })}
                className="w-4 h-4 accent-brand-primary"
              />
              <span className="text-text-primary">
                {t(`testConfig.common.${option.qualifier}`, { size: option.value })}
              </span>
            </label>
          ))}
        </div>
        <FieldError message={errors.frameSizes?.message} />
      </FormSection>

      <FormSection title={t('testConfig.rfc2544.advanced.title')}>
        <label
          title={t('testConfig.rfc2544.advanced.bidirectionalTitle')}
          className="flex items-center gap-default pad-xs rounded-lg cursor-pointer hover:bg-surface-hover"
        >
          <input
            type="checkbox"
            {...register('bidirectional')}
            aria-label={t('testConfig.rfc2544.advanced.bidirectionalAria')}
            className="w-4 h-4 accent-brand-primary"
          />
          <div>
            <div className="font-medium text-sm flex items-center gap-tight">
              {t('testConfig.rfc2544.advanced.bidirectional')}
              <HelpIcon tooltip={t('testConfig.rfc2544.advanced.bidirectionalHelp')} />
            </div>
            <div className="text-xs text-text-muted">
              {t('testConfig.rfc2544.advanced.bidirectionalHint')}
            </div>
          </div>
        </label>
      </FormSection>

      <TestSummary>
        <div>
          {t('testConfig.common.selectedTests')}:{' '}
          {[
            hasThroughput && t('testConfig.rfc2544.tests.throughput'),
            hasLatency && t('testConfig.rfc2544.tests.latency'),
            hasFrameLoss && t('testConfig.rfc2544.tests.frameLoss'),
            hasBackToBack && t('testConfig.rfc2544.tests.backToBack'),
            selectedTests.includes('rfc2544_system_recovery') &&
              t('testConfig.rfc2544.tests.systemRecovery'),
            selectedTests.includes('rfc2544_reset') && t('testConfig.rfc2544.tests.reset'),
          ]
            .filter(Boolean)
            .join(', ')}
        </div>
        <div>{t('testConfig.common.frameSizesSummary', { sizes: frameSizes.join(', ') })}</div>
        <div>
          {warmup > 0
            ? t('testConfig.rfc2544.summary.durationWithWarmup', { duration, trials, warmup })
            : t('testConfig.rfc2544.summary.duration', { duration, trials })}
        </div>
        {hasThroughput ? (
          <div>{t('testConfig.rfc2544.summary.throughput', { resolution, maxLoss })}</div>
        ) : null}
        {bidirectional ? <div>{t('testConfig.rfc2544.summary.bidirectional')}</div> : null}
        <div className="pt-tight border-t border-surface-border mt-tight">
          {t('testConfig.common.estimatedTime', {
            minutes: Math.ceil(
              ((duration + warmup) *
                trials *
                frameSizes.length *
                selectedTests.filter((id) => id.startsWith('rfc2544')).length) /
                60,
            ),
          })}
        </div>
      </TestSummary>
    </div>
  );
}

export default RFC2544ConfigForm;

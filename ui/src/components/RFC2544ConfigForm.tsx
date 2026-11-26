/**
 * @fileoverview Stem - RFC 2544 Benchmark Test Configuration
 * @description Advanced configuration form for RFC 2544 Benchmarking Tests.
 *              Migrated to react-hook-form + valibot per #325. The schema
 *              lives at src/schemas/configs.ts (RFC2544ConfigSchema).
 */

import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { FRAME_SIZE_OPTIONS } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { hasGroupTests } from '../lib/testGroups';
import { RFC2544ConfigSchema } from '../schemas/configs';
import { Checkbox } from '../shared/Checkbox';
import { FormField } from '../shared/FormField';
import { FormSection } from '../shared/FormSection';
import { Input } from '../shared/Input';
import { CheckboxField } from './ConfigFormFields';
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
  /**
   * Overrides the root test id. This form is rendered in two places at once —
   * the benchmark page and the settings drawer's RFC 2544 section — and it
   * used to hardcode `rfc2544-config-form` in both, so with the drawer open
   * the id matched two live elements and any locator using it was a
   * strict-mode violation waiting for someone to open the drawer (#642).
   */
  testId?: string;
}

export function RFC2544ConfigForm({
  config,
  setConfig,
  selectedTests,
  testId = 'rfc2544-config-form',
}: RFC2544ConfigFormProps): ReactElement | null {
  const { t } = useTranslation(['settings', 'help', 'common']);
  const hasRFC2544Tests = hasGroupTests('rfc2544', selectedTests);

  const form = useConfigForm<RFC2544Config>({
    schema: RFC2544ConfigSchema,
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
    <div data-testid={testId} className="stack-lg">
      <FormSection title={t('testConfig.rfc2544.duration.title')}>
        <FormField
          label={t('testConfig.rfc2544.duration.perTest')}
          hint={t('help:tooltips.testConfig.rfc2544.duration.perTestHelp')}
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
          label={t('testConfig.rfc2544.duration.warmup')}
          hint={t('help:tooltips.testConfig.rfc2544.duration.warmupHelp')}
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

        <FormField
          label={t('testConfig.rfc2544.duration.trials')}
          hint={t('help:tooltips.testConfig.rfc2544.duration.trialsHelp')}
          error={errors.trials?.message}
        >
          {(control) => (
            <Input
              {...control}
              type="number"
              step={1}
              {...register('trials', { valueAsNumber: true })}
            />
          )}
        </FormField>
      </FormSection>

      {hasThroughput ? (
        <FormSection title={t('testConfig.rfc2544.throughput.title')}>
          <FormField
            label={t('testConfig.rfc2544.throughput.resolution')}
            hint={t('help:tooltips.testConfig.rfc2544.throughput.resolutionHelp')}
            error={errors.resolution?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={0.01}
                {...register('resolution', { valueAsNumber: true })}
              />
            )}
          </FormField>

          <FormField
            label={t('testConfig.rfc2544.throughput.maxLoss')}
            hint={t('help:tooltips.testConfig.rfc2544.throughput.maxLossHelp')}
            error={errors.maxLoss?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={0.001}
                {...register('maxLoss', { valueAsNumber: true })}
              />
            )}
          </FormField>
        </FormSection>
      ) : null}

      {hasFrameLoss ? (
        <FormSection title={t('testConfig.rfc2544.frameLoss.title')}>
          <FormField
            label={t('testConfig.rfc2544.frameLoss.stepSize')}
            hint={t('help:tooltips.testConfig.rfc2544.frameLoss.stepSizeHelp')}
            error={errors.stepSize?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={1}
                {...register('stepSize', { valueAsNumber: true })}
              />
            )}
          </FormField>
          <div className="text-xs text-text-muted mt-tight">
            {t('testConfig.rfc2544.frameLoss.testsAt', {
              points: Array.from(
                { length: Math.floor(100 / Math.max(1, stepSize)) + 1 },
                (_, i) => `${i * stepSize}%`,
              ).join(', '),
            })}
          </div>
        </FormSection>
      ) : null}

      <FormField
        group
        size="lg"
        label={t('testConfig.common.frameSizesTitle')}
        hint={t('help:tooltips.testConfig.rfc2544.frameSizes.help')}
        error={errors.frameSizes?.message}
      >
        <div className="grid grid-cols-2 gap-compact">
          {FRAME_SIZE_OPTIONS.map((option) => (
            <Checkbox
              key={option.value}
              checked={frameSizes.includes(option.value)}
              onChange={() => toggleFrameSize(option.value)}
              label={t(`testConfig.common.${option.qualifier}`, { size: option.value })}
            />
          ))}
        </div>
      </FormField>

      <FormSection title={t('testConfig.rfc2544.advanced.title')}>
        <CheckboxField
          control={control}
          name="bidirectional"
          label={t('testConfig.rfc2544.advanced.bidirectional')}
          description={t('testConfig.rfc2544.advanced.bidirectionalHint')}
          help={t('help:tooltips.testConfig.rfc2544.advanced.bidirectionalHelp')}
        />
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

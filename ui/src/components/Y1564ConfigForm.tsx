/**
 * @fileoverview Stem - Y.1564 Service Activation Test Configuration
 * @description Advanced configuration form for ITU-T Y.1564 / MEF Service Activation Testing.
 *              Allows users to configure service parameters including CIR, EIR, CBS, EBS,
 *              frame sizes, test duration, and SLA thresholds.
 *
 * Forms-stack pilot (#325): this form is the first to migrate to
 * react-hook-form + valibot. The schema lives at `src/schemas/configs.ts`
 * and is plumbed in via the `useConfigForm` helper. Field-level errors
 * render inline below each input; the cross-field rule (FDV ≤ FD) is
 * shown at the form footer. The 6 remaining ConfigForms follow the
 * same pattern — see issue #325 for the sweep.
 */

import { AlertTriangle } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { EXTENDED_FRAME_SIZE_OPTIONS } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { hasGroupTests } from '../lib/testGroups';
import { Y1564ConfigSchema } from '../schemas/configs';
import { Alert } from '../shared/Alert';
import { Checkbox } from '../shared/Checkbox';
import { FormField } from '../shared/FormField';
import { FormSection } from '../shared/FormSection';
import { HelpIcon } from '../shared/HelpIcon';
import { Input } from '../shared/Input';
import { TestSummary } from './TestSummary';

/** Y.1564 service configuration parameters */
export interface Y1564Config {
  /** Committed Information Rate in Mbps */
  cir: number;
  /** Excess Information Rate in Mbps */
  eir: number;
  /** Committed Burst Size in KB */
  cbs: number;
  /** Excess Burst Size in KB */
  ebs: number;
  /** Frame sizes to test */
  frameSizes: number[];
  /** Configuration test step duration in seconds */
  configStepDuration: number;
  /** Performance test duration in seconds */
  perfTestDuration: number;
  /** Frame Loss Ratio threshold (percentage) */
  flrThreshold: number;
  /** Frame Delay threshold (ms) */
  fdThreshold: number;
  /** Frame Delay Variation threshold (ms) */
  fdvThreshold: number;
}

/** Default Y.1564 configuration */
export const defaultY1564Config: Y1564Config = {
  cir: 100,
  eir: 0,
  cbs: 12,
  ebs: 0,
  frameSizes: [128, 256, 512, 1024, 1280, 1518],
  configStepDuration: 15,
  perfTestDuration: 900,
  flrThreshold: 0.01,
  fdThreshold: 10,
  fdvThreshold: 5,
};

interface Y1564ConfigFormProps {
  config: Y1564Config;
  setConfig: (config: Y1564Config) => void;
  selectedTests: string[];
  /** Overrides the root test id: the drawer and the page can both render this form at once. */
  testId?: string;
}

export function Y1564ConfigForm({
  config,
  setConfig,
  selectedTests,
  testId = 'y1564-config-form',
}: Y1564ConfigFormProps): ReactElement | null {
  const hasY1564Tests = hasGroupTests('y1564', selectedTests);

  const form = useConfigForm<Y1564Config>({
    schema: Y1564ConfigSchema,
    config,
    setConfig,
  });

  const {
    register,
    watch,
    setValue,
    formState: { errors },
  } = form;

  if (!hasY1564Tests) {
    return null;
  }

  // Watched values for derived displays (frame-size checkboxes,
  // summary panel). react-hook-form's watch keeps these in sync with
  // the form's internal state.
  const frameSizes = watch('frameSizes') ?? [];
  const cir = watch('cir') ?? 0;
  const eir = watch('eir') ?? 0;
  const flrThreshold = watch('flrThreshold') ?? 0;
  const fdThreshold = watch('fdThreshold') ?? 0;
  const fdvThreshold = watch('fdvThreshold') ?? 0;
  const configStepDuration = watch('configStepDuration') ?? 0;
  const perfTestDuration = watch('perfTestDuration') ?? 0;

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

  const isConfigTest = selectedTests.some((t) => t.includes('config'));
  const isPerfTest = selectedTests.some((t) => t.includes('perf'));
  const isFullTest = selectedTests.some((t) => t.includes('full'));

  // Cross-field error (fdv > fd). valibot's v.check() surfaces under
  // formState.errors.root.<unique-key>; we render the first one found.
  const { t } = useTranslation(['settings', 'help', 'common']);
  const rootErrors = errors.root;
  const crossFieldError = rootErrors
    ? Object.values(rootErrors).find(
        (e): e is { message: string } =>
          typeof e === 'object' && e !== null && 'message' in e && typeof e.message === 'string',
      )
    : undefined;

  return (
    <div data-testid={testId}>
      <div className="stack-lg">
        <FormSection title={t('testConfig.y1564.bandwidth.title')}>
          {/* CIR */}
          <FormField
            label={t('testConfig.y1564.bandwidth.cir')}
            hint={t('help:tooltips.testConfig.y1564.bandwidth.cirHelp')}
            error={errors.cir?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={1}
                {...register('cir', { valueAsNumber: true })}
              />
            )}
          </FormField>

          {/* EIR */}
          <FormField
            label={t('testConfig.y1564.bandwidth.eir')}
            hint={t('help:tooltips.testConfig.y1564.bandwidth.eirHelp')}
            error={errors.eir?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={1}
                {...register('eir', { valueAsNumber: true })}
              />
            )}
          </FormField>

          {/* CBS */}
          <FormField
            label={t('testConfig.y1564.bandwidth.cbs')}
            hint={t('help:tooltips.testConfig.y1564.bandwidth.cbsHelp')}
            error={errors.cbs?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={1}
                {...register('cbs', { valueAsNumber: true })}
              />
            )}
          </FormField>

          {/* EBS */}
          <FormField
            label={t('testConfig.y1564.bandwidth.ebs')}
            hint={t('help:tooltips.testConfig.y1564.bandwidth.ebsHelp')}
            error={errors.ebs?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={1}
                {...register('ebs', { valueAsNumber: true })}
              />
            )}
          </FormField>
        </FormSection>

        <FormSection title={t('testConfig.y1564.sla.title')}>
          {/* Frame Loss Ratio */}
          <FormField
            label={t('testConfig.y1564.sla.flr')}
            hint={t('help:tooltips.testConfig.y1564.sla.flrHelp')}
            error={errors.flrThreshold?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={0.001}
                {...register('flrThreshold', { valueAsNumber: true })}
              />
            )}
          </FormField>

          {/* Frame Delay */}
          <FormField
            label={t('testConfig.y1564.sla.fd')}
            hint={t('help:tooltips.testConfig.y1564.sla.fdHelp')}
            error={errors.fdThreshold?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={1}
                {...register('fdThreshold', { valueAsNumber: true })}
              />
            )}
          </FormField>

          {/* Frame Delay Variation */}
          <FormField
            label={t('testConfig.y1564.sla.fdv')}
            hint={t('help:tooltips.testConfig.y1564.sla.fdvHelp')}
            error={errors.fdvThreshold?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={1}
                {...register('fdvThreshold', { valueAsNumber: true })}
              />
            )}
          </FormField>
        </FormSection>

        <FormSection
          title={t('testConfig.common.frameSizesTitle')}
          help={
            <HelpIcon
              data-testid="help-icon"
              label={t('common:help.label', { topic: t('testConfig.common.frameSizesTitle') })}
              text={t('help:tooltips.testConfig.y1564.frameSizes.help')}
            />
          }
        >
          <div className="grid grid-cols-2 gap-compact">
            {EXTENDED_FRAME_SIZE_OPTIONS.map((option) => (
              <Checkbox
                key={option.value}
                checked={frameSizes.includes(option.value)}
                onChange={() => toggleFrameSize(option.value)}
                label={t(`testConfig.common.${option.qualifier}`, { size: option.value })}
              />
            ))}
          </div>
          {errors.frameSizes?.message ? (
            <Alert tone="error" size="sm">
              {errors.frameSizes.message}
            </Alert>
          ) : null}
        </FormSection>

        <FormSection title={t('testConfig.y1564.duration.title')}>
          {isConfigTest || isFullTest ? (
            <>
              <FormField
                label={t('testConfig.y1564.duration.configStep')}
                hint={t('help:tooltips.testConfig.y1564.duration.configStepHelp')}
                error={errors.configStepDuration?.message}
              >
                {(control) => (
                  <Input
                    {...control}
                    type="number"
                    step={1}
                    {...register('configStepDuration', { valueAsNumber: true })}
                  />
                )}
              </FormField>
              <div className="text-xs text-text-muted mt-tight">
                {t('testConfig.y1564.duration.configTotal', {
                  seconds: configStepDuration * 4 * frameSizes.length,
                })}
              </div>
            </>
          ) : null}

          {isPerfTest || isFullTest ? (
            <>
              <FormField
                label={t('testConfig.y1564.duration.perf')}
                hint={t('help:tooltips.testConfig.y1564.duration.perfHelp')}
                error={errors.perfTestDuration?.message}
              >
                {(control) => (
                  <Input
                    {...control}
                    type="number"
                    step={60}
                    {...register('perfTestDuration', { valueAsNumber: true })}
                  />
                )}
              </FormField>
              <div className="text-xs text-text-muted mt-tight">
                {t('testConfig.y1564.duration.perfMinutes', {
                  minutes: Math.floor(perfTestDuration / 60),
                })}
              </div>
            </>
          ) : null}
        </FormSection>

        {/* Cross-field error footer */}
        {crossFieldError && (
          <div className="pad-xs rounded-lg bg-status-error/10 text-status-error-strong text-sm flex items-center gap-compact">
            <AlertTriangle className="w-4 h-4" />
            {crossFieldError.message}
          </div>
        )}

        <TestSummary>
          <div>
            {eir > 0
              ? t('testConfig.y1564.summary.serviceWithEir', { cir, eir })
              : t('testConfig.y1564.summary.service', { cir })}
          </div>
          <div>{t('testConfig.common.frameSizesSummary', { sizes: frameSizes.join(', ') })}</div>
          <div>
            {t('testConfig.y1564.summary.sla', {
              flr: flrThreshold,
              fd: fdThreshold,
              fdv: fdvThreshold,
            })}
          </div>
        </TestSummary>
      </div>
    </div>
  );
}

export default Y1564ConfigForm;

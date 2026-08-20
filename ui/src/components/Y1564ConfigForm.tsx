/**
 * @fileoverview The Stem - Y.1564 Service Activation Test Configuration
 * @description Advanced configuration form for ITU-T Y.1564 / MEF Service Activation Testing.
 *              Allows users to configure service parameters including CIR, EIR, CBS, EBS,
 *              frame sizes, test duration, and VLAN settings.
 *
 * Forms-stack pilot (#325): this form is the first to migrate to
 * react-hook-form + valibot. The schema lives at `src/schemas/configs.ts`
 * and is plumbed in via the `useConfigForm` helper. Field-level errors
 * render inline below each input; the cross-field rule (FDV ≤ FD) is
 * shown at the form footer. The 6 remaining ConfigForms follow the
 * same pattern — see issue #325 for the sweep.
 */

import { AlertTriangle, Settings2 } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { FRAME_SIZE_OPTIONS } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { Y1564ConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { FieldError } from './FieldError';
import { FormSection } from './FormSection';
import { HelpIcon } from './HelpIcon';
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
  /** VLAN ID (0 = untagged) */
  vlanId: number;
  /** Priority Code Point (0-7) */
  pcp: number;
  /** Color-aware mode enabled */
  colorAware: boolean;
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
  frameSizes: [64, 128, 256, 512, 1024, 1280, 1518],
  configStepDuration: 15,
  perfTestDuration: 900,
  vlanId: 0,
  pcp: 0,
  colorAware: false,
  flrThreshold: 0.01,
  fdThreshold: 10,
  fdvThreshold: 5,
};

interface Y1564ConfigFormProps {
  config: Y1564Config;
  setConfig: (config: Y1564Config) => void;
  selectedTests: string[];
}

export function Y1564ConfigForm({
  config,
  setConfig,
  selectedTests,
}: Y1564ConfigFormProps): ReactElement | null {
  const hasY1564Tests = selectedTests.some((t) => t.startsWith('y1564') || t.startsWith('mef'));

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
  // summary panel, VLAN PCP conditional render). react-hook-form's
  // watch keeps these in sync with the form's internal state.
  const frameSizes = watch('frameSizes') ?? [];
  const vlanId = watch('vlanId') ?? 0;
  const cir = watch('cir') ?? 0;
  const eir = watch('eir') ?? 0;
  const flrThreshold = watch('flrThreshold') ?? 0;
  const fdThreshold = watch('fdThreshold') ?? 0;
  const fdvThreshold = watch('fdvThreshold') ?? 0;
  const pcp = watch('pcp') ?? 0;
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
  const { t } = useTranslation('settings');
  const rootErrors = errors.root;
  const crossFieldError = rootErrors
    ? Object.values(rootErrors).find(
        (e): e is { message: string } =>
          typeof e === 'object' && e !== null && 'message' in e && typeof e.message === 'string',
      )
    : undefined;

  return (
    <CollapsibleSection
      testId="y1564-config-form"
      title={
        <div className="flex items-center gap-compact">
          <Settings2 className="w-4 h-4" />
          <span>{t('testConfig.y1564.title')}</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="stack-lg">
        <FormSection title={t('testConfig.y1564.bandwidth.title')}>
          {/* CIR */}
          <div>
            <label htmlFor="y1564-cir" className="flex items-center gap-tight label">
              {t('testConfig.y1564.bandwidth.cir')}
              <HelpIcon tooltip={t('testConfig.y1564.bandwidth.cirHelp')} />
            </label>
            <input
              id="y1564-cir"
              type="number"
              step={1}
              {...register('cir', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.cir?.message} />
          </div>

          {/* EIR */}
          <div>
            <label htmlFor="y1564-eir" className="flex items-center gap-tight label">
              {t('testConfig.y1564.bandwidth.eir')}
              <HelpIcon tooltip={t('testConfig.y1564.bandwidth.eirHelp')} />
            </label>
            <input
              id="y1564-eir"
              type="number"
              step={1}
              {...register('eir', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.eir?.message} />
          </div>

          {/* CBS */}
          <div>
            <label htmlFor="y1564-cbs" className="flex items-center gap-tight label">
              {t('testConfig.y1564.bandwidth.cbs')}
              <HelpIcon tooltip={t('testConfig.y1564.bandwidth.cbsHelp')} />
            </label>
            <input
              id="y1564-cbs"
              type="number"
              step={1}
              {...register('cbs', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.cbs?.message} />
          </div>

          {/* EBS */}
          <div>
            <label htmlFor="y1564-ebs" className="flex items-center gap-tight label">
              {t('testConfig.y1564.bandwidth.ebs')}
              <HelpIcon tooltip={t('testConfig.y1564.bandwidth.ebsHelp')} />
            </label>
            <input
              id="y1564-ebs"
              type="number"
              step={1}
              {...register('ebs', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.ebs?.message} />
          </div>
        </FormSection>

        <FormSection title={t('testConfig.y1564.sla.title')}>
          {/* Frame Loss Ratio */}
          <div>
            <label htmlFor="y1564-flr" className="flex items-center gap-tight label">
              {t('testConfig.y1564.sla.flr')}
              <HelpIcon tooltip={t('testConfig.y1564.sla.flrHelp')} />
            </label>
            <input
              id="y1564-flr"
              type="number"
              step={0.001}
              {...register('flrThreshold', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.flrThreshold?.message} />
          </div>

          {/* Frame Delay */}
          <div>
            <label htmlFor="y1564-fd" className="flex items-center gap-tight label">
              {t('testConfig.y1564.sla.fd')}
              <HelpIcon tooltip={t('testConfig.y1564.sla.fdHelp')} />
            </label>
            <input
              id="y1564-fd"
              type="number"
              step={1}
              {...register('fdThreshold', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.fdThreshold?.message} />
          </div>

          {/* Frame Delay Variation */}
          <div>
            <label htmlFor="y1564-fdv" className="flex items-center gap-tight label">
              {t('testConfig.y1564.sla.fdv')}
              <HelpIcon tooltip={t('testConfig.y1564.sla.fdvHelp')} />
            </label>
            <input
              id="y1564-fdv"
              type="number"
              step={1}
              {...register('fdvThreshold', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.fdvThreshold?.message} />
          </div>
        </FormSection>

        <FormSection
          title={t('testConfig.common.frameSizesTitle')}
          help={<HelpIcon tooltip={t('testConfig.y1564.frameSizes.help')} />}
        >
          <div className="grid grid-cols-2 gap-compact">
            {FRAME_SIZE_OPTIONS.map((option) => (
              <label
                key={option.value}
                title={t('testConfig.y1564.frameSizes.includeTitle', { size: option.value })}
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

        <FormSection title={t('testConfig.y1564.duration.title')}>
          {isConfigTest || isFullTest ? (
            <div>
              <label htmlFor="y1564-config-duration" className="flex items-center gap-tight label">
                {t('testConfig.y1564.duration.configStep')}
                <HelpIcon tooltip={t('testConfig.y1564.duration.configStepHelp')} />
              </label>
              <input
                id="y1564-config-duration"
                type="number"
                step={1}
                {...register('configStepDuration', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.configStepDuration?.message} />
              <div className="text-xs text-text-muted mt-tight">
                {t('testConfig.y1564.duration.configTotal', {
                  seconds: configStepDuration * 4 * frameSizes.length,
                })}
              </div>
            </div>
          ) : null}

          {isPerfTest || isFullTest ? (
            <div>
              <label htmlFor="y1564-perf-duration" className="flex items-center gap-tight label">
                {t('testConfig.y1564.duration.perf')}
                <HelpIcon tooltip={t('testConfig.y1564.duration.perfHelp')} />
              </label>
              <input
                id="y1564-perf-duration"
                type="number"
                step={60}
                {...register('perfTestDuration', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.perfTestDuration?.message} />
              <div className="text-xs text-text-muted mt-tight">
                {t('testConfig.y1564.duration.perfMinutes', {
                  minutes: Math.floor(perfTestDuration / 60),
                })}
              </div>
            </div>
          ) : null}
        </FormSection>

        <FormSection title={t('testConfig.y1564.vlan.title')}>
          <div>
            <label htmlFor="y1564-vlan" className="flex items-center gap-tight label">
              {t('testConfig.y1564.vlan.id')}
              <HelpIcon tooltip={t('testConfig.y1564.vlan.idHelp')} />
            </label>
            <input
              id="y1564-vlan"
              type="number"
              step={1}
              {...register('vlanId', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.vlanId?.message} />
            {vlanId === 0 && (
              <div className="text-xs text-text-muted mt-tight">
                {t('testConfig.y1564.vlan.untagged')}
              </div>
            )}
          </div>

          {vlanId > 0 && (
            <div>
              <label htmlFor="y1564-pcp" className="flex items-center gap-tight label">
                {t('testConfig.y1564.vlan.pcp')}
                <HelpIcon tooltip={t('testConfig.y1564.vlan.pcpHelp')} />
              </label>
              <select
                id="y1564-pcp"
                {...register('pcp', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {[0, 1, 2, 3, 4, 5, 6, 7].map((priority) => (
                  <option key={priority} value={priority}>
                    {t(`testConfig.y1564.vlan.pcp${priority}` as never)}
                  </option>
                ))}
              </select>
              <FieldError message={errors.pcp?.message} />
            </div>
          )}

          {/* Color-Aware Mode */}
          <label
            title={t('testConfig.y1564.vlan.colorAwareTitle')}
            className="flex items-center gap-default pad-xs rounded-lg cursor-pointer hover:bg-surface-hover"
          >
            <input
              type="checkbox"
              {...register('colorAware')}
              aria-label={t('testConfig.y1564.vlan.colorAwareAria')}
              className="w-4 h-4 accent-brand-primary"
            />
            <div>
              <div className="font-medium text-sm flex items-center gap-tight">
                {t('testConfig.y1564.vlan.colorAware')}
                <HelpIcon tooltip={t('testConfig.y1564.vlan.colorAwareHelp')} />
              </div>
              <div className="text-xs text-text-muted">
                {t('testConfig.y1564.vlan.colorAwareHint')}
              </div>
            </div>
          </label>
        </FormSection>

        {/* Cross-field error footer */}
        {crossFieldError && (
          <div className="pad-xs rounded-lg bg-status-error/10 text-status-error text-sm flex items-center gap-compact">
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
          {vlanId > 0 && <div>{t('testConfig.y1564.summary.vlan', { vlan: vlanId, pcp })}</div>}
        </TestSummary>
      </div>
    </CollapsibleSection>
  );
}

export default Y1564ConfigForm;

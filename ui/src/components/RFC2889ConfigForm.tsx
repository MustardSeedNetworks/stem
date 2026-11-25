/**
 * @fileoverview Stem - RFC 2889 LAN Switch Test Configuration
 * @description Configuration form for RFC 2889 LAN Switch Benchmarking Tests.
 *              Migrated to react-hook-form + valibot per #325.
 */

import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import type { FrameSizeOption } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { hasGroupTests } from '../lib/testGroups';
import { RFC2889ConfigSchema } from '../schemas/configs';
import { FormField } from '../shared/FormField';
import { Input } from '../shared/Input';
import { NumberSelectField } from './ConfigFormFields';
import { FormSection } from './FormSection';
import { TestSummary } from './TestSummary';

/** RFC 2889 test configuration parameters */
export interface RFC2889Config {
  /** Frame size in bytes */
  frameSize: number;
  /** Test duration in seconds */
  duration: number;
  /** Warmup duration before measurement (seconds) */
  warmup: number;
  /** Number of MAC addresses for learning/caching tests */
  addressCount: number;
  /** Maximum acceptable frame loss (percentage) */
  acceptableLoss: number;
  /** Number of ports to test */
  portCount: number;
  /** Traffic pattern: 0=mesh, 1=pair, 2=broadcast */
  pattern: number;
}

/** Default RFC 2889 configuration */
export const defaultRFC2889Config: RFC2889Config = {
  frameSize: 64,
  duration: 60,
  warmup: 2,
  addressCount: 8192,
  acceptableLoss: 0.0,
  portCount: 2,
  pattern: 0,
};

/** Switch benchmarking stops at the standard MTU; no jumbo. */
const FRAME_SIZE_OPTIONS: FrameSizeOption[] = [
  { value: 64, qualifier: 'frameSizeMin' },
  { value: 128, qualifier: 'frameSize' },
  { value: 256, qualifier: 'frameSize' },
  { value: 512, qualifier: 'frameSize' },
  { value: 1024, qualifier: 'frameSize' },
  { value: 1280, qualifier: 'frameSize' },
  { value: 1518, qualifier: 'frameSizeMax' },
];

/** Wire values, with the key that names each for the operator. */
const PATTERN_OPTIONS: Array<{ value: number; key: string }> = [
  { value: 0, key: 'patternFullMesh' },
  { value: 1, key: 'patternPair' },
  { value: 2, key: 'patternBroadcast' },
];

interface RFC2889ConfigFormProps {
  config: RFC2889Config;
  setConfig: (config: RFC2889Config) => void;
  selectedTests: string[];
  /** Overrides the root test id: the drawer and the page can both render this form at once. */
  testId?: string;
}

export function RFC2889ConfigForm({
  config,
  setConfig,
  selectedTests,
  testId = 'rfc2889-config-form',
}: RFC2889ConfigFormProps): ReactElement | null {
  const hasRFC2889Tests = hasGroupTests('rfc2889', selectedTests);

  const form = useConfigForm<RFC2889Config>({
    schema: RFC2889ConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    control,
    watch,
    formState: { errors },
  } = form;
  const { t } = useTranslation(['settings', 'help']);

  if (!hasRFC2889Tests) {
    return null;
  }

  const frameSize = watch('frameSize') ?? 0;
  const duration = watch('duration') ?? 0;
  const warmup = watch('warmup') ?? 0;
  const portCount = watch('portCount') ?? 0;
  const pattern = watch('pattern') ?? 0;

  const hasForwarding = selectedTests.includes('rfc2889_forwarding');
  const hasCaching = selectedTests.includes('rfc2889_caching');
  const hasLearning = selectedTests.includes('rfc2889_learning');
  const hasBroadcast = selectedTests.includes('rfc2889_broadcast');
  const hasCongestion = selectedTests.includes('rfc2889_congestion');

  return (
    <div data-testid={testId}>
      <div className="stack-lg">
        <FormSection title={t('testConfig.rfc2889.params.title')}>
          <div className="grid grid-cols-2 gap-default">
            <FormField
              label={t('testConfig.rfc2889.params.duration')}
              hint={t('help:tooltips.testConfig.rfc2889.params.durationHelp')}
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
              label={t('testConfig.rfc2889.params.warmup')}
              hint={t('help:tooltips.testConfig.rfc2889.params.warmupHelp')}
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

        <NumberSelectField
          control={control}
          name="frameSize"
          label={t('testConfig.rfc2889.params.frameSize')}
          hint={t('help:tooltips.testConfig.rfc2889.params.frameSizeHelp')}
          error={errors.frameSize?.message}
          options={FRAME_SIZE_OPTIONS.map((option) => ({
            value: option.value,
            label: t(`testConfig.common.${option.qualifier}`, { size: option.value }),
          }))}
        />

        <FormSection title={t('testConfig.rfc2889.switch.title')}>
          <div className="grid grid-cols-2 gap-default">
            <FormField
              label={t('testConfig.rfc2889.switch.portCount')}
              hint={t('help:tooltips.testConfig.rfc2889.switch.portCountHelp')}
              error={errors.portCount?.message}
            >
              {(control) => (
                <Input
                  {...control}
                  type="number"
                  step={1}
                  {...register('portCount', { valueAsNumber: true })}
                />
              )}
            </FormField>

            <NumberSelectField
              control={control}
              name="pattern"
              label={t('testConfig.rfc2889.switch.pattern')}
              hint={t('help:tooltips.testConfig.rfc2889.switch.patternHelp')}
              error={errors.pattern?.message}
              options={PATTERN_OPTIONS.map((option) => ({
                value: option.value,
                label: t(`testConfig.rfc2889.switch.${option.key}` as never),
              }))}
            />
          </div>
        </FormSection>

        {hasCaching || hasLearning ? (
          <FormField
            label={t('testConfig.rfc2889.switch.addressCount')}
            hint={t('help:tooltips.testConfig.rfc2889.switch.addressCountHelp')}
            error={errors.addressCount?.message}
          >
            {(control) => (
              <Input
                {...control}
                type="number"
                step={1}
                {...register('addressCount', { valueAsNumber: true })}
              />
            )}
          </FormField>
        ) : null}

        <FormField
          label={t('testConfig.rfc2889.switch.loss')}
          hint={t('help:tooltips.testConfig.rfc2889.switch.lossHelp')}
          error={errors.acceptableLoss?.message}
        >
          {(control) => (
            <Input
              {...control}
              type="number"
              step={0.001}
              {...register('acceptableLoss', { valueAsNumber: true })}
            />
          )}
        </FormField>

        <TestSummary>
          <div>
            {t('testConfig.common.selectedTests')}:{' '}
            {[
              hasForwarding && t('testConfig.rfc2889.tests.forwarding'),
              hasCaching && t('testConfig.rfc2889.tests.caching'),
              hasLearning && t('testConfig.rfc2889.tests.learning'),
              hasBroadcast && t('testConfig.rfc2889.tests.broadcast'),
              hasCongestion && t('testConfig.rfc2889.tests.congestion'),
            ]
              .filter(Boolean)
              .join(', ')}
          </div>
          <div>{t('testConfig.rfc2889.summary.frameSize', { size: frameSize })}</div>
          <div>
            {t('testConfig.rfc2889.summary.ports', {
              ports: portCount,
              pattern: t(
                `testConfig.rfc2889.switch.${
                  PATTERN_OPTIONS.find((option) => option.value === pattern)?.key ??
                  'patternFullMesh'
                }` as never,
              ),
            })}
          </div>
          <div>{t('testConfig.rfc2889.summary.duration', { duration, warmup })}</div>
        </TestSummary>
      </div>
    </div>
  );
}

export default RFC2889ConfigForm;

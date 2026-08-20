/**
 * @fileoverview The Stem - RFC 2889 LAN Switch Test Configuration
 * @description Configuration form for RFC 2889 LAN Switch Benchmarking Tests.
 *              Migrated to react-hook-form + valibot per #325.
 */

import { Network } from 'lucide-react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import type { FrameSizeOption } from '../forms/frameSizes';
import { useConfigForm } from '../forms/useConfigForm';
import { RFC2889ConfigSchema } from '../schemas/configs';
import { CollapsibleSection } from './CollapsibleSection';
import { FieldError } from './FieldError';
import { FormSection } from './FormSection';
import { HelpIcon } from './HelpIcon';
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
}

export function RFC2889ConfigForm({
  config,
  setConfig,
  selectedTests,
}: RFC2889ConfigFormProps): ReactElement | null {
  const hasRFC2889Tests = selectedTests.some((t) => t.startsWith('rfc2889'));

  const form = useConfigForm<RFC2889Config>({
    schema: RFC2889ConfigSchema,
    config,
    setConfig,
  });
  const {
    register,
    watch,
    formState: { errors },
  } = form;
  const { t } = useTranslation('settings');

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
    <CollapsibleSection
      testId="rfc2889-config-form"
      title={
        <div className="flex items-center gap-compact">
          <Network className="w-4 h-4" />
          <span>RFC 2889 Configuration</span>
        </div>
      }
      defaultOpen={true}
    >
      <div className="stack-lg">
        <FormSection title={t('testConfig.rfc2889.params.title')}>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="rfc2889-duration" className="flex items-center gap-tight label">
                {t('testConfig.rfc2889.params.duration')}
                <HelpIcon tooltip={t('testConfig.rfc2889.params.durationHelp')} />
              </label>
              <input
                id="rfc2889-duration"
                type="number"
                step={1}
                {...register('duration', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.duration?.message} />
            </div>

            <div>
              <label htmlFor="rfc2889-warmup" className="flex items-center gap-tight label">
                {t('testConfig.rfc2889.params.warmup')}
                <HelpIcon tooltip={t('testConfig.rfc2889.params.warmupHelp')} />
              </label>
              <input
                id="rfc2889-warmup"
                type="number"
                step={1}
                {...register('warmup', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.warmup?.message} />
            </div>
          </div>
        </FormSection>

        <div>
          <label htmlFor="rfc2889-framesize" className="flex items-center gap-tight label">
            {t('testConfig.rfc2889.params.frameSize')}
            <HelpIcon tooltip={t('testConfig.rfc2889.params.frameSizeHelp')} />
          </label>
          <select
            id="rfc2889-framesize"
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

        <FormSection title={t('testConfig.rfc2889.switch.title')}>
          <div className="grid grid-cols-2 gap-default">
            <div>
              <label htmlFor="rfc2889-portcount" className="flex items-center gap-tight label">
                {t('testConfig.rfc2889.switch.portCount')}
                <HelpIcon tooltip={t('testConfig.rfc2889.switch.portCountHelp')} />
              </label>
              <input
                id="rfc2889-portcount"
                type="number"
                step={1}
                {...register('portCount', { valueAsNumber: true })}
                className="mt-tight w-full"
              />
              <FieldError message={errors.portCount?.message} />
            </div>

            <div>
              <label htmlFor="rfc2889-pattern" className="flex items-center gap-tight label">
                {t('testConfig.rfc2889.switch.pattern')}
                <HelpIcon tooltip={t('testConfig.rfc2889.switch.patternHelp')} />
              </label>
              <select
                id="rfc2889-pattern"
                {...register('pattern', { valueAsNumber: true })}
                className="mt-tight w-full"
              >
                {PATTERN_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {t(`testConfig.rfc2889.switch.${option.key}` as never)}
                  </option>
                ))}
              </select>
              <FieldError message={errors.pattern?.message} />
            </div>
          </div>
        </FormSection>

        {hasCaching || hasLearning ? (
          <div>
            <label htmlFor="rfc2889-addresscount" className="flex items-center gap-tight label">
              {t('testConfig.rfc2889.switch.addressCount')}
              <HelpIcon tooltip={t('testConfig.rfc2889.switch.addressCountHelp')} />
            </label>
            <input
              id="rfc2889-addresscount"
              type="number"
              step={1}
              {...register('addressCount', { valueAsNumber: true })}
              className="mt-tight w-full"
            />
            <FieldError message={errors.addressCount?.message} />
          </div>
        ) : null}

        <div>
          <label htmlFor="rfc2889-loss" className="flex items-center gap-tight label">
            {t('testConfig.rfc2889.switch.loss')}
            <HelpIcon tooltip={t('testConfig.rfc2889.switch.lossHelp')} />
          </label>
          <input
            id="rfc2889-loss"
            type="number"
            step={0.001}
            {...register('acceptableLoss', { valueAsNumber: true })}
            className="mt-tight w-full"
          />
          <FieldError message={errors.acceptableLoss?.message} />
        </div>

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
    </CollapsibleSection>
  );
}

export default RFC2889ConfigForm;
